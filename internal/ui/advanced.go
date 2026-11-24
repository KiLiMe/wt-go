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
		field(c, "probe", locales.ProbeAgain(),
			locales.ProbeAgainNote(),
			func() {
				myui.Button(c, locales.Probe()).OnClick(func() { act.AskAdvanced(AskRediscover) })
			}, st.Reveal)
	})
	section(c, locales.Folders(), func() {
		field(c, "settingsFolder", locales.SettingsFolder(),
			locales.SettingsFolderNote(),
			func() {
				myui.Button(c, locales.Open()).OnClick(act.OpenSettingsFolder)
			}, st.Reveal)
		field(c, "programFolder", locales.ProgramFolder(),
			locales.ProgramFolderNote(),
			func() {
				myui.Button(c, locales.Open()).OnClick(act.OpenProgramFolder)
			}, st.Reveal)
	})
	// Last, and in the danger color: the row that starts over is the one a
	// person comes to by mistake. Its button sits under the words rather than in
	// the column the page's other buttons make, so that a click meant for a
	// button that opens a folder does not land on the one that deletes the file.
	dangerSection(c, locales.Reset(), func() {
		fieldBlock(c, "reset", locales.DeleteSettings(),
			locales.DeleteSettingsNote(),
			func() {
				myui.Button(c, locales.DeleteAndRestart()).
					Background(t.Danger).TextColor(myui.Hex("#ffffff")).
					OnClick(func() { act.AskAdvanced(AskReset) })
			}, st.Reveal)
	})
}
