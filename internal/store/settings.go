// Package store is settings.json: the shape of the file, its defaults, and
// the way it turns into bytes and back. It is pure: the app reads and writes
// the file, store only knows what is in it.
package store

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// Settings is settings.json, as docs/settings.md describes it.
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

	CopyOnSelect          bool   `json:"copyOnSelect"`
	RightClick            string `json:"rightClick"`
	PasteMultilineConfirm bool   `json:"pasteMultilineConfirm"`
	ConfirmClose          bool   `json:"confirmClose"`
	CloseOnLastTab        bool   `json:"closeOnLastTab"`

	// NewTabPosition is where a tab opens: "last" or "first". TabWidth is
	// how wide tabs are: "compact", to the name, or "equal", all alike.
	NewTabPosition string `json:"newTabPosition"`
	TabWidth       string `json:"tabWidth"`
	// AlwaysOnTop keeps the window over the others; TitleFromActiveTerminal
	// lets the window's title follow the title of the terminal in front.
	AlwaysOnTop             bool `json:"alwaysOnTop"`
	TitleFromActiveTerminal bool `json:"titleFromActiveTerminal"`

	// StartSize is the grid the window starts with, in cells; zero leaves
	// the size to the window.
	StartSize Size `json:"startSize"`

	MaxPanes int `json:"maxPanes"`

	Schemes     []Scheme     `json:"schemes"`
	Profiles    []Profile    `json:"profiles"`
	Keybindings []Keybinding `json:"keybindings"`
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

// Size is a terminal grid: how many cells wide and high.
type Size struct {
	Columns int `json:"columns"`
	Rows    int `json:"rows"`
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

// Default is the settings a window starts with, before any file.
func Default() Settings {
	return Settings{
		Language:              "zh-CN",
		Theme:                 "dark",
		StartOn:               "default",
		Font:                  Font{Family: DefaultFontFamily(), Size: 13, LineHeight: 1.2},
		ColorScheme:           "Campbell",
		Padding:               Padding{X: 12, Y: 8},
		BackgroundOpacity:     1.0,
		CursorShape:           "bar",
		CursorBlink:           true,
		ScrollbackMB:          10,
		RightClick:            "copyOrPaste",
		PasteMultilineConfirm: true,
		ConfirmClose:          true,
		CloseOnLastTab:        true,
		NewTabPosition:        "last",
		TabWidth:              "compact",
		MaxPanes:              4,
		Keybindings:           DefaultKeybindings(),
	}
}

// Equal reports whether two sets of settings are the same, which is how the
// window knows whether what it shows is what the file holds.
func Equal(a, b Settings) bool {
	return reflect.DeepEqual(a, b)
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
	if err := s.Validate(); err != nil {
		return Settings{}, err
	}
	return s, nil
}

// Marshal writes settings.json.
func Marshal(s Settings) ([]byte, error) {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("settings: %w", err)
	}
	return append(b, '\n'), nil
}

// DefaultKeybindings are the keys keys.md lists for the actions the window
// has today.
func DefaultKeybindings() []Keybinding {
	kb := []Keybinding{
		{Key: "ctrl+shift+t", Action: ActionTabNew},
		{Key: "ctrl+t", Action: ActionTabNew},
		{Key: "ctrl+shift+w", Action: ActionPaneClose},
		{Key: "ctrl+tab", Action: ActionTabNext},
		{Key: "ctrl+shift+tab", Action: ActionTabPrev},
		{Key: "ctrl+shift+d", Action: ActionTabDuplicate},
		{Key: "alt+shift+d", Action: ActionPaneSplitAuto},
		{Key: "alt+shift+right", Action: ActionPaneSplitRight},
		{Key: "alt+shift+down", Action: ActionPaneSplitDown},
		{Key: "alt+left", Action: ActionPaneFocusLeft},
		{Key: "alt+right", Action: ActionPaneFocusRight},
		{Key: "alt+up", Action: ActionPaneFocusUp},
		{Key: "alt+down", Action: ActionPaneFocusDown},
		{Key: "alt+shift+w", Action: ActionPaneCloseOthers},
		{Key: "ctrl+comma", Action: ActionSettings},
	}
	for n := 1; n <= 9; n++ {
		kb = append(kb,
			Keybinding{Key: "ctrl+" + strconv.Itoa(n), Action: GotoAction(n)},
			Keybinding{Key: "ctrl+shift+" + strconv.Itoa(n), Action: ProfileAction(n)},
		)
	}
	return kb
}
