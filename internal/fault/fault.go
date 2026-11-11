// Package fault gives each failure the program knows a code. What the user
// reads stays one line — the code and what it means — and the checks, the
// evidence and the fix live in docs/故障排查.md under the same code.
package fault

import (
	"sort"
	"strings"
)

// Code is the tag a known failure carries: the hundreds digit says where it
// happens (E1 before the window is up, E2 reading the settings, E3 once it
// runs).
type Code string

const (
	// Before the window is up.
	NoShell     Code = "E101" // no profile could be started
	NativeLib   Code = "E102" // libghostty-vt is missing or did not load
	NoConfigDir Code = "E103" // the directory the settings live in is not available

	// Reading settings.json.
	BadSettings   Code = "E201" // not JSON, or not settings
	BadField      Code = "E202" // a field is out of range
	BadKeybinding Code = "E203" // a keybinding names a key that cannot be read
	BadFont       Code = "E204" // the font family is not a monospace font
	SettingsWrite Code = "E205" // settings.json could not be written

	// While it runs.
	ChildFailed   Code = "E301" // a shell could not be started
	Crashed       Code = "E302" // the window stopped with a panic
	NoProcessList Code = "E303" // the process list could not be read
	UnknownShell  Code = "E304" // the window did not learn the process of a shell
	NoFontconfig  Code = "E305" // the text color fix could not be applied
)

// about is the one line the program shows for a code.
var about = map[Code]string{
	NoShell:       "No profile could be started: every shell this machine has is missing or not runnable",
	NativeLib:     "libghostty-vt was not found or did not load, so no terminal can be drawn",
	NoConfigDir:   "The directory for settings.json is not available (LOCALAPPDATA is not set or not writable)",
	BadSettings:   "settings.json is not JSON settings; the copy is kept and the defaults are used",
	BadField:      "A field of settings.json is out of range; the defaults are used",
	BadKeybinding: "A keybinding names a key wt-go cannot read; that one binding is dropped, the rest work",
	BadFont:       "The font family is not a monospace font, or is not installed; the terminal draws with monospace instead",
	SettingsWrite: "settings.json could not be written, so the profiles the window found are not in the file",
	ChildFailed:   "A shell could not be started, so the tab has nothing in it",
	Crashed:       "The window stopped with a panic; the stack is in the crash log",
	NoProcessList: "The process list could not be read, so whether a tab still runs a program is unknown",
	UnknownShell:  "The window did not learn the process of a shell, so whether its tab still runs a program is unknown",
	NoFontconfig:  "fontconfig was not found, so text may draw in the wrong color",
}

// doc is where the details for a code are kept, from the project root.
const doc = "docs/故障排查.md"

// String makes a code print as itself.
func (c Code) String() string { return string(c) }

// Short is what the code means, in one line.
func (c Code) Short() string { return about[c] }

// Known says whether this program hands out the code: Find relies on it to
// tell a fault line from a log line.
func (c Code) Known() bool {
	_, ok := about[c]
	return ok
}

// All is every code, in order, for the tests and the doc.
func All() []Code {
	out := make([]Code, 0, len(about))
	for c := range about {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Line is what the program prints for a fault: the code, what it means, and
// where the rest is written. detail carries what only this run knows (a field
// name, an error) and is left out when empty.
func Line(c Code, detail string) string {
	s := string(c) + " " + c.Short()
	if detail != "" {
		s += " (" + detail + ")"
	}
	return s + ", see " + doc
}

// Find picks the fault line out of what a log holds, so that a window that
// died still has a line to show for it. A log line starts with what log
// prints (a date, "wt-go:"), so the code is looked for among the words of a
// line rather than at its start.
func Find(lines []string) (string, bool) {
	for _, l := range lines {
		for _, word := range strings.Fields(l) {
			if Code(strings.Trim(word, "[](),:;")).Known() {
				return strings.TrimSpace(l), true
			}
		}
	}
	return "", false
}

// In picks the fault out of text, splitting it into lines first: a crash log
// is read whole, and the code sits somewhere in it.
func In(text string) (string, bool) {
	return Find(strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n"))
}
