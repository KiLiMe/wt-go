package ui

import (
	myui "github.com/egoist/mygo/ui"
)

// confirm is the dialog that asks before the window does something it cannot
// take back. Dismissing it is the same as its no.
func confirm(c *myui.Context, st State, act Actions) {
	if st.Confirm == nil {
		return
	}
	q := st.Confirm
	myui.Modal(c, &q.Open, func() {
		myui.Column(c).
			Width(320).
			Gap(12).
			Children(func() {
				myui.Text(c, q.Title).FontWeight(600)
				myui.Text(c, q.Text).FontSize(12).TextColor(c.Theme().TextMuted)
				myui.Row(c).
					Gap(8).
					Justify(myui.End).
					Children(func() {
						myui.Button(c, q.No).OnClick(act.ConfirmNo)
						myui.PrimaryButton(c, q.Yes).OnClick(act.ConfirmYes)
					})
			})
	})
}
