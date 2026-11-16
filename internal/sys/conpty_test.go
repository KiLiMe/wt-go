package sys

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestConPTYLibraryPrefersTheConptyDir covers the copy the build ships: the
// three files sit in conpty\ beside the exe, and that is the one the report
// names and the terminal plugin loads, over a conpty.dll a person may have
// left beside the exe from an older unpack.
func TestConPTYLibraryPrefersTheConptyDir(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("ConPTY is Windows' own")
	}
	root := t.TempDir()
	withExecutableDir(t, root)
	// adopted is the package's, not the test's: a test that sets it leaves it
	// for the next one otherwise.
	t.Cleanup(func() { adopted = "" })
	write := func(path string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("not a library"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	bundled := filepath.Join(root, conptyDirName)
	write(filepath.Join(bundled, conptyName))
	write(filepath.Join(root, conptyName))

	t.Setenv("MYGO_CONPTY", "")
	path, fromEnv, err := ConPTYLibrary()
	if err != nil {
		t.Fatalf("ConPTYLibrary: %v", err)
	}
	if want := filepath.Join(bundled, conptyName); path != want {
		t.Errorf("ConPTYLibrary = %q, want %q", path, want)
	}
	if fromEnv {
		t.Error("ConPTYLibrary said the environment named it")
	}

	// The window sets the variable the plugin reads, at the same directory, so
	// that the console a terminal is made in is the one the report named.
	UseBundledConPTY()
	if got := os.Getenv("MYGO_CONPTY"); got != bundled {
		t.Errorf("MYGO_CONPTY = %q, want %q", got, bundled)
	}
	// The variable then holds the program's own pick, which is not a person's:
	// the report must say the ConPTY beside the program, not one chosen.
	path, fromEnv, err = ConPTYLibrary()
	if want := filepath.Join(bundled, conptyName); path != want || fromEnv || err != nil {
		t.Errorf("after UseBundledConPTY: ConPTYLibrary = %q, %v, %v; want %q, not from the environment", path, fromEnv, err, want)
	}

	// A directory a person names is theirs, and the report says so.
	mine := t.TempDir()
	write(filepath.Join(mine, conptyName))
	t.Setenv("MYGO_CONPTY", mine)
	adopted = ""
	path, fromEnv, err = ConPTYLibrary()
	if want := filepath.Join(mine, conptyName); path != want || !fromEnv || err != nil {
		t.Errorf("a directory MYGO_CONPTY names: ConPTYLibrary = %q, %v, %v; want %q, from the environment", path, fromEnv, err, want)
	}

	// A person who set the variable decides, including "system", which is how
	// a machine asks for Windows' own console.
	t.Setenv("MYGO_CONPTY", "system")
	UseBundledConPTY()
	if got := os.Getenv("MYGO_CONPTY"); got != "system" {
		t.Errorf("MYGO_CONPTY = %q, want the value a person set", got)
	}
}

// TestConPTYLibraryWithoutTheConptyDir covers a copy unpacked before the
// directory existed, and one unpacked without it: the flat file is still
// found, and with neither Windows' own console answers.
func TestConPTYLibraryWithoutTheConptyDir(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("ConPTY is Windows' own")
	}
	root := t.TempDir()
	withExecutableDir(t, root)
	t.Setenv("MYGO_CONPTY", "")
	t.Cleanup(func() { adopted = "" })

	if _, _, err := ConPTYLibrary(); err == nil {
		t.Error("ConPTYLibrary found a conpty.dll in an empty directory")
	}
	UseBundledConPTY()
	if got := os.Getenv("MYGO_CONPTY"); got != "" {
		t.Errorf("MYGO_CONPTY = %q, want it left alone", got)
	}

	flat := filepath.Join(root, conptyName)
	if err := os.WriteFile(flat, []byte("not a library"), 0o600); err != nil {
		t.Fatal(err)
	}
	path, fromEnv, err := ConPTYLibrary()
	if err != nil {
		t.Fatalf("ConPTYLibrary: %v", err)
	}
	if path != flat || fromEnv {
		t.Errorf("ConPTYLibrary = %q, %v; want the flat file", path, fromEnv)
	}
}
