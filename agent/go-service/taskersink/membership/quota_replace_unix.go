//go:build !windows

package membership

import (
	"os"
	"path/filepath"
)

func replaceQuotaStateFile(source string, destination string) error {
	if err := os.Rename(source, destination); err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(destination))
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}
