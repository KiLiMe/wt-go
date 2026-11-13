package ui

import (
	"sort"
	"strings"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/action"
	"wtgo/internal/i18n"
	"wtgo/internal/store"
	"wtgo/locales"
)

// actionsPage lists a row for every action the window knows, by the group
// the catalog puts it in, and one for the keybindings of actions it does not
// know; it records the key of the row that asks for one. The toolbar narrows
// the list, and the numbered actions past the third stay folded away until
// the page asks for them, so that the list stays short without hiding the
// groups it belongs to.
func actionsPage(c *myui.Context, st Settings, act Actions) {
	s := st.Draft
	changed := changedActions(s.Keybindings)
	clashing := clashingKeys(s.Keybindings)
	query := strings.ToLower(searchText(st.ActionFilter))
	actionsToolbar(c, act, len(changed), len(clashing) > 0, st.OnlyChanged, st.ActionFilter)
	listed := 0
	for _, g := range groups(s.Keybindings) {
		rows, folded, more := listedRows(g.rows, s.Keybindings, changed, st, query)
		if len(rows) == 0 && folded == 0 {
			continue
		}
		listed++
		section(c, i18n.T(g.title), func() {
			for _, e := range rows {
				keybindingRow(c, st, act, e, s.Keybindings, changed[e.action], clashing)
			}
			switch {
			case folded > 0:
				row(c, func() {
					myui.Button(c, locales.ShowNumbered(folded)).FontSize(11).
						OnClick(func() { act.ToggleNumberedKeys() })
				})
			case more > 0:
				row(c, func() {
					myui.Button(c, locales.HideNumbered()).FontSize(11).
						OnClick(func() { act.ToggleNumberedKeys() })
				})
			}
		})
	}
	if listed == 0 {
		row(c, func() {
			myui.Text(c, locales.NoActions()).FontSize(12).TextColor(c.Theme().TextMuted)
		})
	}
	card(c, cardFace(c.Theme()), cardLine(c.Theme()), func() {
		row(c, func() {
			myui.Button(c, locales.RestoreDefaults()).OnClick(act.ResetKeybindings)
		})
	})
	if st.Recording != "" {
		capture(c, act, st.Recording)
	}
}

// actionsToolbar is what narrows the list: the words an action's name, one of
// its keys or its id match, the switch that keeps only the actions the file
// binds on other keys than the catalog's, and the word that says some key runs
// more than one action, which the page cannot settle for a person.
func actionsToolbar(c *myui.Context, act Actions, changed int, conflict, onlyChanged bool, filter *string) {
	t := c.Theme()
	myui.Row(c).Gap(8).AlignItems(myui.Center).Padding(0, 0, 8, 0).Children(func() {
		if filter != nil {
			myui.TextInput(c, filter).Placeholder(locales.KeySearch()).Grow(1)
		}
		btn := myui.Button(c, locales.OnlyChanged(changed)).FontSize(11)
		if onlyChanged {
			btn.Background(t.Accent).TextColor(t.AccentText)
		}
		btn.OnClick(func() { act.ToggleChangedKeys() })
		if conflict {
			myui.Text(c, locales.KeysConflict()).FontSize(11).TextColor(t.Danger).SingleLine()
		}
	})
}

// searchText is what the actions page searches for, "" for nothing.
func searchText(filter *string) string {
	if filter == nil {
		return ""
	}
	return *filter
}

// keybindingRow is one row of the page: the keys that run the action, which a
// click changes, the action it runs, and the two buttons that take those keys
// away or put the catalog's back. Both buttons keep their room while they have
// nothing to do, so that a row never shifts under the pointer as keys change.
func keybindingRow(c *myui.Context, st Settings, act Actions, e entry, all []store.Keybinding, isChanged bool, clashing map[string]int) {
	t := c.Theme()
	keys := keysOf(all, e.action)
	shown := strings.Join(keys, " · ")
	color := t.Text
	for _, key := range keys {
		if clashing[key] > 1 {
			color = t.Danger
		}
	}
	switch {
	case st.Recording == e.action:
		shown = locales.PressKey()
	case len(keys) == 0:
		shown = locales.None()
	}
	myui.Row(c).Padding(8, 16).AlignItems(myui.Center).Gap(8).
		BorderWidth(0, 0, 1, 0).BorderColor(cardLine(t)).Children(func() {
		cell := myui.Box(c).Width(160).Height(26).Padding(0, 8).Justify(myui.Center).Radius(4).
			Background(t.Surface).Cursor(myui.CursorPointer).
			OnClick(func() { act.RecordKeybinding(e.action) })
		if len(keys) > 1 {
			// The cell is narrow enough that two keys may not both fit; the
			// hint says what the ellipsis cut.
			cell = cell.Tooltip(strings.Join(keys, ", "))
		}
		cell.Children(func() {
			myui.Text(c, shown).FontSize(12).TextColor(color).SingleLine().Ellipsis("…")
		})
		myui.Text(c, action.Title(e.action)).Grow(1).FontSize(12).TextColor(t.TextMuted).SingleLine().Tooltip(e.action)
		if isChanged {
			myui.Text(c, locales.Changed()).FontSize(10).TextColor(t.Warning).
				Background(t.Warning.Alpha(0.16)).Padding(0, 6).Radius(4).SingleLine()
		}
		reset := myui.Button(c, locales.ResetKey()).FontSize(11).Disabled(!isChanged)
		if a, ok := action.Lookup(e.action); ok && len(a.Keys) > 0 && isChanged {
			reset = reset.Tooltip(locales.DefaultKeys(strings.Join(a.Keys, ", ")))
		}
		reset.OnClick(func() { act.ResetKeybinding(e.action) })
		myui.Button(c, locales.Remove()).FontSize(11).Disabled(len(keys) == 0).
			OnClick(func() { act.RemoveKeybinding(e.action) })
	})
}

// actionGroup is one group of actions, and the rows it holds, in the order the
// catalog lists them.
type actionGroup struct {
	title string
	rows  []entry
}

// entry is one row of the page: an action, whatever the file binds it on.
type entry struct {
	action string
}

// groupOrder is where the catalog puts each group, for the settings page to
// list them in that order.
var groupOrder = func() map[string]int {
	at := map[string]int{}
	for _, a := range action.All() {
		if _, ok := at[a.GroupKey]; !ok {
			at[a.GroupKey] = len(at)
		}
	}
	return at
}()

// groups is every action of the catalog by the group it is in, in the order
// the catalog lists the groups, and the keybindings of actions the window does
// not know last, under Other. The group it calls Other is the catalog's word,
// which the page translates as it draws.
func groups(all []store.Keybinding) []actionGroup {
	listed := map[string]bool{}
	rows := map[string][]entry{}
	var titles []string
	at := func(group string) {
		if _, ok := rows[group]; !ok {
			titles = append(titles, group)
		}
	}
	for _, a := range action.All() {
		at(a.GroupKey)
		listed[a.ID] = true
		rows[a.GroupKey] = append(rows[a.GroupKey], entry{action: a.ID})
	}
	for _, kb := range all {
		if action.Group(kb.Action) != "" || listed[kb.Action] {
			continue
		}
		at(action.GroupOther)
		listed[kb.Action] = true
		rows[action.GroupOther] = append(rows[action.GroupOther], entry{action: kb.Action})
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

// numberedPast is the number after which the page folds the numbered actions
// away: the first three stay, so that the keys a person presses all day show
// and the ones they look up now and then do not.
const numberedPast = 3

// listedRows are the rows of a group the page draws, how many numbered ones it
// folds away, and how many it would fold away were they folded: the page draws
// the bar that unfolds them while there are any, and the one that folds them
// back while they show.
func listedRows(rows []entry, all []store.Keybinding, changed map[string]bool, st Settings, query string) (out []entry, folded, more int) {
	folding := !st.KeysExpanded && query == "" && !st.OnlyChanged
	for _, e := range rows {
		if st.OnlyChanged && !changed[e.action] {
			continue
		}
		if query != "" && !matchesQuery(e, all, query) {
			continue
		}
		if n, ok := action.Number(e.action); ok && n > numberedPast {
			more++
			if folding {
				folded++
				continue
			}
		}
		out = append(out, e)
	}
	return out, folded, more
}

// matchesQuery reports whether a row is one the page searches for: its name,
// its id, or a key the file binds it on, in any case.
func matchesQuery(e entry, all []store.Keybinding, query string) bool {
	if strings.Contains(strings.ToLower(e.action), query) ||
		strings.Contains(strings.ToLower(action.Title(e.action)), query) {
		return true
	}
	for _, key := range keysOf(all, e.action) {
		if strings.Contains(strings.ToLower(key), query) {
			return true
		}
	}
	return false
}

// keysOf are the keys the file binds an action on, in the order it lists them.
func keysOf(all []store.Keybinding, id string) []string {
	var out []string
	for _, kb := range all {
		if kb.Action == id {
			out = append(out, kb.Key)
		}
	}
	return out
}

// changedActions are the actions the file binds on other keys than the catalog
// starts them on, which the page marks and resets alone.
func changedActions(all []store.Keybinding) map[string]bool {
	have := map[string][]string{}
	for _, kb := range all {
		have[kb.Action] = append(have[kb.Action], kb.Key)
	}
	out := map[string]bool{}
	for _, a := range action.All() {
		if !sameKeys(a.Keys, have[a.ID]) {
			out[a.ID] = true
		}
	}
	for id := range have {
		// A keybinding of an action the window does not know has no default
		// to go back to, so it is always the person's own.
		if _, ok := action.Lookup(id); !ok {
			out[id] = true
		}
	}
	return out
}

// sameKeys reports whether two sets of keys are the same, in any order.
func sameKeys(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	x := append([]string(nil), a...)
	y := append([]string(nil), b...)
	sort.Strings(x)
	sort.Strings(y)
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

// clashingKeys are the keys of the file that run more than one action, and how
// many they run. A key the file binds once is not in the map: the page asks it
// whether there is a clash at all, and a key that runs one action is none.
func clashingKeys(all []store.Keybinding) map[string]int {
	seen := map[string]int{}
	out := map[string]int{}
	for _, kb := range all {
		if kb.Key == "" {
			continue
		}
		seen[kb.Key]++
		if seen[kb.Key] > 1 {
			out[kb.Key] = seen[kb.Key]
		}
	}
	return out
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
