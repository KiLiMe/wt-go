package sys

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// SettingsFile says what the file the settings live in is, without reading
// it: a diagnostic cares whether it exists and when it was written.
type SettingsFile struct {
	Path    string
	Exists  bool
	Size    int64
	ModTime time.Time
}

// SettingsFileInfo is the file the settings live in.
func SettingsFileInfo() (SettingsFile, error) {
	path, err := SettingsPath()
	if err != nil {
		return SettingsFile{}, err
	}
	out := SettingsFile{Path: path}
	st, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return out, nil
	}
	if err != nil {
		return out, fmt.Errorf("sys: %w", err)
	}
	out.Exists = true
	out.Size = st.Size()
	out.ModTime = st.ModTime()
	return out, nil
}

// CrashLogs is the crash logs the program wrote, newest first. A window that
// died leaves one; the newest is the one a diagnostic reads.
func CrashLogs() ([]string, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("sys: read %s: %w", dir, err)
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "crash-") || !strings.HasSuffix(e.Name(), ".log") {
			continue
		}
		out = append(out, filepath.Join(dir, e.Name()))
	}
	sort.Sort(sort.Reverse(sort.StringSlice(out)))
	return out, nil
}

// Tail is the last limit bytes of the file at path, starting at a line
// boundary, and whether the file was longer than that.
func Tail(path string, limit int64) (string, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", false, fmt.Errorf("sys: %w", err)
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return "", false, fmt.Errorf("sys: %w", err)
	}
	cut := st.Size() > limit
	if cut {
		if _, err := f.Seek(st.Size()-limit, 0); err != nil {
			return "", false, fmt.Errorf("sys: %w", err)
		}
	}
	buf := make([]byte, min(st.Size(), limit))
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return "", false, fmt.Errorf("sys: %w", err)
	}
	text := string(buf[:n])
	if cut {
		// The first line is half a line: drop it, the rest is what was asked.
		if i := strings.IndexByte(text, '\n'); i >= 0 {
			text = text[i+1:]
		}
	}
	return text, cut, nil
}

// ErrNoNative is what NativeLibrary answers when no libghostty-vt is in any
// of the places the terminal plugin looks.
var ErrNoNative = errors.New("sys: libghostty-vt is not in any of the places it is looked for")

// nativeName is what libghostty-vt is called on this platform.
func nativeName() string {
	switch runtime.GOOS {
	case "windows":
		return "ghostty-vt.dll"
	case "darwin":
		return "libghostty-vt.dylib"
	}
	return "libghostty-vt.so"
}

// NativeLibrary is the libghostty-vt the terminal plugin would load, in the
// order it looks: $MYGO_GHOSTTY_VT, next to the executable, then the copies
// the mygo tool downloaded under %LOCALAPPDATA%\mygo\natives.
func NativeLibrary() (string, error) {
	name := nativeName()
	if p := os.Getenv("MYGO_GHOSTTY_VT"); p != "" {
		return p, nil
	}
	if exe, err := os.Executable(); err == nil {
		p := filepath.Join(filepath.Dir(exe), name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", ErrNoNative
	}
	dir := filepath.Join(cache, "mygo", "natives")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", ErrNoNative
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p := filepath.Join(dir, e.Name(), name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", ErrNoNative
}
