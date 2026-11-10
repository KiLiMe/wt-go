package ui

import (
	"sort"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/action"
	"wtgo/internal/i18n"
	"wtgo/internal/store"
)

// actionsPage lists the keybindings by the group of the action they run, and
// records the keys of the rows that ask for one.
func actionsPage(c *myui.Context, st Settings, act Actions) {
	s := st.Draft
	section(c, i18n.T("Actions"), func() {
		for _, g := range groups(s.Keybindings) {
			myui.Text(c, i18n.T(g.title)).FontSize(12).TextColor(c.Theme().TextMuted)
			myui.Column(c).Gap(2).Children(func() {
				for _, i := range g.rows {
					keybindingRow(c, st, act, i, s.Keybindings[i], s.Keybindings)
				}
			})
		}
		myui.Button(c, i18n.T("Restore the defaults")).OnClick(act.ResetKeybindings)
	})
	if st.Recording >= 0 && st.Recording < len(s.Keybindings) {
		capture(c, act, st.Recording)
	}
}

// actionGroup is the keybindings of one group, and the rows they hold, in the
// order the file lists them.
type actionGroup struct {
	title string
	rows  []int
}

// groupOrder is where the catalog puts each group, for the settings page to
// list them in that order.
var groupOrder = func() map[string]int {
	at := map[string]int{}
	for _, a := range action.All() {
		if _, ok := at[a.Group]; !ok {
			at[a.Group] = len(at)
		}
	}
	return at
}()

// groups is the keybindings by the group of the action they run, in the order
// the catalog lists the groups, and the ones of actions the window does not
// know last, under Other. The group it calls Other is the catalog's word, which
// the page translates as it draws.
func groups(all []store.Keybinding) []actionGroup {
	rows := map[string][]int{}
	var titles []string
	for i, kb := range all {
		group := action.Group(kb.Action)
		if group == "" {
			group = "Other"
		}
		if _, ok := rows[group]; !ok {
			titles = append(titles, group)
		}
		rows[group] = append(rows[group], i)
	}
	sort.SliceStable(titles, func(i, j int) bool { return rank(titles[i]) < rank(titles[j]) })
	out := make([]actionGroup, 0, len(titles))
	for _, title := range titles {
		out = append(out, actionGroup{title: title, rows: rows[title]})
	}
	return out
}

// rank is where the catalog puts a group, after the ones it knows.
func rank(group string) int {
	if i, ok := groupOrder[group]; ok {
		return i
	}
	return len(groupOrder)
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
		myui.Text(c, action.Title(kb.Action)).Grow(1).FontSize(12).TextColor(t.TextMuted).SingleLine().Tooltip(kb.Action)
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
	for _, m := range captureMods() {
		for _, k := range KeyNames() {
			name, ok := FormatKey(m, k.Key)
			if !ok {
				continue
			}
			c.OnWindowShortcut(m, k.Key, func() { act.SetKeybinding(row, name) })
		}
	}
}

// captureMods are the modifiers capture listens for: every combination of the
// modifiers it records, with and without Super.
func captureMods() []myui.Modifiers {
	base := []myui.Modifiers{
		0,
		myui.Ctrl,
		myui.Shift,
		myui.Alt,
		myui.Ctrl | myui.Shift,
		myui.Ctrl | myui.Alt,
		myui.Shift | myui.Alt,
		myui.Ctrl | myui.Shift | myui.Alt,
	}
	mods := make([]myui.Modifiers, 0, len(base)*2)
	mods = append(mods, base...)
	for _, m := range base {
		mods = append(mods, m|myui.Super)
	}
	return mods
}
