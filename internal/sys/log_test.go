package sys

import (
	"path/filepath"
	"runtime"
	"testing"
)

// TestIsLogPath covers the names the diagnostic reads and the paths a client
// would have to write to reach anything else.
func TestIsLogPath(t *testing.T) {
	dir, err := Dir()
	if err != nil {
		t.Skipf("Dir: %v", err)
	}
	cleanDir := filepath.Clean(dir)
	join := func(name string) string { return filepath.Join(cleanDir, name) }

	accepted := []string{
		join("wt-go.log"),
		join("wt-go.log.1"),
		join("crash-20250101-120000.log"),
	}
	// The path the tools hand back is clean already, but a client may spell
	// it the long way: same file, same answer.
	if !filepath.IsAbs(cleanDir) {
		t.Fatal("Dir() must be absolute")
	}
	accepted = append(accepted, cleanDir+string(filepath.Separator)+"."+string(filepath.Separator)+"crash-1.log")

	for _, p := range accepted {
		if !IsLogPath(p) {
			t.Errorf("IsLogPath(%q) = false, want true", p)
		}
	}

	rejected := []string{
		"",
		// Not a log wt-go writes, even though it sits in the directory.
		join("settings.json"),
		join("wt-go.log.2"),
		join("crash-1.txt"),
		join("crash-.log.tmp"),
		join("xcrash-1.log"),
		// Inside the directory, but not directly in it.
		filepath.Join(cleanDir, "sub", "wt-go.log"),
		// Outside it.
		filepath.Join(cleanDir, "..", "wt-go.log"),
		filepath.Join(cleanDir, "..", "..", "wt-go.log"),
		filepath.Join(cleanDir, "sub", "..", "..", "wt-go.log"),
		// A traversal that leaves and does not come back.
		join("..") + string(filepath.Separator) + "crash-1.log",
		// The directory itself.
		cleanDir,
	}
	if runtime.GOOS == "windows" {
		rejected = append(rejected,
			`C:\Windows\System32\drivers\etc\hosts`,
			`\\server\share\wt-go.log`,
			`C:\wt-go.log`,
		)
	} else {
		rejected = append(rejected, "/etc/passwd", "/tmp/wt-go.log")
	}
	for _, p := range rejected {
		if IsLogPath(p) {
			t.Errorf("IsLogPath(%q) = true, want false", p)
		}
	}
}

// A path spelled with another case is the same file on Windows, and a
// different one elsewhere: the answer follows the disk.
func TestIsLogPathCase(t *testing.T) {
	dir, err := Dir()
	if err != nil {
		t.Skipf("Dir: %v", err)
	}
	p := filepath.Join(dir, "WT-GO.LOG")
	if got, want := IsLogPath(p), runtime.GOOS == "windows"; got != want {
		t.Errorf("IsLogPath(%q) = %v, want %v on %s", p, got, want, runtime.GOOS)
	}
}

// TestIsLogPathRelative covers a relative path: it names a file under the
// working directory, not one under Dir(), however it is spelled, so the check
// refuses it rather than resolving it into the directory the logs live in.
func TestIsLogPathRelative(t *testing.T) {
	for _, p := range []string{
		"wt-go.log",
		"wt-go.log.1",
		"crash-20250101-120000.log",
		"." + string(filepath.Separator) + "wt-go.log",
		"sub" + string(filepath.Separator) + ".." + string(filepath.Separator) + "wt-go.log",
		".." + string(filepath.Separator) + "wt-go.log",
	} {
		if IsLogPath(p) {
			t.Errorf("IsLogPath(%q) = true, want false: it is relative", p)
		}
	}
}

// TestIsLogPathNames covers the names a client may hand the diagnostic: the
// ones wt-go writes pass, and everything one character off is refused rather
// than read.
func TestIsLogPathNames(t *testing.T) {
	dir, err := Dir()
	if err != nil {
		t.Skipf("Dir: %v", err)
	}
	join := func(name string) string { return filepath.Join(dir, name) }

	accepted := []string{
		join("wt-go.log"),
		join("wt-go.log.1"),
		join("crash-.log"),
		join("crash-x.log"),
		join("crash-2025.log"),
	}
	for _, p := range accepted {
		if !IsLogPath(p) {
			t.Errorf("IsLogPath(%q) = false, want true", p)
		}
	}

	rejected := []string{
		join(".wt-go.log"),
		join("wt-go.log "),
		join("wt-go.log.10"),
		join("wt-go.log.old"),
		join("wt-go.log.gz"),
		join("wt-go"),
		join("crash-1.logs"),
		join("crash-1.log.1"),
		join("crash-1"),
		join("crash"),
		join("crashx.log"),
		join("other-crash-1.log"),
		join("settings.json.bad-20250101"),
		join(string(filepath.Separator)),
		join("wt-go.log") + string(filepath.Separator),
	}
	// Windows compares names case-insensitively, so these are the logs on
	// disk there and something else elsewhere.
	if runtime.GOOS != "windows" {
		rejected = append(rejected,
			join("crash-1.LOG"),
			join("Crash-1.log"),
		)
	}
	for _, p := range rejected {
		if IsLogPath(p) {
			t.Errorf("IsLogPath(%q) = true, want false", p)
		}
	}
}
