package membership

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestHashConsistency(t *testing.T) {
	testCases := map[string]string{
		"CPU":   "178BFBFF00B40F40",
		"UUID":  "9CF90480-50DA-ADBA-CB6C-BCFCE7CD3B82",
		"BIOS":  "System Serial Number",
		"Board": "250454514000142",
		"Disk":  "0000_0000_0000_0000_A428_B701_77BF_0003.",
		"GUID":  "a83696f9-a692-4c68-9717-3637e90ce463",
	}

	expected := map[string]string{
		"CPU":   "1561b5a4fc04bb159ca2da9d5f2dec4b8b3a5ea03d1ec31c441156698acbd6bb",
		"UUID":  "f31b1eac0049b9ba4ea44dce09ce31c8987b21d2ca2e30a53d6cc0abddd45a37",
		"BIOS":  "237f23730628f9cdbf7e414f2e2098da90a1301f62f538920ac403ca3f3dd590",
		"Board": "2d5e82331f6f9e485c15da816c15fc9c745548e88fde35e0b9ae092bd89bc76e",
		"Disk":  "53a013f79f8e8cc0f4f96ecf2ce660ee5dd53c13eca7eb005cba5d9f556070cf",
		"GUID":  "25a63f5fe7d9c0ebb0fef90d2225d598c18ef8e124981bd1f16036e83ebaa4e6",
	}

	allMatch := true
	for key, value := range testCases {
		h := sha256.Sum256([]byte(value))
		actual := hex.EncodeToString(h[:])
		match := actual == expected[key]
		status := "✅"
		if !match {
			status = "❌"
			allMatch = false
		}
		fmt.Printf("%s %s:\n  MDA:     %s\n  DoroPay: %s\n", status, key, actual, expected[key])
	}

	if allMatch {
		fmt.Println("\n✅ All hashes match!")
	} else {
		t.Fatal("❌ Some hashes do not match!")
	}
}
