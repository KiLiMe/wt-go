package ui

import (
	"strings"

	myui "github.com/egoist/mygo/ui"

	"wtgo/locales"
)

// about is the dialog that says what the program is: the name, the build that
// answered, what it was built from and what it is licensed under, which is what
// a person looks for when a bug report asks which build they run.
func about(c *myui.Context, st State, act Actions) {
	a := st.About
	if a == nil {
		return
	}
	t := c.Theme()
	myui.Modal(c, &a.Open, func() {
		myui.Column(c).Width(360).Gap(18).Children(func() {
			myui.Column(c).Gap(3).Children(func() {
				myui.Text(c, "wt-go").FontSize(20).FontWeight(600)
				myui.Text(c, a.Version).FontSize(12).TextColor(t.TextMuted).SingleLine()
			})
			myui.Column(c).Gap(9).Children(func() {
				if a.Revision != "" {
					aboutLine(c, locales.AboutCommit(), a.Revision)
				}
				if a.Mygo != "" {
					aboutLine(c, locales.AboutMygo(), a.Mygo)
				}
				aboutLine(c, locales.AboutRuntime(), a.Go+" "+a.Platform)
				mode := locales.AboutModeInstalled()
				if a.Portable {
					mode = locales.AboutModePortable()
				}
				aboutLine(c, locales.AboutMode(), mode)
				aboutTerminal(c, a.Terminal, a.TerminalPath)
				aboutLine(c, locales.License(), "MIT")
				aboutLine(c, locales.ProjectPage(), projectShown(a.Project))
			})
			// The files are long, and a path cut off at the dialog's edge
			// says nothing: each one gets a line of its own, the value under
			// its name, wrapping as far as it needs.
			if a.Settings != "" || a.Log != "" {
				myui.Column(c).Gap(9).Children(func() {
					if a.Settings != "" {
						aboutPath(c, locales.AboutSettings(), a.Settings)
					}
					if a.Log != "" {
						aboutPath(c, locales.AboutLog(), a.Log)
					}
				})
			}
			if a.Note != "" {
				myui.Text(c, a.Note).FontSize(11).TextColor(t.Danger)
			}
			myui.Row(c).Gap(8).Justify(myui.End).Children(func() {
				myui.Button(c, locales.ProjectPage()).OnClick(act.OpenProjectPage)
				myui.PrimaryButton(c, locales.Close()).OnClick(act.CloseAbout)
			})
		})
	})
}

// aboutLine is one row of the dialog: what it is called on the left, and what
// it says on the right.
func aboutLine(c *myui.Context, label, value string) {
	myui.Row(c).Justify(myui.SpaceBetween).Gap(16).Children(func() {
		myui.Text(c, label).FontSize(12).TextColor(c.Theme().TextMuted)
		myui.Text(c, value).FontSize(12).SingleLine()
	})
}

// aboutPath is a line of the dialog that holds a file's path: the name, and the
// path under it, wrapped rather than cut, since a path is only useful whole.
func aboutPath(c *myui.Context, label, path string) {
	myui.Column(c).Gap(2).Children(func() {
		myui.Text(c, label).FontSize(12).TextColor(c.Theme().TextMuted)
		pathText(c, path)
	})
}

// aboutTerminal is the line that says where the terminals' console comes from,
// with the dll under it when there is one: a report about colors or the mouse
// starts by asking which console it is.
func aboutTerminal(c *myui.Context, kind, path string) {
	myui.Column(c).Gap(2).Children(func() {
		aboutLine(c, locales.AboutTerminal(), terminalName(kind))
		if path != "" {
			pathText(c, path)
		}
	})
}

// terminalName is where the terminals' console comes from, in the words the
// dialog shows.
func terminalName(kind string) string {
	switch kind {
	case TerminalApp:
		return locales.AboutTerminalApp()
	case TerminalEnv:
		return locales.AboutTerminalEnv()
	}
	return locales.AboutTerminalInherent()
}

// pathText is a path in the dialog, smaller than the text around it and wrapped
// rather than cut: it is long, and half of one says nothing.
func pathText(c *myui.Context, path string) {
	myui.Text(c, path).FontSize(11).Wrap()
}

// projectShown is the address without the scheme, which a line of a dialog
// says well enough and reads without one.
func projectShown(project string) string {
	for _, scheme := range []string{"https://", "http://"} {
		if rest, ok := strings.CutPrefix(project, scheme); ok {
			return rest
		}
	}
	return project
}
