package app

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"wtgo/internal/fault"
	"wtgo/internal/sys"
)

// openLink is what a Control+click on a link in a terminal does. A URL is
// left to the terminal, which opens it with the desktop; a file path printed
// as text is a link only because this window answers for it (the terminal's
// OpenLink), so it is opened here, from dir, where the program that printed
// it started.
func (a *App) openLink(dir string) func(string) bool {
	return func(link string) bool {
		if isURL(link) {
			return false
		}
		path := printedPath(link)
		if !filepath.IsAbs(path) {
			path = filepath.Join(dir, path)
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
