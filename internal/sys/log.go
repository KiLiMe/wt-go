package sys

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// logName is the file the window writes what it says to, rotated in one
	// generation to logName+".1".
	logName = "wt-go.log"
	// logExt ends both that file and a crash log; crashName starts the latter,
	// and crash.go adds the time.
	logExt    = ".log"
	crashName = "crash-"
)

// logMax is how large wt-go.log may grow before it is moved aside at startup:
// a log that never rotates fills a disk, and a log that rotates on every line
// loses the run that is being read.
const logMax = 1 << 20

// LogPath is the file the window writes what it says to. A packaged window
// has no console: without this file a bad keybinding or a shell that would
// not start is a silent failure.
func LogPath() (string, error) {
	dir, err := LogDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, logName), nil
}

// IsLogPath reports whether p names a log wt-go itself wrote: wt-go.log or
// wt-go.log.1 directly under LogDir(), or a crash-*.log directly under
// CrashDir(). The MCP server reads a path handed to it by a client, so a path
// outside those two directories is refused rather than opened: the shape of
// the name alone would let ../ reach any file on the disk.
func IsLogPath(p string) bool {
	if p == "" {
		return false
	}
	// A trailing separator names a directory, not a file: Clean would drop it
	// and let "wt-go.log/" pass as the log file, which no reader can open.
	if strings.HasSuffix(p, string(filepath.Separator)) || strings.HasSuffix(p, "/") {
		return false
	}
	// A relative path is relative to the working directory, not to either
	// directory, so it never names a file there: "wt-go.log" alone would
	// otherwise pass the containment test while naming a file the caller's
	// directory holds.
	if !filepath.IsAbs(p) {
		return false
	}
	// Clean: "a/../b" and "a//b" name the same file, and the comparison must
	// be about the file, not the spelling.
	p = filepath.Clean(p)
	base := filepath.Base(p)
	if runtime.GOOS == "windows" {
		// Windows compares names case-insensitively, so a path spelled with
		// another case is still the log on disk.
		base = strings.ToLower(base)
	}
	var dirOf func() (string, error)
	switch {
	case base == logName || base == logName+".1":
		dirOf = LogDir
	case isCrashLog(base):
		dirOf = CrashDir
	default:
		return false
	}
	dir, err := dirOf()
	if err != nil {
		return false
	}
	// Rel is the containment test: anything that leaves dir ("..") or is
	// absolute on its own (a UNC share, another drive) is refused. The name
	// must sit directly in dir, so a subdirectory is refused too.
	rel, err := filepath.Rel(filepath.Clean(dir), p)
	if err != nil {
		return false
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return !strings.ContainsRune(rel, filepath.Separator)
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
