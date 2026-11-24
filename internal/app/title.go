package app

import (
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"wtgo/internal/pane"
	"wtgo/locales"
)

// title is the window's title: the name of the tab in front.
func (a *App) title() string {
	if tab, ok := a.panes.Front(); ok {
		return a.tabName(tab)
	}
	return "wt-go"
}

// tabName is what the strip calls a tab. A tab whose program ended says how
// it ended.
func (a *App) tabName(tab *pane.Tab) string {
	if tab.Focus == nil {
		return tab.Name
	}
	n := tab.Focus
	if !a.settings.TitleFromActiveTerminal {
		n = tab.First()
		// A tab with no pane in its tree has nothing to take a title from;
		// it still has a name, which is what the strip should show.
		if n == nil {
			return label(tab.Name, "", tab.ProfileName, tab.Runs)
		}
	}
	term := n.Term
	if term == nil {
		return label(tab.Name, "", tab.ProfileName, tab.Runs)
	}
	name := label(tab.Name, term.Title(), tab.ProfileName, tab.Runs)
	select {
	case <-term.Done():
		return name + " - " + locales.TabExited() + " (" + strconv.Itoa(term.ExitCode()) + ")"
	default:
		return name
	}
}

// label is what a tab is called: the name it was given, else the title its
// program set, unless that title is the path of the program, which is what a
// shell sets when it has no title of its own; else the name of its profile;
// else the name of the program itself, because a tab that runs no named
// profile runs whatever shell the machine has, and a tab with no name at all
// cannot be told from the empty strip beside it; else the shell it started,
// for the same reason while that shell has set no title yet. Only a machine
// with no shell at all falls through to the word for following the device.
func label(name, title, profileName, runs string) string {
	switch {
	case name != "":
		return name
	case title != "" && !isProgramPath(title):
		return title
	case profileName != "":
		return profileName
	case programName(title) != "":
		return programName(title)
	case runs != "":
		return runs
	default:
		return locales.SystemProfile()
	}
}

// programName is the name of the program a path points at, without the .exe
// that says it is one: a title that is a path still names its shell, and
// "pwsh" says more than nothing does.
func programName(p string) string {
	if p == "" {
		return ""
	}
	// A shell's title can carry the other platform's separator (a remote or
	// WSL session), so split on both instead of on the host's own.
	base := path.Base(strings.ReplaceAll(p, `\`, "/"))
	for _, ext := range []string{".exe", ".EXE"} {
		base = strings.TrimSuffix(base, ext)
	}
	if base == "." || base == "/" {
		return ""
	}
	return base
}

// isProgramPath reports whether s is the path of a program rather than a
// title, which a shell with no title of its own sets.
func isProgramPath(s string) bool {
	if filepath.VolumeName(s) != "" {
		return true // D:\Applications\pwsh\pwsh.exe, \\server\share\pwsh.exe
	}
	if strings.HasPrefix(s, `\\`) || strings.HasPrefix(s, "/") {
		return true
	}
	return strings.HasSuffix(strings.ToLower(s), ".exe")
}

// syncTitle gives the window the title of the tab in front, when it changed.
func (a *App) syncTitle() {
	if a.win == nil || a.closed {
		return
	}
	if title := a.title(); title != a.shownTitle {
		a.shownTitle = title
		a.win.SetTitle(title)
	}
}
