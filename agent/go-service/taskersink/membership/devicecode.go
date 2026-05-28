package membership

import (
	"crypto/sha256"
	"encoding/hex"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

// DeviceCodeV7 holds six independent SHA-256 hardware hashes.
type DeviceCodeV7 struct {
	CPUHash   string `json:"cpu_hash"`
	UUIDHash  string `json:"uuid_hash"`
	BIOSHash  string `json:"bios_hash"`
	BoardHash string `json:"board_hash"`
	DiskHash  string `json:"disk_hash"`
	GUIDHash  string `json:"guid_hash"`
}

// v7Weights defines the match weight for each hardware hash.
var v7Weights = map[string]int{
	"cpu":   15,
	"uuid":  45,
	"bios":  5,
	"board": 10,
	"disk":  15,
	"guid":  10,
}

// GenerateDeviceCodeV7 generates a V7 device code by querying hardware identifiers.
func GenerateDeviceCodeV7() DeviceCodeV7 {
	code := DeviceCodeV7{
		CPUHash:   hashString(queryWMI("Win32_Processor", "ProcessorID")),
		UUIDHash:  hashString(queryWMI("Win32_ComputerSystemProduct", "UUID")),
		BIOSHash:  hashString(queryWMIFiltered("Win32_BIOS", "SerialNumber", notPlaceholder)),
		BoardHash: hashString(queryWMIFiltered("Win32_BaseBoard", "SerialNumber", notPlaceholder)),
		DiskHash:  hashString(queryWMIFirstFixed("Win32_DiskDrive", "SerialNumber")),
		GUIDHash:  hashString(readMachineGuid()),
	}
	return code
}

// MatchDeviceCodeV7 performs weighted matching between current and saved device codes.
// Returns the match score (0-100). Threshold >= 80 means a match.
func MatchDeviceCodeV7(current, saved DeviceCodeV7) int {
	score := 0
	if current.CPUHash != "" && current.CPUHash == saved.CPUHash {
		score += v7Weights["cpu"]
	}
	if current.UUIDHash != "" && current.UUIDHash == saved.UUIDHash {
		score += v7Weights["uuid"]
	}
	if current.BIOSHash != "" && current.BIOSHash == saved.BIOSHash {
		score += v7Weights["bios"]
	}
	if current.BoardHash != "" && current.BoardHash == saved.BoardHash {
		score += v7Weights["board"]
	}
	if current.DiskHash != "" && current.DiskHash == saved.DiskHash {
		score += v7Weights["disk"]
	}
	if current.GUIDHash != "" && current.GUIDHash == saved.GUIDHash {
		score += v7Weights["guid"]
	}
	return score
}

func hashString(s string) string {
	if s == "" {
		return ""
	}
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func notPlaceholder(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || s == "UNKNOWN" {
		return false
	}
	lower := strings.ToLower(s)
	if strings.Contains(lower, "to be filled") || strings.Contains(lower, "default string") {
		return false
	}
	return true
}

// queryWMI runs a PowerShell WMI query and returns the first value of the specified property.
func queryWMI(class, property string) string {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		"Get-CimInstance -ClassName "+class+" | Select-Object -First 1 -ExpandProperty "+property)
	out, err := cmd.Output()
	if err != nil {
		log.Debug().Str("class", class).Str("property", property).Err(err).Msg("WMI query failed")
		return ""
	}
	return strings.TrimSpace(string(out))
}

// queryWMIFiltered runs a WMI query and returns the first value that passes the filter.
func queryWMIFiltered(class, property string, filter func(string) bool) string {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		"Get-CimInstance -ClassName "+class+" | ForEach-Object { $_."+property+" }")
	out, err := cmd.Output()
	if err != nil {
		log.Debug().Str("class", class).Str("property", property).Err(err).Msg("WMI query failed")
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		val := strings.TrimSpace(line)
		if filter(val) {
			return val
		}
	}
	return "UNKNOWN"
}

// queryWMIFirstFixed queries Win32_DiskDrive for the first fixed disk serial number.
func queryWMIFirstFixed(class, property string) string {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		"Get-CimInstance -ClassName "+class+" -Filter \"MediaType='Fixed hard disk media'\" | Select-Object -First 1 -ExpandProperty "+property)
	out, err := cmd.Output()
	if err != nil {
		log.Debug().Str("class", class).Err(err).Msg("WMI query failed for fixed disk")
		return ""
	}
	result := strings.TrimSpace(string(out))
	if result == "" {
		return "UNKNOWN"
	}
	return result
}

// readMachineGuid reads the Windows MachineGuid from the registry.
func readMachineGuid() string {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		"Get-ItemPropertyValue -Path 'HKLM:\\SOFTWARE\\Microsoft\\Cryptography' -Name MachineGuid")
	out, err := cmd.Output()
	if err != nil {
		log.Debug().Err(err).Msg("Failed to read MachineGuid from registry")
		return "UNKNOWN"
	}
	result := strings.TrimSpace(string(out))
	if result == "" {
		return "UNKNOWN"
	}
	return result
}
