package ui

import (
	"sort"
	"strings"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/action"
	"wtgo/internal/store"
	"wtgo/locales"
)

// Boundary: the actions page as it is drawn -- groups, rows, the toolbar and
// the filter. What a row says about the file is derived in keybindings_model.go;
// recording a key is keybindings_capture.go.

// actionsPage lists a row for every action the window knows, by the group
// the catalog puts it in, and one for the keybindings of actions it does not
// know; it records the key of the row that asks for one. The toolbar narrows
// the list, and the numbered actions past the third stay folded away until
// the page asks for them, so that the list stays short without hiding the
// groups it belongs to.
func actionsPage(c *myui.Context, st Settings, act Actions) {
	s := st.Draft
	ks := pageKeys(st.KeysCache, s.Keybindings, st)
	query := strings.ToLower(text(st.ActionFilter))
	actionsToolbar(c, act, len(ks.changed), len(ks.clashing) > 0, st.OnlyChanged, st.ActionFilter, st.Reveal)
	listed := 0
	for _, g := range ks.groups {
		rows, folded, more := listedRows(g.rows, ks, st, query)
		if len(rows) == 0 && folded == 0 {
			continue
		}
		listed++
		section(c, action.GroupTitle(g.title), func() {
			for _, e := range rows {
				keybindingRow(c, st, act, e, ks)
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
			btn := myui.Button(c, locales.RestoreDefaults()).OnClick(act.ResetKeybindings)
			mark(c, btn, keyRestoreDefaults, st.Reveal)
		})
	})
	if st.Recording != "" {
		capture(c, st.Capture, act, st.Recording)
	}
}

// actionsToolbar is what narrows the list: the words an action's name, one of
// its keys or its id match, the switch that keeps only the actions the file
// binds on other keys than the catalog's, and the word that says some key runs
// more than one action, which the page cannot settle for a person.
func actionsToolbar(c *myui.Context, act Actions, changed int, conflict, onlyChanged bool, filter *string, reveal *Reveal) {
	t := c.Theme()
	myui.Row(c).Gap(8).AlignItems(myui.Center).Padding(0, 0, 8, 0).Children(func() {
		if filter != nil {
			input := myui.TextInput(c, filter).Placeholder(locales.KeySearch()).Grow(1)
			mark(c, input, keySearchActions, reveal)
		}
		btn := myui.Button(c, locales.OnlyChanged(changed)).FontSize(11)
		if onlyChanged {
			btn.Background(t.Accent).TextColor(t.AccentText)
		}
		btn.OnClick(func() { act.ToggleChangedKeys() })
		mark(c, btn, keyOnlyChanged, reveal)
		if conflict {
			myui.Text(c, locales.KeysConflict()).FontSize(11).TextColor(t.Danger).SingleLine()
		}
	})
}

// keybindingRow is one row of the page: the keys that run the action, which a
// click changes, the action it runs, and the two buttons that take those keys
// away or put the catalog's back. Both buttons keep their room while they have
// nothing to do, so that a row never shifts under the pointer as keys change.
func keybindingRow(c *myui.Context, st Settings, act Actions, e entry, ks *PageKeybindings) {
	t := c.Theme()
	isChanged := ks.changed[e.action]
	keys := ks.keys(e.action)
	shown := strings.Join(keys, " · ")
	color := t.Text
	for _, key := range keys {
		if ks.clashing[key] > 1 {
			color = t.Danger
		}
	}
	switch {
	case st.Recording == e.action:
		shown = locales.PressKey()
	case len(keys) == 0:
		shown = locales.None()
	}
	myui.Row(c).Padding(8, 16).AlignItems(myui.Center).Gap(8).Children(func() {
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
func listedRows(rows []entry, ks *PageKeybindings, st Settings, query string) (out []entry, folded, more int) {
	folding := !st.KeysExpanded && query == "" && !st.OnlyChanged
	for _, e := range rows {
		if st.OnlyChanged && !ks.changed[e.action] {
			continue
		}
		if query != "" && !matchesQuery(e, ks, query) {
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
func matchesQuery(e entry, ks *PageKeybindings, query string) bool {
	if strings.Contains(strings.ToLower(e.action), query) ||
		strings.Contains(strings.ToLower(action.Title(e.action)), query) {
		return true
	}
	for _, key := range ks.keys(e.action) {
		if strings.Contains(strings.ToLower(key), query) {
			return true
		}
	}
	return false
}
