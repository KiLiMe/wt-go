package app

import (
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/egoist/mygo/plugins/terminal"
)

// TestLinkDir covers where a printed relative path is opened from: the
// directory the shell reported last when it reports one, else the one the
// terminal started in.
func TestLinkDir(t *testing.T) {
	if got := linkDir(nil, `C:\src`); got != `C:\src` {
		t.Errorf("no terminal: linkDir = %q, want the start dir", got)
	}
	if got := linkDir(headlessTerminal(t), `C:\src`); got != `C:\src` {
		t.Errorf("no OSC 7 yet: linkDir = %q, want the start dir", got)
	}
}

// TestLinkDirFollowsOSC7 covers a shell that reports its directory: the
// relative path printed after a cd is opened from where the shell went.
func TestLinkDirFollowsOSC7(t *testing.T) {
	// A shell reports a file URL; on Windows the plugin hands the drive back
	// with a slash in front of it (see TestLinkDirKeepsADriveWhole), so the
	// path alone is what is reported there.
	reported, start := "/tmp/somewhere", "/start"
	if runtime.GOOS == "windows" {
		reported, start = `C:\somewhere`, `C:\start`
	}
	term := reportingTerminal(t, "\x1b]7;"+reported+"\x07")
	defer term.Close()
	awaitDir(t, term)
	if got := linkDir(term, start); got != reported {
		t.Errorf("linkDir = %q, want the reported directory", got)
	}
}

// TestLinkDirKeepsADriveWhole covers the shape a file URL of a Windows drive
// takes: the slash the URL puts in front of the drive is not part of it, so a
// relative path printed there is joined under the drive itself. It fails if
// the plugin goes back to handing out /C:/src, which is no path on Windows.
func TestLinkDirKeepsADriveWhole(t *testing.T) {
	term := reportingTerminal(t, "\x1b]7;file:///C:/src\x07")
	defer term.Close()
	awaitDir(t, term)
	dir := term.Dir()
	if !filepath.IsAbs(dir) {
		if runtime.GOOS == "windows" {
			t.Fatalf("the plugin reports %q, which is no absolute path on Windows", dir)
		}
		t.Skipf("the plugin reports %q, which is no path here", dir)
	}
	if got := linkDir(term, `C:\start`); got != dir {
		t.Errorf("linkDir = %q, want the reported %q", got, dir)
	}
}

// reportingTerminal is a terminal whose other end says what a shell says once
// and then falls silent, as when it reports its directory.
func reportingTerminal(t *testing.T, seq string) *terminal.Terminal {
	t.Helper()
	return headlessTerminalWith(t, &report{data: []byte(seq)})
}

// awaitDir waits for the terminal to read a reported directory, and for the
// plugin to hand it out.
func awaitDir(t *testing.T, term *terminal.Terminal) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for term.Dir() == "" {
		if time.Now().After(deadline) {
			t.Fatal("the terminal never read the reported directory")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// report is the other end of a terminal that says one thing and then falls
// silent, which is how a shell reports its directory once.
type report struct {
	sink
	data []byte
	sent bool
}

func (r *report) Read(p []byte) (int, error) {
	if r.sent {
		select {}
	}
	r.sent = true
	return copy(p, r.data), nil
}

// TestPathWithoutPlace covers the place in a file the terminal appends to a
// path it finds: :line and :line:col come off, and the colon of a drive does
// not.
func TestPathWithoutPlace(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"main.go", "main.go"},
		{"main.go:12", "main.go"},
		{"main.go:12:3", "main.go"},
		{`C:\src\main.go`, `C:\src\main.go`},
		{`C:\src\main.go:12`, `C:\src\main.go`},
		{`C:\src\main.go:12:3`, `C:\src\main.go`},
		{"/usr/src/main.go:3:14", "/usr/src/main.go"},
		{"main.go:", "main.go:"},
		{"", ""},
		// A time is not a place, but it reads as one: the tail comes off,
		// and the click opens nothing as no file is called 12.
		{"12:30", "12"},
	}
	for _, tt := range tests {
		if got := pathWithoutPlace(tt.path); got != tt.want {
			t.Errorf("pathWithoutPlace(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

// TestLinkPath covers both halves of a printed link: a scheme is a URL, which
// the terminal opens with the desktop, and a path comes back with the place in
// the file off, as the path a program would open.
func TestLinkPath(t *testing.T) {
	tests := []struct {
		link string
		want string
	}{
		// URLs are the terminal's business.
		{"https://example.com/a?b=1", ""},
		{"file:///C:/x/a.txt", ""},
		{"mailto:me@example.com", ""},
		{"vscode://file/C:/x", ""},
		// A drive letter is no scheme: the path stays, without its place.
		{`C:\src\main.go`, `C:\src\main.go`},
		{`C:\src\main.go:12:3`, `C:\src\main.go`},
		{"src/a:b", "src/a:b"},
		{"./build/app.js", "./build/app.js"},
		{"./build/app.js:12", "./build/app.js"},
		{"main.go:12", "main.go"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := linkPath(tt.link); got != tt.want {
			t.Errorf("linkPath(%q) = %q, want %q", tt.link, got, tt.want)
		}
	}
}
