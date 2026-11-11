package ui

import (
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/i18n"
)

// The kinds of question the advanced page asks, which is what the window does
// with the answer.
const (
	AskRediscover = "rediscover"
	AskReset      = "reset"
)

// advancedPage is what a person rarely needs: probing the machine again, and
// starting over from the defaults. Both change the file rather than the draft,
// so the page has no fields to save.
func advancedPage(c *myui.Context, st Settings, act Actions) {
	t := c.Theme()
	section(c, i18n.T("Shells"), func() {
		field(c, i18n.T("Probe the shells again"),
			i18n.T("The window probes this machine when it writes its first settings file. "+
				"Probing again gives every built-in profile the command that is installed now, "+
				"hides the ones whose program is gone and shows the ones that are back. "+
				"Names, colors and directories stay."),
			func() {
				myui.Button(c, i18n.T("Probe")).OnClick(func() { act.AskAdvanced(AskRediscover) })
			})
	})
	section(c, i18n.T("Reset"), func() {
		field(c, i18n.T("Delete settings.json"),
			i18n.T("Deletes the file and starts the window again, which writes the defaults "+
				"and probes the machine the way the first start did."),
			func() {
				myui.Button(c, i18n.T("Delete and restart")).
					Background(t.Danger).TextColor(myui.Hex("#ffffff")).
					OnClick(func() { act.AskAdvanced(AskReset) })
			})
	})
	section(c, i18n.T("Folders"), func() {
		field(c, i18n.T("Settings folder"),
			i18n.T("Where settings.json lives. The file itself needs a program the desktop "+
				"edits JSON with, which not every Linux desktop has, so the window opens the folder."),
			func() {
				myui.Button(c, i18n.T("Open")).OnClick(act.OpenSettingsFolder)
			})
		field(c, i18n.T("Program folder"),
			i18n.T("Where the program runs from."),
			func() {
				myui.Button(c, i18n.T("Open")).OnClick(act.OpenProgramFolder)
			})
	})
}
