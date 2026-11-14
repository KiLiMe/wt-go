// Package store is settings.json: the shape of the file, its defaults, and
// the way it turns into bytes and back. It is pure: the app reads and writes
// the file, store only knows what is in it.
package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"wtgo/internal/action"
)

// Confirm is when the window asks before it closes a tab that still runs
// something: "running" only when a program runs besides the shell, "always"
// whenever the tab lives, "never" without asking. A file written when the
// setting was true or false reads as "running" and "never".
type Confirm string

// The answers Confirm takes.
const (
	ConfirmRunning Confirm = "running"
	ConfirmAlways  Confirm = "always"
	ConfirmNever   Confirm = "never"
)

// Ask reports whether closing a tab asks first, given whether a program still
// runs in it.
func (c Confirm) Ask(running bool) bool {
	switch c {
	case ConfirmAlways:
		return true
	case ConfirmNever:
		return false
	}
	return running
}

// Valid reports whether c is one of the answers a file may write.
func (c Confirm) Valid() bool {
	switch c {
	case ConfirmRunning, ConfirmAlways, ConfirmNever:
		return true
	}
	return false
}

// UnmarshalJSON reads the answer, and the true or false older files wrote.
func (c *Confirm) UnmarshalJSON(data []byte) error {
	var on bool
	if err := json.Unmarshal(data, &on); err == nil {
		*c = ConfirmNever
		if on {
			*c = ConfirmRunning
		}
		return nil
	}
	var word string
	if err := json.Unmarshal(data, &word); err != nil {
		return err
	}
	*c = Confirm(word)
	return nil
}

// MarshalJSON writes the answer as the word the setting documents.
func (c Confirm) MarshalJSON() ([]byte, error) { return json.Marshal(string(c)) }

// What startOn takes: the default profile first, or a new tab of the machine's
// shell.
const (
	StartDefault = "default"
	StartNew     = "new"
)

// Settings is settings.json, as docs/使用说明.md describes it.
type Settings struct {
	Language       string `json:"language"`
	Theme          string `json:"theme"`
	DefaultProfile string `json:"defaultProfile"`
	StartOn        string `json:"startOn"`
	StartDirectory string `json:"startDirectory"`

	Font              Font    `json:"font"`
	ColorScheme       string  `json:"colorScheme"`
	Padding           Padding `json:"padding"`
	BackgroundOpacity float64 `json:"backgroundOpacity"`
	CursorShape       string  `json:"cursorShape"`
	CursorBlink       bool    `json:"cursorBlink"`
	ScrollbackMB      int     `json:"scrollbackMB"`

	// PaddingColor is what the room around a terminal's grid shows:
	// "background" is the window's, "extend" the cells nearest to it, so
	// that a program that paints its own background reaches the edges,
	// and "always" extends them whatever they hold.
	PaddingColor string `json:"paddingColor"`
	// MinimumContrast is the least contrast text has with the background
	// it is drawn on, as a WCAG ratio from 0 to 21: 0 leaves every cell as
	// the program drew it, 4.5 is the ratio the guidelines call readable.
	MinimumContrast float32 `json:"minimumContrast"`

	CopyOnSelect          bool    `json:"copyOnSelect"`
	RightClick            string  `json:"rightClick"`
	PasteMultilineConfirm bool    `json:"pasteMultilineConfirm"`
	ConfirmClose          Confirm `json:"confirmClose"`
	CloseOnLastTab        bool    `json:"closeOnLastTab"`
	// ConfirmCloseWindow is whether the window asks before it closes with the
	// tabs it holds.
	ConfirmCloseWindow bool `json:"confirmCloseWindow"`
	// ScrollMultiplier scales how far the wheel and the touchpad scroll:
	// 1 scrolls the way the wheel's own setting says.
	ScrollMultiplier float32 `json:"scrollMultiplier"`
	// ClipboardRead is what a program reading the clipboard gets:
	// "deny", "allow" or "ask". ClipboardWrite is whether a program may
	// set the clipboard: "allow" or "deny". Both are the policies of
	// Ghostty, whose default for a read is to ask; here a read is refused
	// until the user allows it, since the window has no way to ask a
	// question a program is waiting on.
	ClipboardRead  string `json:"clipboardRead"`
	ClipboardWrite string `json:"clipboardWrite"`

	// NewTabPosition is where a tab opens: "last" or "first". TabWidth is
	// how wide tabs are: "compact", to the name, or "equal", all alike.
	NewTabPosition string `json:"newTabPosition"`
	TabWidth       string `json:"tabWidth"`
	// AlwaysOnTop keeps the window over the others; TitleFromActiveTerminal
	// lets the window's title follow the title of the terminal in front.
	AlwaysOnTop             bool `json:"alwaysOnTop"`
	TitleFromActiveTerminal bool `json:"titleFromActiveTerminal"`

	MaxPanes int `json:"maxPanes"`

	Schemes     []Scheme     `json:"schemes"`
	Profiles    []Profile    `json:"profiles"`
	Keybindings []Keybinding `json:"keybindings"`

	// Extra is the keys the file holds that no field above names, kept as
	// they were written so that saving from the settings page does not drop
	// a hand-written key or one a later version of the file knows. It is not
	// part of the settings themselves: Equal, Clone and Validate leave it be.
	Extra map[string]json.RawMessage `json:"-"`
}

// Font is the font of a terminal's text.
type Font struct {
	Family     string  `json:"family"`
	Size       float32 `json:"size"`
	LineHeight float32 `json:"lineHeight"`
}

// Padding is the room around a terminal's grid, in DIPs.
type Padding struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

// Scheme is a color scheme: a background, a foreground, a cursor, a
// selection, and the sixteen colors programs ask for.
type Scheme struct {
	Name       string   `json:"name"`
	Background string   `json:"background"`
	Foreground string   `json:"foreground"`
	Cursor     string   `json:"cursor"`
	Selection  string   `json:"selection"`
	Palette    []string `json:"palette"`
}

// Profile is a program a tab can run.
type Profile struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Command    string   `json:"command"`
	Args       []string `json:"args"`
	Dir        string   `json:"dir"`
	Color      string   `json:"color"`
	Hidden     bool     `json:"hidden"`
	Discovered bool     `json:"discovered"`
}

// Keybinding is one key and the action it stands for.
type Keybinding struct {
	Key    string `json:"key"`
	Action string `json:"action"`
}

// languageNames names the language settings of versions before the catalogs,
// which said a region: the tags the catalogs are filed under. A file that
// holds one keeps its language instead of falling back to the system's.
var languageNames = map[string]string{"zh-CN": "zh-Hans", "en-US": "en"}

// Default is the settings a window starts with, before any file.
func Default() Settings {
	return Settings{
		Language:              "",
		Theme:                 "dark",
		StartOn:               "default",
		Font:                  Font{Family: DefaultFontFamily(), Size: 13, LineHeight: 1.2},
		ColorScheme:           "Campbell",
		Padding:               Padding{X: 12, Y: 8},
		BackgroundOpacity:     1.0,
		CursorShape:           "bar",
		CursorBlink:           true,
		ScrollbackMB:          10,
		PaddingColor:          "background",
		MinimumContrast:       0,
		ScrollMultiplier:      1,
		ClipboardRead:         "deny",
		ClipboardWrite:        "allow",
		RightClick:            "copyOrPaste",
		PasteMultilineConfirm: true,
		ConfirmClose:          ConfirmRunning,
		CloseOnLastTab:        true,
		ConfirmCloseWindow:    true,
		NewTabPosition:        "last",
		TabWidth:              "compact",
		MaxPanes:              4,
		Keybindings:           DefaultKeybindings(),
	}
}

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

// Clone is a copy of the settings that shares nothing with them, so that a
// form may edit it, and the file's copy stay as it was.
func (s Settings) Clone() Settings {
	out := s
	out.Schemes = append([]Scheme(nil), s.Schemes...)
	for i := range out.Schemes {
		out.Schemes[i].Palette = append([]string(nil), s.Schemes[i].Palette...)
	}
	out.Profiles = append([]Profile(nil), s.Profiles...)
	for i := range out.Profiles {
		out.Profiles[i].Args = append([]string(nil), s.Profiles[i].Args...)
	}
	out.Keybindings = append([]Keybinding(nil), s.Keybindings...)
	// Extra is shared like the strings are: a form edits the settings, never
	// the map of keys it does not know.
	return out
}

// Parse reads settings.json: a key the file leaves out keeps its default,
// so a shorter file is still a whole set of settings. A file that is JSON but
// names a setting outside the range the window can run with is rejected whole:
// the caller keeps the settings it already has rather than running a
// half-applied file. Parse answers with no settings when it rejects a file, so
// a caller cannot mistake the defaults for what the file said.
func Parse(b []byte) (Settings, error) {
	s := Default()
	if err := json.Unmarshal(b, &s); err != nil {
		return Settings{}, fmt.Errorf("settings: %w", err)
	}
	// A key no field names is kept as it was written, so that saving from the
	// settings page does not drop what a person or a later version put there.
	// Reading b twice rather than teaching Settings to read itself keeps the
	// struct a plain shape to write, and a file that is not an object at all is
	// already rejected above.
	var all map[string]json.RawMessage
	if err := json.Unmarshal(b, &all); err != nil {
		return Settings{}, fmt.Errorf("settings: %w", err)
	}
	for key := range knownKeys {
		delete(all, key)
	}
	if len(all) > 0 {
		s.Extra = all
	} else {
		s.Extra = nil
	}
	if tag, ok := languageNames[s.Language]; ok {
		s.Language = tag
	}
	// A word no setting takes is a hand-written mistake, not a range to
	// refuse a file over: it goes back to the default, as docs/配置参考.md
	// says of the names that are not one of the words. The ranges below are
	// different — a font no program can draw is a file the window cannot run.
	s.fallbackUnknown()
	if err := s.Validate(); err != nil {
		return Settings{}, err
	}
	return s, nil
}

// knownKeys are the keys Settings has a field for, and so the ones Extra does
// not keep. A key added to Settings without being added here would be kept in
// Extra too, which Marshal writes second: the field would win, and the test of
// the two lists would say so.
var knownKeys = map[string]struct{}{
	"language": {}, "theme": {}, "defaultProfile": {}, "startOn": {},
	"startDirectory": {},

	"font": {}, "colorScheme": {}, "padding": {}, "backgroundOpacity": {},
	"cursorShape": {}, "cursorBlink": {}, "scrollbackMB": {},
	"paddingColor": {}, "minimumContrast": {},

	"copyOnSelect": {}, "rightClick": {}, "pasteMultilineConfirm": {},
	"confirmClose": {}, "closeOnLastTab": {}, "confirmCloseWindow": {},
	"scrollMultiplier": {}, "clipboardRead": {}, "clipboardWrite": {},

	"newTabPosition": {}, "tabWidth": {}, "alwaysOnTop": {},
	"titleFromActiveTerminal": {},

	"maxPanes": {},

	"schemes": {}, "profiles": {}, "keybindings": {},
}

// Marshal writes settings.json, with the keys Parse kept that no field names:
// a file that a person extended by hand, or one a later version wrote, keeps
// what this window does not understand. A key a field names is written as the
// field now holds it.
func Marshal(s Settings) ([]byte, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("settings: %w", err)
	}
	if len(s.Extra) > 0 {
		fields := map[string]json.RawMessage{}
		if err := json.Unmarshal(b, &fields); err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}
		for k, v := range s.Extra {
			if _, known := fields[k]; !known {
				fields[k] = v
			}
		}
		if b, err = json.Marshal(fields); err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}
	}
	var indented bytes.Buffer
	if err := json.Indent(&indented, b, "", "  "); err != nil {
		return nil, fmt.Errorf("settings: %w", err)
	}
	indented.WriteByte('\n')
	return indented.Bytes(), nil
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
