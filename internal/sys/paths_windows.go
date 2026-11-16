//go:build windows

package sys

import (
	"errors"
	"os"
	"path/filepath"
)

// defaultDir is where an installed copy keeps its files:
// %LOCALAPPDATA%\wt-go. A copy with a data directory beside the executable
// keeps them there instead (paths.go).
func defaultDir() (string, error) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		return "", errors.New("sys: LOCALAPPDATA is not set")
	}
	return filepath.Join(base, "wt-go"), nil
}
