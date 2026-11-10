package sys

import (
	"fmt"
	"os"
	"path/filepath"
)

// logMax is how large wt-go.log may grow before it is moved aside at startup:
// a log that never rotates fills a disk, and a log that rotates on every line
// loses the run that is being read.
const logMax = 1 << 20

// LogPath is the file the window writes what it says to. A packaged window
// has no console: without this file a bad keybinding or a shell that would
// not start is a silent failure.
func LogPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "wt-go.log"), nil
}

// OpenLog opens wt-go.log for appending, moving an oversized one aside first.
func OpenLog() (*os.File, error) {
	path, err := LogPath()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("sys: create %s: %w", filepath.Dir(path), err)
	}
	if st, err := os.Stat(path); err == nil && st.Size() > logMax {
		// One generation of history is enough: the failed run is the newest.
		_ = os.Rename(path, path+".1")
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("sys: open %s: %w", path, err)
	}
	return f, nil
}
