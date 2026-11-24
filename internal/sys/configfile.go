package sys

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Boundary: a file of the settings directory -- where it lives, and reading and
// writing it as one file. What the settings and the window's place hold is
// settings.go and windowstate.go.

// configFile is a file of the settings directory by name.
func configFile(name string) (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}

// readConfigFile is a file of the settings directory, and whether there was
// one. A window that finds none starts with its defaults and writes the file,
// so that there is one to edit.
func readConfigFile(name string) ([]byte, bool, error) {
	path, err := configFile(name)
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

// writeConfigFile makes the settings directory and writes a file of it.
func writeConfigFile(name string, b []byte) error {
	path, err := configFile(name)
	if err != nil {
		return err
	}
	return writeFileAtomic(path, b, 0o644)
}
