package sys

import (
	"errors"
	"fmt"
	"os"
)

// RemoveSettings deletes settings.json, and counts one that is already gone as
// deleted: a window that starts next finds no file and writes the defaults.
func RemoveSettings() error {
	path, err := SettingsPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("sys: %w", err)
	}
	return nil
}
