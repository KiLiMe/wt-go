package store

import (
	"reflect"
	"slices"
	"strings"
)

// Boundary: comparing and copying -- what tells a changed file from an
// unchanged one, field by field where the JSON would compare equal anyway.

// Equal reports whether two sets of settings are the same, which is how the
// window knows whether what it shows is what the file holds. The keys no field
// names are not settings: a file that differs only in them holds what the
// window shows.
func Equal(a, b Settings) bool {
	return reflect.DeepEqual(withoutExtra(a), withoutExtra(b))
}

// withoutExtra is s without the keys no field names, for the comparisons that
// ask what the settings say rather than what the file also holds.
func withoutExtra(s Settings) Settings {
	s.Extra = nil
	return s
}

// EqualProfile reports whether two profiles are the same, which is how the
// window knows whether the profile page changed the profile it opened.
func EqualProfile(a, b Profile) bool {
	return reflect.DeepEqual(a, b)
}

// EqualSchemes reports whether two lists write the same color schemes: the same
// names in the same order, with the same colors. The settings page keeps the
// list of schemes it built while this holds, since the built-in names and
// Ghostty's are the same for as long as the window runs.
func EqualSchemes(a, b []Scheme) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		x, y := a[i], b[i]
		if !strings.EqualFold(x.Name, y.Name) || x.Background != y.Background ||
			x.Foreground != y.Foreground || x.Cursor != y.Cursor ||
			x.Selection != y.Selection || len(x.Palette) != len(y.Palette) {
			return false
		}
		for j := range x.Palette {
			if x.Palette[j] != y.Palette[j] {
				return false
			}
		}
	}
	return true
}

// Clone is a copy of the settings that shares nothing with them, so that a
// form may edit it, and the file's copy stay as it was.
func (s Settings) Clone() Settings {
	out := s
	out.Schemes = slices.Clone(s.Schemes)
	for i := range out.Schemes {
		out.Schemes[i].Palette = slices.Clone(s.Schemes[i].Palette)
	}
	out.Profiles = slices.Clone(s.Profiles)
	for i := range out.Profiles {
		out.Profiles[i].Args = slices.Clone(s.Profiles[i].Args)
	}
	out.Keybindings = slices.Clone(s.Keybindings)
	// Extra is shared like the strings are: a form edits the settings, never
	// the map of keys it does not know.
	return out
}
