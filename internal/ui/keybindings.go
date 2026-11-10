package ui

import (
	"sort"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/action"
	"wtgo/internal/i18n"
	"wtgo/internal/store"
)

// actionsPage lists a row for every action the window knows, by the group
// the catalog puts it in, and one for the keybindings of actions it does not
// know; it records the key of the row that asks for one.
func actionsPage(c *myui.Context, st Settings, act Actions) {
	s := st.Draft
	section(c, i18n.T("Actions"), func() {
		for _, g := range groups(s.Keybindings) {
			myui.Text(c, i18n.T(g.title)).FontSize(12).TextColor(c.Theme().TextMuted)
			myui.Column(c).Gap(2).Children(func() {
				for _, e := range g.rows {
					keybindingRow(c, st, act, e, s.Keybindings)
				}
			})
		}
		myui.Button(c, i18n.T("Restore the defaults")).OnClick(act.ResetKeybindings)
	})
	if st.Recording != "" {
		capture(c, act, st.Recording)
	}
}

// actionGroup is one group of actions, and the rows it holds, in the order the
// catalog lists them.
type actionGroup struct {
	title string
	rows  []entry
}

// entry is one row of the page: an action, and the keybinding the file holds
// for it, -1 when it holds none.
type entry struct {
	action string
	row    int
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

// groups is every action of the catalog by the group it is in, in the order
// the catalog lists the groups, and the keybindings of actions the window does
// not know last, under Other. The group it calls Other is the catalog's word,
// which the page translates as it draws.
func groups(all []store.Keybinding) []actionGroup {
	bound := map[string]int{}
	for i, kb := range all {
		if _, ok := bound[kb.Action]; !ok {
			bound[kb.Action] = i
		}
	}
	rows := map[string][]entry{}
	var titles []string
	at := func(group string) {
		if _, ok := rows[group]; !ok {
			titles = append(titles, group)
		}
	}
	for _, a := range action.All() {
		at(a.Group)
		rows[a.Group] = append(rows[a.Group], entry{action: a.ID, row: take(bound, a.ID)})
	}
	for _, kb := range all {
		if action.Group(kb.Action) != "" {
			continue
		}
		at("Other")
		rows["Other"] = append(rows["Other"], entry{action: kb.Action, row: take(bound, kb.Action)})
	}
	sort.SliceStable(titles, func(i, j int) bool { return rank(titles[i]) < rank(titles[j]) })
	out := make([]actionGroup, 0, len(titles))
	for _, title := range titles {
		out = append(out, actionGroup{title: title, rows: rows[title]})
	}
	return out
}

// take is the row of an action the page has not listed yet, and -1 for the
// next one: an action binds one key, so its later keybindings, if the file has
// them, are left out.
func take(bound map[string]int, id string) int {
	i, ok := bound[id]
	if !ok {
		return -1
	}
	delete(bound, id)
	return i
}

// rank is where the catalog puts a group, after the ones it knows.
func rank(group string) int {
	if i, ok := groupOrder[group]; ok {
		return i
	}
	return len(groupOrder)
}

// keybindingRow is one row of the page: the key that runs the action, which a
// click changes, and the action it runs.
func keybindingRow(c *myui.Context, st Settings, act Actions, e entry, all []store.Keybinding) {
	t := c.Theme()
	key := ""
	if e.row >= 0 {
		key = all[e.row].Key
	}
	switch {
	case st.Recording == e.action:
		key = i18n.T("Press a key…")
	case key == "":
		key = i18n.T("None")
	}
	color := t.Text
	if clashes(all, key) {
		color = t.Danger
	}
	myui.Row(c).Height(30).AlignItems(myui.Center).Gap(8).Children(func() {
		myui.Box(c).Width(160).Height(26).Padding(0, 8).Justify(myui.Center).Radius(4).
			Background(t.Surface).Cursor(myui.CursorPointer).
			OnClick(func() { act.RecordKeybinding(e.action) }).Children(func() {
			myui.Text(c, key).FontSize(12).TextColor(color).SingleLine()
		})
		myui.Text(c, action.Title(e.action)).Grow(1).FontSize(12).TextColor(t.TextMuted).SingleLine().Tooltip(e.action)
		if e.row >= 0 {
			myui.Button(c, i18n.T("Remove")).OnClick(func() { act.RemoveKeybinding(e.action) })
		}
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

// capture listens for the key the action being recorded takes. The window
// hears keys as shortcuts only, so it listens for every key there is, with
// every combination of the modifiers it knows.
func capture(c *myui.Context, act Actions, action string) {
	for _, m := range captureMods() {
		for _, k := range KeyNames() {
			name, ok := FormatKey(m, k.Key)
			if !ok {
				continue
			}
			c.OnWindowShortcut(m, k.Key, func() { act.SetKeybinding(action, name) })
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
