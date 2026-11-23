package app

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo/plugins/terminal"

	"wtgo/internal/fault"
	"wtgo/internal/store"
	"wtgo/internal/sys"
)

// openLink is what a click on a link in a terminal does, opened from dir. A
// file path printed as text is a link only because this window answers for it
// (the terminal's OpenLink), so it is opened here; a URL is opened here too,
// rather than left to the terminal, because the terminal drops the failure of
// the desktop's program and a click that does nothing would say nothing either.
//
// Whether a click opens anything at all is not decided here: the terminal finds
// no link while the setting says so (setLinks), which is also what keeps the
// pointer from underlining one.
func (a *App) openLink(link, dir string) bool {
	// An empty link is no link: false leaves it to the terminal, which found
	// nothing to hand over in the first place.
	if link == "" {
		return false
	}
	path := linkPath(link)
	if path == "" {
		if err := sys.OpenURL(link); err != nil {
			log.Printf("wt-go: %s", fault.Line(fault.OpenLink, err.Error()))
		}
		return true
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}
	// A path that is not there is not opened: explorer.exe would show a folder
	// nobody asked for, and a click in a terminal has nowhere on screen to say
	// why nothing happened.
	if _, err := os.Stat(path); err != nil {
		return true
	}
	if err := sys.OpenPath(path); err != nil {
		log.Printf("wt-go: %s", fault.Line(fault.OpenLink, err.Error()))
	}
	return true
}

// setLinks hands the settings to the plugin's knobs: the switch a click and
// every kind answer to, which kinds of link are one, and whether the pointer
// shows a link on a hover alone. The hint follows the switch: links off have
// nothing to underline, so the kinds are moot and the hint goes with them.
func setLinks(term *terminal.Terminal, l store.Links) {
	term.SetLinks(l.Open)
	term.SetLinkKinds(terminal.LinkKinds{Program: l.Program, URL: l.URL, Path: l.Path})
	term.SetLinkHoverHint(l.Open && l.Hover)
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

// linkPath is the path a click on a printed link opens, or "" for a URL. A
// URL's scheme says it is one, and the colon of a drive letter says nothing.
// The path comes back with the home directory and the variables expanded and
// the place in the file taken off: the terminal hands over main.go:12:3, which
// no program opens as it stands.
func linkPath(link string) string {
	if strings.HasPrefix(link, "mailto:") {
		return ""
	}
	if i := strings.Index(link, "://"); i > 0 && !strings.ContainsAny(link[:i], `/\`) {
		return ""
	}
	return pathWithoutPlace(sys.ExpandPath(link))
}

// pathWithoutPlace is path with the ":line" or ":line:col" the terminal
// appends to a file it finds taken off. The colon after a drive letter is no
// place, or a path on C: would lose its drive.
func pathWithoutPlace(path string) string {
	for {
		i := strings.LastIndexByte(path, ':')
		if i <= 1 || !allDigits(path[i+1:]) {
			return path
		}
		path = path[:i]
	}
}

// allDigits reports whether s is a run of decimal digits. An empty string is
// none: the place in a file is a place only when it has a number.
func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
