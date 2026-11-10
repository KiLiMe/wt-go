package sys

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ArchiveSettings keeps a copy of settings.json that turned out not to be
// settings, so that the next save does not lose it, and answers with the path
// it wrote. It never replaces the file the copy is taken from.
func ArchiveSettings(data []byte) (string, error) {
	path, err := SettingsPath()
	if err != nil {
		return "", err
	}
	return archiveFile(path, data, time.Now())
}

// archiveFile writes data next to path, under a name that says it is the bad
// copy and when it was taken. A name already taken gets a number, so two bad
// copies in the same second both survive.
func archiveFile(path string, data []byte, now time.Time) (string, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("sys: create %s: %w", dir, err)
	}
	base := filepath.Join(dir, filepath.Base(path)+".bad-"+now.Format("20060102-150405"))
	for n := 1; n <= 100; n++ {
		name := base
		if n > 1 {
			name = fmt.Sprintf("%s-%d", base, n)
		}
		f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("sys: keep %s: %w", name, err)
		}
		if _, err := f.Write(data); err != nil {
			f.Close()
			os.Remove(name)
			return "", fmt.Errorf("sys: keep %s: %w", name, err)
		}
		if err := f.Close(); err != nil {
			os.Remove(name)
			return "", fmt.Errorf("sys: keep %s: %w", name, err)
		}
		return name, nil
	}
	return "", fmt.Errorf("sys: keep %s: too many names taken", base)
}
