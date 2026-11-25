package store

import (
	"strings"

	"wtgo/internal/action"
)

// Boundary: what a window starts with, before any file -- every field's
// default, the default keys, and the font family a terminal falls back to.

// EmptyFontFamily is the generic name every system resolves to a monospace
// font: the last name of the default list, and the one the font field's menu
// offers on its own for a person who wants fontconfig to choose.
const EmptyFontFamily = "monospace"

// FillEmptyFont names the default families where the family is empty, which is
// what a terminal draws with then: a family the file left out reads as the
// default, so a family the file emptied reads as it too. The settings page asks
// for it when a field is submitted, and the window saves nothing without it.
func (s *Settings) FillEmptyFont() {
	if strings.TrimSpace(s.Font.Family) == "" {
		s.Font.Family = DefaultFontFamily()
	}
}

// FontNames are the names a family list holds, split as the text engine splits
// them: on commas, with the spaces around a name and the quotes it may wear
// taken off, and with the empty ones dropped. The terminal reads the list in
// order and draws with the first name it can, so the order is what the list is
// for, and the window walks it the same way.
func FontNames(list string) []string {
	var out []string
	for _, f := range strings.Split(list, ",") {
		f = strings.TrimSpace(f)
		if len(f) > 1 {
			if q := f[:1]; (q == `"` || q == `'`) && f[len(f)-1:] == q {
				f = strings.TrimSpace(f[1 : len(f)-1])
			}
		}
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

// languageNames names the language settings of versions before the catalogs,
// which said a region: the tags the catalogs are filed under. A file that
// holds one keeps its language instead of falling back to the system's.
var languageNames = map[string]string{"zh-CN": "zh-Hans", "en-US": "en"}

// Default is the settings a window starts with, before any file.
func Default() Settings {
	return Settings{
		Language:          "",
		Theme:             "dark",
		StartOn:           "default",
		Font:              Font{Family: DefaultFontFamily(), Weight: 500, Size: 16, LineHeight: 1.2},
		ColorScheme:       "Campbell",
		Padding:           Padding{X: 12, Y: 8},
		BackgroundOpacity: 1.0,
		CursorShape:       "bar",
		CursorBlink:       true,
		ScrollbackMB:      10,
		PaddingColor:      "background",
		MinimumContrast:   0,
		ScrollMultiplier:  1,
		ClipboardRead:     "deny",
		ClipboardWrite:    "allow",
		Links:             Links{Program: true, Path: true},
		RightClick:        "copyOrPaste",
		// A right click that pastes is what a Windows Terminal user already
		// does, and a program that takes the mouse (a TUI, a shell plugin)
		// would otherwise take the click away with the person never knowing
		// Shift gives it back. Those who want that program's menu turn it off.
		RightClickCopyPastePriority: true,
		PasteMultilineConfirm:       true,
		ConfirmClose:                ConfirmRunning,
		CloseOnLastTab:              true,
		ConfirmCloseWindow:          true,
		NewTabPosition:              "last",
		TabWidth:                    "equal",
		MaxPanes:                    4,
		Keybindings:                 DefaultKeybindings(),
		// A list the file may leave out is still a list: an empty one writes
		// itself as [] rather than null, which is what a person reading the file
		// expects of a setting that holds many things.
		Schemes:  []Scheme{},
		Profiles: []Profile{},
	}
}

// DefaultKeybindings are the keys the actions start on, as the action catalog
// names them.
func DefaultKeybindings() []Keybinding {
	var kb []Keybinding
	for _, a := range action.All() {
		for _, key := range a.Keys {
			kb = append(kb, Keybinding{Key: key, Action: a.ID})
		}
	}
	return kb
}
