package ui

import (
	myui "github.com/egoist/mygo/ui"

	"wtgo/locales"
)

// rename is the dialog that renames a tab.
func rename(c *myui.Context, st State, act Actions) {
	if st.Rename == nil {
		return
	}
	r := st.Rename
	myui.Modal(c, &r.Open, func() {
		myui.Column(c).
			Width(280).
			Gap(12).
			Children(func() {
				myui.Text(c, locales.RenameTab()).FontWeight(600)
				myui.TextInput(c, &r.Text).
					SingleLine().
					AutoFocus().
					OnSubmit(func() { act.RenameTab(r.Tab, r.Text) })
				myui.Row(c).
					Gap(8).
					Justify(myui.End).
					Children(func() {
						myui.Button(c, locales.Cancel()).OnClick(act.CancelRename)
						myui.PrimaryButton(c, locales.Ok()).
							OnClick(func() { act.RenameTab(r.Tab, r.Text) })
					})
			})
	})
}
