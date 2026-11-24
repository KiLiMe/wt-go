package sys

import (
	"os"
	"path/filepath"
)

// wt-go keeps its files in one directory. An installed copy puts them under
// the user's application data; a copy that has a data directory beside the
// executable keeps them in it, so that unzipping the program somewhere else
// carries the settings along (ADR-83).

const (
	// dataName is the directory beside the executable that makes a copy
	// portable.
	dataName = "data"
	// The subdirectories of that data directory, one per kind of file; an
	// installed copy keeps all three kinds loose in Dir().
	configName  = "config"
	logsName    = "logs"
	crashesName = "crashes"
)

// executableDir is the folder the program runs from. A variable, so that a
// test can stand a data directory, or a ConPTY, somewhere it may write.
var executableDir = func() (string, error) {
	exe, err := Executable()
	if err != nil {
		return "", err
	}
	// A link is not where the person unzipped: the data directory is beside
	// the real executable, next to the files it was unpacked with.
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}
	return filepath.Dir(exe), nil
}

// DataDir is the data directory beside the executable of a portable copy, and
// "" for an installed one. The directory's presence is the whole test: a file
// that happens to be named data would not hold anything.
func DataDir() string {
	dir, err := executableDir()
	if err != nil {
		return ""
	}
	p := filepath.Join(dir, dataName)
	st, err := os.Stat(p)
	if err != nil || !st.IsDir() {
		return ""
	}
	return p
}

// Portable reports whether this copy keeps its files beside itself rather than
// under the user's application data.
func Portable() bool { return DataDir() != "" }

// ConfigDir is where settings.json and window-state.json live: data\config in
// a portable copy. An installed copy predates the split and keeps the files
// loose in Dir(), which is also where an upgrade finds them.
func ConfigDir() (string, error) { return dataSubDir(configName) }

// LogDir is where wt-go.log lives: data\logs in a portable copy, Dir() in an
// installed one.
func LogDir() (string, error) { return dataSubDir(logsName) }

// CrashDir is where the crash logs live: data\crashes in a portable copy,
// Dir() in an installed one.
func CrashDir() (string, error) { return dataSubDir(crashesName) }

// PortableConfigDir is where mygo keeps the one file of its own, the language
// last chosen: a portable copy's config directory, so that a folder carried to
// another machine does not leave the language under %APPDATA%. "" for an
// installed copy, which keeps mygo's default.
func PortableConfigDir() string {
	d := DataDir()
	if d == "" {
		return ""
	}
	return filepath.Join(d, configName)
}

// Dir is the directory wt-go keeps its files in: a portable copy's data
// directory, or the place this system keeps a program's files.
func Dir() (string, error) { return dataSubDir("") }

// dataSubDir is the named subdirectory of a portable copy's data directory,
// and the place this system keeps a program's files for an installed copy,
// which has no data directory of its own to name.
func dataSubDir(name string) (string, error) {
	if d := DataDir(); d != "" {
		return filepath.Join(d, name), nil
	}
	return defaultDir()
}
