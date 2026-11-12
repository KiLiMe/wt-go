package ui

import (
	myui "github.com/egoist/mygo/ui"

	"wtgo/locales"
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
	section(c, locales.Shells(), func() {
		field(c, locales.ProbeAgain(),
			locales.ProbeAgainNote(),
			func() {
				myui.Button(c, locales.Probe()).OnClick(func() { act.AskAdvanced(AskRediscover) })
			})
	})
	section(c, locales.Folders(), func() {
		field(c, locales.SettingsFolder(),
			locales.SettingsFolderNote(),
			func() {
				myui.Button(c, locales.Open()).OnClick(act.OpenSettingsFolder)
			})
		field(c, locales.ProgramFolder(),
			locales.ProgramFolderNote(),
			func() {
				myui.Button(c, locales.Open()).OnClick(act.OpenProgramFolder)
			})
	})
	// Last, and in the danger color: the row that starts over is the one a
	// person comes to by mistake.
	dangerSection(c, locales.Reset(), func() {
		field(c, locales.DeleteSettings(),
			locales.DeleteSettingsNote(),
			func() {
				myui.Button(c, locales.DeleteAndRestart()).
					Background(t.Danger).TextColor(myui.Hex("#ffffff")).
					OnClick(func() { act.AskAdvanced(AskReset) })
			})
	})
}
