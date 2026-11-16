package app

import (
	"testing"
	"time"

	"github.com/egoist/mygo/plugins/terminal"
)

// TestLinkDir covers where a printed relative path is opened from: the
// directory the shell reported last when it reports one, else the one the
// terminal started in.
func TestLinkDir(t *testing.T) {
	if err := terminal.Load(); err != nil {
		t.Skip(err)
	}
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
	if err := terminal.Load(); err != nil {
		t.Skip(err)
	}
	conn := &report{data: []byte("\x1b]7;file:///tmp/somewhere\x07")}
	term, err := terminal.New(terminal.Options{Conn: conn, NoExitMessage: true})
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	deadline := time.Now().Add(5 * time.Second)
	for term.Dir() == "" {
		if time.Now().After(deadline) {
			t.Fatal("the terminal never read the reported directory")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if got := linkDir(term, "/start"); got != "/tmp/somewhere" {
		t.Errorf("linkDir = %q, want the reported directory", got)
	}
}

// report is the other end of a terminal that says one thing and then falls
// silent, which is how a shell reports its directory once.
// report is the other end of a terminal that says one thing and then falls
// silent, which is how a shell reports its directory once.
type report struct {
	data []byte
	done bool
}

func (r *report) Read(p []byte) (int, error) {
	if r.done {
		select {}
	}
	r.done = true
	return copy(p, r.data), nil
}

func (r *report) Write(p []byte) (int, error) { return len(p), nil }
func (r *report) Close() error                { return nil }

// TestPlace covers the place in a file the terminal appends to a path it
// finds: :line and :line:col come off, and what only looks like one stays.
func TestPlace(t *testing.T) {
	tests := []struct {
		path string
		want int
	}{
		{"main.go", 0},
		{"main.go:12", 3},
		{"main.go:12:3", 5},
		{`C:\src\main.go`, 0},
		{`C:\src\main.go:12`, 3},
		{"/usr/src/main.go:3:14", 5},
		// A name that only looks like a place: nothing is there to open.
		{"12:30", 3},
		{"main.go:", 0},
		{"", 0},
	}
	for _, tt := range tests {
		if got := place(tt.path); got != tt.want {
			t.Errorf("place(%q) = %d, want %d", tt.path, got, tt.want)
		}
	}
}

// TestPrintedPath covers what a printed path is once the place is off: the
// path a program would open.
func TestPrintedPath(t *testing.T) {
	tests := []struct {
		link string
		want string
	}{
		{"./build/app.js", "./build/app.js"},
		{"./build/app.js:12", "./build/app.js"},
		{`C:\src\main.go:12:3`, `C:\src\main.go`},
		{"main.go:12", "main.go"},
	}
	for _, tt := range tests {
		if got := printedPath(tt.link); got != tt.want {
			t.Errorf("printedPath(%q) = %q, want %q", tt.link, got, tt.want)
		}
	}
}

// TestIsURL covers which links this window leaves to the desktop: a scheme
// says it is a URL, and a drive letter does not.
func TestIsURL(t *testing.T) {
	tests := []struct {
		link string
		want bool
	}{
		{"https://example.com/a?b=1", true},
		{"file:///C:/x/a.txt", true},
		{"mailto:me@example.com", true},
		{"vscode://file/C:/x", true},
		{`C:\src\main.go`, false},
		{"main.go", false},
		{"src/a:b", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isURL(tt.link); got != tt.want {
			t.Errorf("isURL(%q) = %v, want %v", tt.link, got, tt.want)
		}
	}
}
