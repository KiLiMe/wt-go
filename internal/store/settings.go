// Package store is settings.json: the shape of the file, its defaults, and
// the way it turns into bytes and back. It is pure: the app reads and writes
// the file, store only knows what is in it.
package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

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
	// PointerText gives the pointer of a terminal the I-beam while no
	// program asks for one, as Ghostty shows it; off, it is the arrow, as
	// Windows Terminal shows it. Read when a tab is made, like the cursor.
	PointerText  bool `json:"pointerText"`
	ScrollbackMB int  `json:"scrollbackMB"`

	// PaddingColor is what the room around a terminal's grid shows:
	// "background" is the window's, "extend" the cells nearest to it, so
	// that a program that paints its own background reaches the edges,
	// and "always" extends them whatever they hold.
	PaddingColor string `json:"paddingColor"`
	// MinimumContrast is the least contrast text has with the background
	// it is drawn on, as a WCAG ratio from 0 to 21: 0 leaves every cell as
	// the program drew it, 4.5 is the ratio the guidelines call readable.
	MinimumContrast float32 `json:"minimumContrast"`

	CopyOnSelect                bool    `json:"copyOnSelect"`
	RightClick                  string  `json:"rightClick"`
	RightClickCopyPastePriority bool    `json:"rightClickCopyPastePriority"`
	PasteMultilineConfirm       bool    `json:"pasteMultilineConfirm"`
	ConfirmClose                Confirm `json:"confirmClose"`
	CloseOnLastTab              bool    `json:"closeOnLastTab"`
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

	// Links is what a link in a terminal is: a path a program printed, a URL,
	// or a hyperlink the program set (OSC 8), and what a click and the pointer
	// do with one.
	Links Links `json:"links"`

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

// Links is what a link in a terminal is, what a click does with one, and what
// the pointer shows of one. A link comes from three places, and they are worth
// telling apart: a program that says this text is a link, a URL printed as
// text, and a file path the terminal guesses at from the words on the screen.
// The last is the loosest, and the one a pointer resting on build output turns
// into a hand over.
type Links struct {
	// Open is the switch over the three below, and over a click: off, a click
	// opens nothing and nothing is underlined, whatever they say. Off by
	// default: a click that hands a path to another program is the user's to
	// ask for, not something the first stray click finds out.
	Open bool `json:"open"`
	// Program is a hyperlink the program set (OSC 8). Nothing is guessed: the
	// program says the text is a link, to a file or a place in a web page.
	Program bool `json:"program"`
	// URL is a URL printed as text, which the terminal's pattern finds: what a
	// program printed it meant to be read. Off by default, so that turning
	// clicks on is a decision about the one kind the user asked for.
	URL bool `json:"url"`
	// Path is a file path printed as text, which the terminal guesses at -- a
	// word with an extension or a slash, and the place in the file after it.
	// On by default: grep and a compiler name files in a form a click can use.
	Path bool `json:"path"`
	// Hover is whether the pointer underlines a link and shows the hand as
	// soon as it is over one, as Windows Terminal does. Off, Control has to be
	// held, as in Ghostty: a link is then underlined on a deliberate gesture
	// rather than wherever the pointer happens to rest.
	Hover bool `json:"hover"`
}

// Font is the font of a terminal's text.
type Font struct {
	// Family is the families a terminal tries, in order, and is left out of the
	// file while it holds the families the window ships: Marshal drops it and
	// FillEmptyFont names it again, so that \"\" and a missing key say one thing.
	Family string `json:"family,omitempty"`
	// Weight is how heavy the text is, from 100 to 900, of the weights the
	// family has: 400 is the regular face. Bold text is drawn 300 heavier, and
	// at least 700, which is what terminal.Font.Weight does.
	Weight     int     `json:"weight"`
	Size       float32 `json:"size"`
	LineHeight float32 `json:"lineHeight"`
}

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
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Dir     string   `json:"dir"`
	Color   string   `json:"color"`
	// Icon is the mark the tab wears: a preset such as "pwsh7:official",
	// one character, or "" for the one the command asks for.
	Icon       string `json:"icon"`
	Hidden     bool   `json:"hidden"`
	Discovered bool   `json:"discovered"`
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
		Font:                  Font{Family: DefaultFontFamily(), Weight: 400, Size: 13, LineHeight: 1.2},
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
		Links:                 Links{Program: true, Path: true},
		RightClick:            "copyOrPaste",
		PasteMultilineConfirm: true,
		ConfirmClose:          ConfirmRunning,
		CloseOnLastTab:        true,
		ConfirmCloseWindow:    true,
		NewTabPosition:        "last",
		TabWidth:              "equal",
		MaxPanes:              4,
		Keybindings:           DefaultKeybindings(),
		// A list the file may leave out is still a list: an empty one writes
		// itself as [] rather than null, which is what a person reading the file
		// expects of a setting that holds many things.
		Schemes:  []Scheme{},
		Profiles: []Profile{},
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
	s.fillEmptyLists()
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
	"cursorShape": {}, "cursorBlink": {}, "pointerText": {}, "scrollbackMB": {},
	"paddingColor": {}, "minimumContrast": {},

	"copyOnSelect": {}, "rightClick": {}, "rightClickCopyPastePriority": {},
	"pasteMultilineConfirm": {},
	"confirmClose":          {}, "closeOnLastTab": {}, "confirmCloseWindow": {},
	"scrollMultiplier": {}, "clipboardRead": {}, "clipboardWrite": {},
	"links": {},

	"newTabPosition": {}, "tabWidth": {}, "alwaysOnTop": {},
	"titleFromActiveTerminal": {},

	"maxPanes": {},

	"schemes": {}, "profiles": {}, "keybindings": {},
}

// Marshal writes settings.json, with the keys Parse kept that no field names:
// a file that a person extended by hand, or one a later version wrote, keeps
// what this window does not understand. A key a field names is written as the
// field now holds it, except the families the window ships: those are left out,
// since a file that leaves the family out reads as the default families, and a
// machine drawing with the default list follows the list the version it runs
// ships rather than the one it was first opened with. Every other setting is
// written as it is, so that a file says what a window does.
func Marshal(s Settings) ([]byte, error) {
	if s.Font.Family == DefaultFontFamily() {
		s.Font.Family = ""
	}
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
