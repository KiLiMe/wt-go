package app

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo/plugins/terminal"

	"wtgo/internal/fault"
	"wtgo/internal/sys"
)

// openLink is what a Control+click on a link in a terminal does. A URL is
// left to the terminal, which opens it with the desktop; a file path printed
// as text is a link only because this window answers for it (the terminal's
// OpenLink), so it is opened here, relative to dirOf: the directory the shell
// reported last (OSC 7), or the one the terminal started in.
func (a *App) openLink(dirOf func() string) func(string) bool {
	return func(link string) bool {
		if isURL(link) {
			return false
		}
		path := printedPath(link)
		if !filepath.IsAbs(path) {
			path = filepath.Join(dirOf(), path)
		}
		// A path that is not there is not opened: explorer.exe would show a
		// folder nobody asked for, and a click in a terminal has nowhere on
		// screen to say why nothing happened.
		if _, err := os.Stat(path); err != nil {
			return true
		}
		if err := sys.OpenPath(path); err != nil {
			log.Printf("wt-go: %s", fault.Line(fault.OpenLink, err.Error()))
		}
		return true
	}
}

// linkDir is the directory a printed path is opened from: the one the shell
// reported last (OSC 7), else the one the terminal started in. A shell that
// never reports its directory, as PowerShell does on Windows, leaves the
// second, which is where a relative path printed before any cd resolves.
// A reported directory that is not absolute is not used either: a file URL
// of a Windows drive comes out of the plugin as /C:/src, which would put
// every relative path on a drive that does not exist.
func linkDir(term *terminal.Terminal, start string) string {
	if term != nil {
		if dir := term.Dir(); filepath.IsAbs(dir) {
			return dir
		}
	}
	return start
}

// isURL says whether link carries a scheme: what the terminal finds as a URL,
// OSC 8 or spelled out in the text, is none of this window's business. A path
// has no scheme, and the colon of a drive letter does not make one.
func isURL(link string) bool {
	if strings.HasPrefix(link, "mailto:") {
		return true
	}
	i := strings.Index(link, "://")
	return i > 0 && !strings.ContainsAny(link[:i], `/\`)
}

// printedPath is link with the home directory and the variables expanded and
// the place in the file taken off: the terminal hands over main.go:12:3,
// which no program can open as it stands.
func printedPath(link string) string {
	link = sys.ExpandPath(link)
	if n := place(link); n > 0 {
		link = link[:len(link)-n]
	}
	return link
}

// place is the length of the ":line" or ":line:col" a path ends with, or 0:
// the colon after a drive letter is not a place, or a path on C: would lose
// its drive.
func place(path string) int {
	i := strings.LastIndexByte(path, ':')
	if i <= 1 || path[i+1:] == "" || !digits(path[i+1:]) {
		return 0
	}
	if j := strings.LastIndexByte(path[:i], ':'); j > 1 && digits(path[j+1:i]) {
		return len(path) - j
	}
	return len(path) - i
}
