package theme

import (
	"strings"
	"sync"

	"github.com/egoist/mygo/plugins/terminal"

	"wtgo/internal/store"
)

// Scheme is the scheme called name: one built in or one of Ghostty's, with
// the colors the settings write over it when they write a scheme of that
// name, and the colors of the window's theme for a name nothing knows.
func Scheme(s store.Settings, name string) (*terminal.Theme, bool) {
	if name == "" {
		return nil, false
	}
	t, known := named(name)
	sc, own := custom(s, name)
	if !known && !own {
		return nil, false
	}
	if t == nil {
		t = base(s)
	}
	if own {
		apply(t, sc)
	}
	return t, true
}

// Schemes are the names a setting may give a color scheme: the ones the file
// writes, the ones this program carries, and the ones Ghostty ships, in
// order, once each.
func Schemes(s store.Settings) []string {
	// The names Ghostty ships come from a directory read and a sort of several
	// hundred of them: asked for once here, not once for the room and once for
	// the list.
	ghostty := terminal.GhosttyThemes()
	seen := make(map[string]bool)
	out := make([]string, 0, len(s.Schemes)+len(ghostty)+8)
	add := func(name string) {
		if name == "" || seen[strings.ToLower(name)] {
			return
		}
		seen[strings.ToLower(name)] = true
		out = append(out, name)
	}
	for _, sc := range s.Schemes {
		add(sc.Name)
	}
	for _, name := range BuiltinSchemes() {
		add(name)
	}
	for _, name := range ghostty {
		add(name)
	}
	return out
}

// named is the scheme called name that the terminal ships: one built in, else
// one of Ghostty's. It is a copy, since a caller writes the colors of its own
// over it.
func named(name string) (*terminal.Theme, bool) {
	t, ok := shipped(name)
	if !ok {
		return nil, false
	}
	copied := *t
	return &copied, true
}

// shipped is the scheme called name that the terminal ships, read once per
// name: a theme of Ghostty's is a file read and parsed, one the terminal
// carries a scan of its six hundred names, and the appearance page asks for the
// same name on every frame it draws. A theme file changed while the window runs
// is read again at the next launch.
func shipped(name string) (*terminal.Theme, bool) {
	namedCache.Lock()
	defer namedCache.Unlock()
	if entry, ok := namedCache.byName[name]; ok {
		return entry.theme, entry.ok
	}
	t, ok := read(name)
	if namedCache.byName == nil || len(namedCache.byName) >= namedCacheMax {
		namedCache.byName = make(map[string]namedEntry, namedCacheMax)
	}
	namedCache.byName[name] = namedEntry{theme: t, ok: ok}
	return t, ok
}

// read is the scheme called name that the terminal ships, read from it.
func read(name string) (*terminal.Theme, bool) {
	if t := builtin(name); t != nil {
		return t, true
	}
	if t, err := terminal.GhosttyTheme(name); err == nil {
		return t, true
	}
	return nil, false
}

// namedCache keeps the shipped schemes the frames before this one read, for the
// frames about to read the same ones.
var namedCache struct {
	sync.Mutex
	byName map[string]namedEntry
}

// namedCacheMax bounds it: a person may name any of Ghostty's schemes, and the
// settings page visits few of them.
const namedCacheMax = 64

// namedEntry is what the cache keeps of a name: the scheme the terminal ships
// under it, or that it ships none.
type namedEntry struct {
	theme *terminal.Theme
	ok    bool
}

// custom is the scheme the settings write by the name, if they write one.
func custom(s store.Settings, name string) (store.Scheme, bool) {
	for _, sc := range s.Schemes {
		if strings.EqualFold(sc.Name, name) {
			return sc, true
		}
	}
	return store.Scheme{}, false
}

// apply writes the colors of sc over t, and keeps the ones sc does not write.
func apply(t *terminal.Theme, sc store.Scheme) {
	override(&t.Background, sc.Background)
	override(&t.Foreground, sc.Foreground)
	override(&t.Cursor, sc.Cursor)
	override(&t.Selection, sc.Selection)
	for i, s := range sc.Palette {
		if i == len(t.Palette) {
			break
		}
		override(&t.Palette[i], s)
	}
}

// Stored copies a terminal theme into a color scheme the settings hold, so
// that a scheme the window shows can be kept and edited.
func Stored(th *terminal.Theme) store.Scheme {
	palette := make([]string, len(th.Palette))
	for i, col := range th.Palette {
		palette[i] = Hex(col)
	}
	return store.Scheme{
		Background: Hex(th.Background),
		Foreground: Hex(th.Foreground),
		Cursor:     Hex(th.Cursor),
		Selection:  Hex(th.Selection),
		Palette:    palette,
	}
}
