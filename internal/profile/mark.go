package profile

import (
	"strings"
	"unicode"

	"wtgo/internal/store"
)

// Boundary: the mark a tab wears -- the presets the settings page offers, the
// key a mark names, the one a program's name guesses, and the glyph a mark
// holds. Nothing here resolves a profile.

// Preset is one of the marks a tab wears by name: the key a settings file
// writes, the shape it draws and the palette that shape is painted in.
type Preset struct {
	Key   string
	Shape string
	Name  string
}

// presets is what this build draws, in the order the settings page lists it.
// It names the same keys as internal/icons, which draws them, and the two are
// kept together by a test in internal/ui: a key this table lists and that
// package does not draw would be a tab with no mark at all.
var presets = []Preset{
	{Key: "pwsh7:official", Shape: "PowerShell 7", Name: "official"},
	{Key: "pwsh7:violet", Shape: "PowerShell 7", Name: "violet"},
	{Key: "pwsh7:light", Shape: "PowerShell 7", Name: "light"},
	{Key: "powershell:classic", Shape: "Windows PowerShell", Name: "classic"},
	{Key: "powershell:ocean", Shape: "Windows PowerShell", Name: "ocean"},
	{Key: "powershell:amber", Shape: "Windows PowerShell", Name: "amber"},
	{Key: "cmd:black", Shape: "Command Prompt", Name: "black"},
	{Key: "cmd:matrix", Shape: "Command Prompt", Name: "matrix"},
	{Key: "cmd:amber", Shape: "Command Prompt", Name: "amber"},
	{Key: "bash:gnu", Shape: "Bash", Name: "gnu"},
	{Key: "bash:ubuntu", Shape: "Bash", Name: "ubuntu"},
	{Key: "bash:arch", Shape: "Bash", Name: "arch"},
}

// Presets are the marks the settings page offers.
func Presets() []Preset { return append([]Preset(nil), presets...) }

// IconKey is the preset a mark names, "" when it names none: a mark is a key
// when it holds a colon, and a colon in a key is what keeps the two kinds of
// mark apart without a second field to say which one it is.
func IconKey(mark string) string {
	if !strings.Contains(mark, ":") {
		return ""
	}
	if _, ok := presetByKey(mark); ok {
		return mark
	}
	return ""
}

// presetByKey is the preset a key names.
func presetByKey(key string) (Preset, bool) {
	for _, p := range presets {
		if p.Key == key {
			return p, true
		}
	}
	return Preset{}, false
}

// GuessIcon is the mark a program's name asks for: the shape of the shell it
// is, in that shape's first palette, and "" for a program no shape is drawn
// for, which wears the plain terminal mark instead.
//
// The name is what the profile's command ends in, so a profile someone writes
// for "pwsh" gets the mark without naming one, and a profile that names a path
// into a shell's own directory gets it too.
func GuessIcon(command string) string {
	if i := strings.LastIndexAny(command, `\/`); i >= 0 {
		command = command[i+1:]
	}
	command = strings.ToLower(strings.TrimSuffix(strings.TrimSuffix(command, ".exe"), ".exe"))
	switch command {
	case "pwsh", "pwsh7", "pwsh-preview":
		return "pwsh7:official"
	case "powershell", "powershell_ise":
		return "powershell:classic"
	case "cmd":
		return "cmd:black"
	case "bash":
		return "bash:gnu"
	}
	return ""
}

// IconFor is the mark a profile wears: the preset it names, else the one
// character it names, else the one its command asks for. A mark that is
// neither a preset nor a character (a key of a build that drew it, say) is
// read as the command's, so a settings file never leaves a tab without a mark.
func IconFor(p store.Profile) string {
	if key := IconKey(p.Icon); key != "" {
		return key
	}
	// A mark is a character only when the whole of it is one: a word with no
	// colon in it is a key someone mistyped, not a character and two more.
	if g := Glyph(p.Icon); g != "" && g == strings.TrimSpace(p.Icon) {
		return g
	}
	return GuessIcon(p.Command)
}

// Glyph is the one character a mark holds, "" when it holds none: a tab has
// room for a character, not for a word, and the first is the one the person
// meant. A flag is two runes, and a keycap or an emoji carries marks after the
// rune it modifies, so a character is more than a rune.
func Glyph(mark string) string {
	if strings.Contains(mark, ":") {
		return ""
	}
	runes := []rune(strings.TrimSpace(mark))
	if len(runes) == 0 {
		return ""
	}
	n := 1
	if regional(runes[0]) && len(runes) > 1 && regional(runes[1]) {
		n = 2
	}
	for n < len(runes) {
		r := runes[n]
		if r == '\u200d' || r == '\ufe0f' || r == '\ufe0e' || r == '\u20e3' ||
			unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) ||
			runes[n-1] == '\u200d' {
			n++
			continue
		}
		break
	}
	return string(runes[:n])
}

// regional reports whether r is half of a flag: the two are one character.
func regional(r rune) bool { return r >= 0x1f1e6 && r <= 0x1f1ff }
