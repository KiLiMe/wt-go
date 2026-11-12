package sys

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// WindowStatePath is where the window's place, size and state live: beside
// the settings, since the geometry a window was left in is a preference too.
func WindowStatePath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "window-state.json"), nil
}

// ReadWindowState is window-state.json, and whether there was one.
func ReadWindowState() ([]byte, bool, error) {
	path, err := WindowStatePath()
	if err != nil {
		return nil, false, err
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("sys: %w", err)
	}
	return b, true, nil
}

// WriteWindowState makes the settings directory and writes window-state.json.
func WriteWindowState(b []byte) error {
	path, err := WindowStatePath()
	if err != nil {
		return err
	}
	return writeFileAtomic(path, b, 0o644)
}
