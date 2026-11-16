//go:build !windows

package sys

import (
	"fmt"
	"os"
	"path/filepath"
)

// defaultDir is where an installed copy keeps its files:
// $XDG_CONFIG_HOME/wt-go, or ~/.config/wt-go. A copy with a data directory
// beside the executable keeps them there instead (paths.go).
func defaultDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("sys: %w", err)
	}
	return filepath.Join(base, "wt-go"), nil
}
