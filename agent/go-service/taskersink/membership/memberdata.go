package membership

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

var appVersion string

// SetVersion sets the application version for debug-mode detection.
func SetVersion(v string) {
	appVersion = v
}

// isDebugVersion returns true when the version is below 1.0.0 (dev builds, pre-release).
func isDebugVersion() bool {
	if appVersion == "" || appVersion == "dev" {
		return true
	}
	v := strings.TrimPrefix(appVersion, "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) == 0 {
		return true
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return true
	}
	return major < 1
}

// MemberRecord represents a single member entry from the V6 JSON data.
type MemberRecord struct {
	UserID           string      `json:"user_id"`
	CPUHash          string      `json:"cpu_hash"`
	UUIDHash         string      `json:"uuid_hash"`
	BIOSHash         string      `json:"bios_hash"`
	BoardHash        string      `json:"board_hash"`
	DiskHash         string      `json:"disk_hash"`
	GUIDHash         string      `json:"guid_hash"`
	Tier             string      `json:"tier"`
	AccountValue     json.Number `json:"account_value"`
	RegistrationDate string      `json:"registration_date"`
}

// MembershipStatus represents the calculated current membership state.
type MembershipStatus struct {
	MembershipType  string
	UserLevel       int
	RemainingValue  float64
	VirtualExpiry   string
	IsMember        bool
	UserID          string
	UnsupportedTier bool
	DeviceCode      DeviceCodeV6
}

var (
	cachedData       []MemberRecord
	cachedDataMu     sync.RWMutex
	cachedDataTime   time.Time
	cachedStatus     *MembershipStatus
	cachedStatusMu   sync.RWMutex
	cachedDeviceCode DeviceCodeV6
)

const (
	cacheExpiry         = 1 * time.Hour
	matchThreshold      = 80
	httpTimeout         = 15 * time.Second
	memberDataCachePath = "config/memberdata-v6-cache.json"
)

// GetMembershipStatus returns the current membership status, using cache if available.
func GetMembershipStatus() *MembershipStatus {
	cachedStatusMu.RLock()
	if cachedStatus != nil && time.Since(cachedDataTime) < cacheExpiry {
		status := cachedStatus
		cachedStatusMu.RUnlock()
		return status
	}
	cachedStatusMu.RUnlock()

	return checkMembership()
}

// checkMembership performs the full membership check flow.
func checkMembership() *MembershipStatus {
	defaultStatus := &MembershipStatus{
		MembershipType: "普通用户",
		UserLevel:      0,
		IsMember:       false,
	}

	// Debug versions (below 1.0.0) bypass membership verification
	if isDebugVersion() {
		log.Info().Str("version", appVersion).Msg("Debug version detected, bypassing membership verification")
		return &MembershipStatus{
			MembershipType: "金Doro会员",
			UserLevel:      3,
			IsMember:       true,
			VirtualExpiry:  "99991231",
		}
	}

	// Generate device code
	deviceCode := GenerateDeviceCodeV6()
	cachedDeviceCode = deviceCode
	defaultStatus.DeviceCode = deviceCode

	log.Info().
		Str("cpu_hash", shortHash(deviceCode.CPUHash)).
		Str("uuid_hash", shortHash(deviceCode.UUIDHash)).
		Msg("Generated V6 device code")

	// Fetch member data
	records, err := fetchMemberData()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch member data, treating as non-member")
		return defaultStatus
	}

	// Match device code
	record, score := matchDeviceCode(deviceCode, records)
	if record == nil {
		log.Info().Msg("No matching member record found")
		return defaultStatus
	}

	log.Info().
		Str("user_id", record.UserID).
		Int("score", score).
		Str("tier", record.Tier).
		Msg("Matched member record")

	// Check for unsupported tiers
	unsupported := false
	if unsupportedTiers[record.Tier] {
		log.Warn().
			Str("tier", record.Tier).
			Msg("检测到不再支持的会员等级，该等级在 MDA 中不享受会员功能。请升级至金Doro会员或以上。")
		unsupported = true
	}

	// Calculate current membership status
	status := calculateStatus(record)
	status.UnsupportedTier = unsupported
	status.DeviceCode = deviceCode

	// Cache the result
	cachedStatusMu.Lock()
	cachedStatus = status
	cachedDataTime = time.Now()
	cachedStatusMu.Unlock()

	return status
}

// fetchMemberData fetches V6 member data from the API.
func fetchMemberData() ([]MemberRecord, error) {
	cachedDataMu.RLock()
	if cachedData != nil && time.Since(cachedDataTime) < cacheExpiry {
		data := cachedData
		cachedDataMu.RUnlock()
		return data, nil
	}
	cachedDataMu.RUnlock()

	client := &http.Client{Timeout: httpTimeout}
	records, err := fetchFromSource(client, MemberDataURL)
	if err != nil {
		cachedRecords, cacheErr := loadCachedMemberData()
		if cacheErr == nil {
			cacheData(cachedRecords)
			log.Warn().
				Err(err).
				Int("count", len(cachedRecords)).
				Msg("Failed to fetch member data, using cached V6 member data")
			return cachedRecords, nil
		}
		return nil, fmt.Errorf("failed to fetch member data from %s: %w", MemberDataURL, err)
	}

	cacheData(records)
	if err := saveCachedMemberData(records); err != nil {
		log.Warn().Err(err).Msg("Failed to save V6 member data cache")
	}

	log.Info().Str("url", MemberDataURL).Int("count", len(records)).Msg("Fetched V6 member data")
	return records, nil
}

func cacheData(records []MemberRecord) {
	cachedDataMu.Lock()
	cachedData = records
	cachedDataTime = time.Now()
	cachedDataMu.Unlock()
}

func loadCachedMemberData() ([]MemberRecord, error) {
	body, err := os.ReadFile(memberDataCachePath)
	if err != nil {
		return nil, err
	}

	var records []MemberRecord
	if err := json.Unmarshal(body, &records); err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("cached member data is empty")
	}
	return records, nil
}

func saveCachedMemberData(records []MemberRecord) error {
	body, err := json.Marshal(records)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(memberDataCachePath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(memberDataCachePath, body, 0o644)
}

// fetchFromSource fetches and parses member data from a single URL.
func fetchFromSource(client *http.Client, url string) ([]MemberRecord, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "MDA/"+appVersion)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var records []MemberRecord
	if err := json.Unmarshal(body, &records); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from %s: %w", url, err)
	}

	return records, nil
}

// matchDeviceCode finds the best matching member record using weighted scoring.
func matchDeviceCode(current DeviceCodeV6, records []MemberRecord) (*MemberRecord, int) {
	var bestRecord *MemberRecord
	bestScore := 0

	for i := range records {
		saved := DeviceCodeV6{
			CPUHash:   records[i].CPUHash,
			UUIDHash:  records[i].UUIDHash,
			BIOSHash:  records[i].BIOSHash,
			BoardHash: records[i].BoardHash,
			DiskHash:  records[i].DiskHash,
			GUIDHash:  records[i].GUIDHash,
		}
		score := MatchDeviceCodeV6(current, saved)
		if score > bestScore {
			bestScore = score
			bestRecord = &records[i]
		}
	}

	if bestScore >= matchThreshold {
		return bestRecord, bestScore
	}
	return nil, bestScore
}

// calculateStatus computes the current membership status using the prepaid credit burn model.
func calculateStatus(record *MemberRecord) *MembershipStatus {
	status := &MembershipStatus{
		MembershipType: "普通用户",
		UserLevel:      0,
		IsMember:       false,
		UserID:         record.UserID,
	}

	tier := record.Tier
	level, ok := membershipLevels[tier]
	if !ok {
		log.Warn().Str("tier", tier).Msg("Unknown membership tier")
		return status
	}

	accountValue, err := record.AccountValue.Float64()
	if err != nil {
		log.Warn().Stringer("value", record.AccountValue).Err(err).Msg("Failed to parse account value")
		return status
	}

	cost, ok := monthlyCost[tier]
	if !ok || cost <= 0 {
		// Free tier or unknown - if they have value, treat as active
		if accountValue > 0 {
			status.MembershipType = tier
			status.UserLevel = level
			status.RemainingValue = accountValue
			status.IsMember = level >= minMemberLevel
			status.VirtualExpiry = "99991231"
		}
		return status
	}

	dailyCost := cost / 30.0

	// Calculate days passed since registration
	daysPassed := daysSince(record.RegistrationDate)
	if daysPassed < 0 {
		daysPassed = 0
	}

	consumedValue := float64(daysPassed) * dailyCost
	remainingValue := accountValue - consumedValue

	if remainingValue < 0.001 {
		// Membership expired
		status.RemainingValue = 0
		return status
	}

	// Calculate virtual expiry date
	daysLeft := int(math.Floor(remainingValue / dailyCost))
	expiry := time.Now().AddDate(0, 0, daysLeft).Format("20060102")

	status.MembershipType = tier
	status.UserLevel = level
	status.RemainingValue = remainingValue
	status.VirtualExpiry = expiry
	status.IsMember = level >= minMemberLevel

	return status
}

// daysSince calculates the number of days from the given date string (YYYYMMDD) to now.
func daysSince(dateStr string) int {
	dateStr = strings.TrimSpace(dateStr)
	if len(dateStr) < 8 {
		return 0
	}
	t, err := time.ParseInLocation("20060102", dateStr[:8], time.Local)
	if err != nil {
		log.Debug().Str("date", dateStr).Err(err).Msg("Failed to parse date")
		return 0
	}
	return int(time.Since(t).Hours() / 24)
}

func shortHash(s string) string {
	if len(s) > 8 {
		return s[:8] + "..."
	}
	if s == "" {
		return "(empty)"
	}
	return s
}
