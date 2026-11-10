package ui

import (
	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/i18n"
	"wtgo/internal/store"
)

// actionsPage lists the keybindings, and records the keys of the rows that ask
// for one.
func actionsPage(c *myui.Context, st Settings, act Actions) {
	s := st.Draft
	section(c, i18n.T("Actions"), func() {
		myui.Column(c).Gap(2).Children(func() {
			for i, kb := range s.Keybindings {
				keybindingRow(c, st, act, i, kb, s.Keybindings)
			}
		})
		myui.Button(c, i18n.T("Restore the defaults")).OnClick(act.ResetKeybindings)
	})
	if st.Recording >= 0 && st.Recording < len(s.Keybindings) {
		capture(c, act, st.Recording)
	}
}

// keybindingRow is one keybinding: the key that runs it, which a click changes,
// and the action it runs.
func keybindingRow(c *myui.Context, st Settings, act Actions, i int, kb store.Keybinding, all []store.Keybinding) {
	t := c.Theme()
	key := kb.Key
	switch {
	case st.Recording == i:
		key = i18n.T("Press a key…")
	case key == "":
		key = i18n.T("None")
	}
	color := t.Text
	if clashes(all, kb.Key) {
		color = t.Danger
	}
	myui.Row(c).Height(30).AlignItems(myui.Center).Gap(8).Children(func() {
		myui.Box(c).Width(160).Height(26).Padding(0, 8).Justify(myui.Center).Radius(4).
			Background(t.Surface).Cursor(myui.CursorPointer).
			OnClick(func() { act.RecordKeybinding(i) }).Children(func() {
			myui.Text(c, key).FontSize(12).TextColor(color).SingleLine()
		})
		myui.Text(c, kb.Action).Grow(1).FontSize(12).TextColor(t.TextMuted).SingleLine()
		myui.Button(c, i18n.T("Remove")).OnClick(func() { act.RemoveKeybinding(i) })
	})
}

// clashes reports whether another keybinding runs on the same key.
func clashes(all []store.Keybinding, key string) bool {
	if key == "" {
		return false
	}
	n := 0
	for _, kb := range all {
		if kb.Key == key {
			n++
		}
	}
	return n > 1
}

// capture listens for the key the row being recorded takes. The window hears
// keys as shortcuts only, so it listens for every key there is, with every
// combination of the modifiers it knows.
func capture(c *myui.Context, act Actions, row int) {
	mods := []myui.Modifiers{
		0,
		myui.Ctrl,
		myui.Shift,
		myui.Alt,
		myui.Ctrl | myui.Shift,
		myui.Ctrl | myui.Alt,
		myui.Shift | myui.Alt,
		myui.Ctrl | myui.Shift | myui.Alt,
	}
	for _, m := range mods {
		for _, k := range KeyNames() {
			name, ok := FormatKey(m, k.Key)
			if !ok {
				continue
			}
			c.OnWindowShortcut(m, k.Key, func() { act.SetKeybinding(row, name) })
		}
	}
}
