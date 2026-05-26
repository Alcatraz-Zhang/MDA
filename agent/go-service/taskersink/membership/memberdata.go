package membership

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
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

type MemberStatusResponse struct {
	Matched       bool   `json:"matched"`
	Score         int    `json:"score"`
	IsMember      bool   `json:"is_member"`
	UserID        string `json:"user_id"`
	Tier          string `json:"tier"`
	PlanCode      string `json:"plan_code"`
	PlanName      string `json:"plan_name"`
	StartsOn      string `json:"starts_on"`
	ExpiresOn     string `json:"expires_on"`
	RemainingDays int    `json:"remaining_days"`
}

// MembershipStatus represents the current membership state.
type MembershipStatus struct {
	Tier          string
	PlanCode      string
	PlanName      string
	StartsOn      string
	ExpiresOn     string
	RemainingDays int
	IsMember      bool
	UserID        string
	DeviceCode    DeviceCodeV6
}

var (
	cachedStatus     *MembershipStatus
	cachedStatusMu   sync.RWMutex
	cachedStatusTime time.Time
	cachedDeviceCode DeviceCodeV6
)

const (
	cacheExpiry      = 1 * time.Hour
	httpTimeout      = 15 * time.Second
	maxFetchAttempts = 3
)

// GetMembershipStatus returns the current membership status, using cache if available.
func GetMembershipStatus() *MembershipStatus {
	cachedStatusMu.RLock()
	if cachedStatus != nil && time.Since(cachedStatusTime) < cacheExpiry {
		status := cachedStatus
		cachedStatusMu.RUnlock()
		return status
	}
	cachedStatusMu.RUnlock()

	return checkMembership()
}

// checkMembership performs the full membership check flow.
func checkMembership() *MembershipStatus {
	deviceCode := GenerateDeviceCodeV6()
	cachedDeviceCode = deviceCode

	defaultStatus := &MembershipStatus{
		Tier:       "普通用户",
		IsMember:   false,
		DeviceCode: deviceCode,
	}

	// Debug versions (below 1.0.0) bypass membership verification
	if isDebugVersion() {
		log.Info().Str("version", appVersion).Msg("Debug version detected, bypassing membership verification")
		return &MembershipStatus{
			Tier:          "金Doro会员",
			PlanCode:      "debug",
			PlanName:      "金Doro会员调试订阅",
			StartsOn:      "00000000",
			ExpiresOn:     "99991231",
			RemainingDays: 9999,
			IsMember:      true,
			DeviceCode:    deviceCode,
		}
	}

	log.Info().
		Str("cpu_hash", shortHash(deviceCode.CPUHash)).
		Str("uuid_hash", shortHash(deviceCode.UUIDHash)).
		Msg("Generated V6 device code")

	response, err := fetchMemberStatus(deviceCode)
	if err != nil {
		log.Warn().Err(err).Msg("Membership verification unavailable, treating as non-member for this check")
		return defaultStatus
	}

	if !response.Matched {
		log.Info().Int("score", response.Score).Msg("No matching member device found")
		return defaultStatus
	}

	if !response.IsMember {
		log.Info().Str("user_id", response.UserID).Int("score", response.Score).Msg("Member device matched but no active subscription")
		defaultStatus.UserID = response.UserID
		return defaultStatus
	}

	status := statusFromResponse(response, deviceCode)
	log.Info().
		Str("user_id", status.UserID).
		Int("score", response.Score).
		Str("tier", status.Tier).
		Str("plan_code", status.PlanCode).
		Str("plan_name", status.PlanName).
		Str("expiry", status.ExpiresOn).
		Int("remaining_days", status.RemainingDays).
		Msg("Matched active member subscription")

	cacheStatus(status)
	return status
}

func cacheStatus(status *MembershipStatus) {
	cachedStatusMu.Lock()
	cachedStatus = status
	cachedStatusTime = time.Now()
	cachedStatusMu.Unlock()
}

func statusFromResponse(response *MemberStatusResponse, deviceCode DeviceCodeV6) *MembershipStatus {
	tier := response.Tier
	if tier == "" {
		tier = "普通用户"
	}

	return &MembershipStatus{
		Tier:          tier,
		PlanCode:      response.PlanCode,
		PlanName:      response.PlanName,
		StartsOn:      response.StartsOn,
		ExpiresOn:     response.ExpiresOn,
		RemainingDays: response.RemainingDays,
		IsMember:      response.IsMember,
		UserID:        response.UserID,
		DeviceCode:    deviceCode,
	}
}

func fetchMemberStatus(deviceCode DeviceCodeV6) (*MemberStatusResponse, error) {
	client := &http.Client{Timeout: httpTimeout}
	payload, err := json.Marshal(deviceCode)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 1; attempt <= maxFetchAttempts; attempt++ {
		startedAt := time.Now()
		status, statusCode, err := fetchMemberStatusOnce(client, payload)
		duration := time.Since(startedAt)
		if err == nil {
			log.Info().
				Int("attempt", attempt).
				Int("status", statusCode).
				Bool("matched", status.Matched).
				Bool("is_member", status.IsMember).
				Dur("duration", duration).
				Msg("Fetched membership status")
			return status, nil
		}

		lastErr = err
		log.Warn().
			Int("attempt", attempt).
			Int("status", statusCode).
			Dur("duration", duration).
			Err(err).
			Msg("Failed to fetch membership status")

		if !shouldRetryFetch(statusCode, err) || attempt == maxFetchAttempts {
			break
		}
		time.Sleep(time.Duration(attempt*300) * time.Millisecond)
	}

	return nil, fmt.Errorf("failed to fetch membership status: %w", lastErr)
}

func fetchMemberStatusOnce(client *http.Client, payload []byte) (*MemberStatusResponse, int, error) {
	req, err := http.NewRequest("POST", MemberStatusURL, bytes.NewReader(payload))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", "MDA/"+appVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, resp.StatusCode, fmt.Errorf("HTTP %d from membership status source: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	var status MemberStatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to parse membership status JSON: %w", err)
	}

	return &status, resp.StatusCode, nil
}

func shouldRetryFetch(statusCode int, err error) bool {
	if statusCode >= http.StatusInternalServerError {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return statusCode == 0
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
