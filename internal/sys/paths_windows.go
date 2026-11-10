//go:build windows

package sys

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Dir is where wt-go keeps its files: %LOCALAPPDATA%\wt-go.
func Dir() (string, error) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		return "", errors.New("sys: LOCALAPPDATA is not set")
	}
	return filepath.Join(base, "wt-go"), nil
}

// SettingsPath is the file the settings live in.
func SettingsPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.json"), nil
}

// ReadSettings is settings.json, and whether there was one. A window that
// finds none starts with the defaults and writes them, so the file is there
// to edit.
func ReadSettings() ([]byte, bool, error) {
	path, err := SettingsPath()
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

// WriteSettings makes the settings directory and writes settings.json.
func WriteSettings(b []byte) error {
	path, err := SettingsPath()
	if err != nil {
		return err
	}
	return writeFileAtomic(path, b, 0o644)
}
