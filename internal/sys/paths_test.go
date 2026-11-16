package sys

import (
	"os"
	"path/filepath"
	"testing"
)

// withExecutableDir makes the process look like it runs from dir, which is
// where a portable copy's data directory would be.
func withExecutableDir(t *testing.T, dir string) {
	t.Helper()
	old := executableDir
	executableDir = func() (string, error) { return dir, nil }
	t.Cleanup(func() { executableDir = old })
}

// TestPortableDirs covers a copy with a data directory beside the executable:
// every file it writes goes under that directory, in the subdirectory named
// after what the file is.
func TestPortableDirs(t *testing.T) {
	root := t.TempDir()
	data := filepath.Join(root, dataName)
	if err := os.MkdirAll(data, 0o755); err != nil {
		t.Fatal(err)
	}
	withExecutableDir(t, root)

	if !Portable() {
		t.Fatal("Portable() = false with a data directory beside the exe")
	}
	if dir, err := Dir(); err != nil || dir != data {
		t.Errorf("Dir() = %q, %v; want %q", dir, err, data)
	}
	dirs := []struct {
		name string
		got  func() (string, error)
	}{
		{configName, ConfigDir},
		{logsName, LogDir},
		{crashesName, CrashDir},
	}
	for _, d := range dirs {
		want := filepath.Join(data, d.name)
		if got, err := d.got(); err != nil || got != want {
			t.Errorf("%s dir = %q, %v; want %q", d.name, got, err, want)
		}
	}
	// mygo keeps the language it was told to speak in the same directory, so
	// that it travels with the folder.
	if got, want := PortableConfigDir(), filepath.Join(data, configName); got != want {
		t.Errorf("PortableConfigDir() = %q, want %q", got, want)
	}
	// The files the window writes land in those directories, not beside them.
	files := []struct {
		name string
		got  func() (string, error)
		dir  string
	}{
		{"settings", SettingsPath, configName},
		{"window state", WindowStatePath, configName},
		{"log", LogPath, logsName},
	}
	for _, f := range files {
		got, err := f.got()
		if err != nil {
			t.Fatalf("%s: %v", f.name, err)
		}
		if want := filepath.Join(data, f.dir); filepath.Dir(got) != want {
			t.Errorf("%s = %q, want it under %q", f.name, got, want)
		}
	}
	// The MCP server reads a path a client hands it, so the containment test
	// has to follow the files: a log of either directory is read, and one that
	// is in the wrong directory is not.
	if !IsLogPath(filepath.Join(data, logsName, "wt-go.log")) {
		t.Error("IsLogPath refused the log in logs/")
	}
	if !IsLogPath(filepath.Join(data, crashesName, "crash-19700101-000000.log")) {
		t.Error("IsLogPath refused the crash log in crashes/")
	}
	if IsLogPath(filepath.Join(data, logsName, "crash-19700101-000000.log")) {
		t.Error("IsLogPath took a crash log from the log directory")
	}
}

// TestInstalledDirs covers a copy with no data directory: the files stay where
// the app has always kept them, all in one directory, so that an upgrade finds
// them.
func TestInstalledDirs(t *testing.T) {
	root := t.TempDir()
	withExecutableDir(t, root)
	base := t.TempDir()
	t.Setenv("LOCALAPPDATA", base)
	t.Setenv("XDG_CONFIG_HOME", base)
	t.Setenv("HOME", base)

	if Portable() {
		t.Fatal("Portable() = true with no data directory beside the exe")
	}
	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir: %v", err)
	}
	for name, got := range map[string]func() (string, error){
		configName:  ConfigDir,
		logsName:    LogDir,
		crashesName: CrashDir,
	} {
		if d, err := got(); err != nil || d != dir {
			t.Errorf("%s dir = %q, %v; want %q", name, d, err, dir)
		}
	}
	if got := PortableConfigDir(); got != "" {
		t.Errorf("PortableConfigDir() = %q, want no override", got)
	}
	// What "portable" means is a directory: a file named data says nothing,
	// since nothing could be written into it.
	if err := os.WriteFile(filepath.Join(root, dataName), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if Portable() {
		t.Error("Portable() = true for a file named data")
	}
}
