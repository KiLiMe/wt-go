package app

import "testing"

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
