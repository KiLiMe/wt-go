package ui

import (
	"sort"

	"wtgo/internal/action"
	"wtgo/internal/store"
)

// Boundary: what the actions page derives from the file -- every action's
// keys, which of them differ from the defaults, which keys run two actions —
// and the comparisons that keep the page from rebuilding when nothing changed.
// Nothing here draws.

// PageKeybindings is what the actions page derives from the keybindings of the
// file: the groups it lists, the actions whose keys are not the catalog's and
// the keys more than one action runs. Every row asks for the keys of its
// action and for both of those, so a page of hundreds of rows built them anew
// every frame. The page keeps one from frame to frame, in Settings.KeysCache.
type PageKeybindings struct {
	// _ keeps `==` from comparing two caches equal, which would say the
	// settings page holds the one and not the other.
	_ [0]func()
	// all are the keybindings the rest was derived from, kept as the page was
	// handed them: the settings page holds the slice, and it is the same one
	// until the file changes.
	all      []store.Keybinding
	only     bool
	expanded bool
	groups   []actionGroup
	changed  map[string]bool
	clashing map[string]int
	// of holds the keys of each action the page has asked about: a page asks
	// about the same actions frame after frame.
	of map[string][]string
}

// pageKeys is what the page derives from the keybindings of the file, built
// again only as the file or the switches that show what it holds change. The
// page keeps the answer from frame to frame.
func pageKeys(k *PageKeybindings, all []store.Keybinding, st Settings) *PageKeybindings {
	// A page drawn without a window to keep its answer keeps it for the frame
	// alone: the next frame derives it again.
	if k == nil {
		return newPageKeybindings(all, st)
	}
	// The list is the one the page was handed before: the settings page holds
	// it, and builds a new one only as the file changes, so the same slice
	// holding the same keys is the file unchanged.
	if k.groups != nil && k.only == st.OnlyChanged && k.expanded == st.KeysExpanded &&
		sameKeybindings(k.all, all) {
		return k
	}
	*k = *newPageKeybindings(all, st)
	return k
}

// newPageKeybindings is what the page derives from the keybindings of the file.
func newPageKeybindings(all []store.Keybinding, st Settings) *PageKeybindings {
	return &PageKeybindings{
		all:      all,
		only:     st.OnlyChanged,
		expanded: st.KeysExpanded,
		groups:   groups(all),
		changed:  changedActions(all),
		clashing: clashingKeys(all),
		of:       map[string][]string{},
	}
}

// keys are the keys the file binds an action on, in the order it lists them.
func (k *PageKeybindings) keys(id string) []string {
	if keys, ok := k.of[id]; ok {
		return keys
	}
	// An action with no key is stored as an empty list, not nil: a nil slice
	// would be looked up again, and built again, every frame.
	keys := keysOf(k.all, id)
	if keys == nil {
		keys = []string{}
	}
	k.of[id] = keys
	return keys
}

// sameKeybindings reports whether two lists bind the same keys to the same
// actions, in the same order: the length, then what they hold, since the page
// hands the cache the list it already kept.
func sameKeybindings(a, b []store.Keybinding) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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
