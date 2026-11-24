// Package store is settings.json: the shape of the file, its defaults, and
// the way it turns into bytes and back. It is pure: the app reads and writes
// the file, store only knows what is in it.
package store

import (
	"encoding/json"
)

// Boundary: the shape of settings.json -- the struct and the value types it
// holds, with the words each may say. Defaults are defaults.go, comparisons
// equal.go, and reading and writing a file file.go.

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
