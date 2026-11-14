// Package profile is what a tab runs: the profiles the settings name, and
// the ones the machine has. It is pure: sys finds the shells, profile turns
// them into profiles and picks the default.
package profile

import (
	"strings"
	"unicode"

	"wtgo/internal/store"
)

// Profile is a program a tab can run, and what the tab strip calls it.
type Profile struct {
	ID      string
	Name    string
	Command string
	Args    []string
	Dir     string
	Color   string
	// Icon is the mark the tab wears: a preset of this package, one
	// character, or "" for the one the command asks for.
	Icon       string
	Hidden     bool
	Discovered bool
}

// Found is a shell sys found on this machine, before profile gives it a name
// and a place among the ones the settings name.
type Found struct {
	ID      string
	Command string
	Args    []string
}

// known is what the strip calls a shell, and the dot it wears, by the id of
// its program. A shell it does not know is called by the name of its program
// and wears no dot.
var known = map[string]struct{ Name, Color string }{
	"pwsh7":      {"PowerShell 7", "#3b78ff"},
	"pwsh":       {"PowerShell 7", "#3b78ff"},
	"powershell": {"Windows PowerShell", "#2671be"},
	"cmd":        {"Command Prompt", "#cccccc"},
	"bash":       {"Bash", "#4eaa25"},
	"zsh":        {"Zsh", "#e5c07b"},
	"fish":       {"Fish", "#56b6c2"},
	"sh":         {"sh", "#b0b0b0"},
	"nu":         {"Nushell", "#7c6cf5"},
	"dash":       {"dash", "#ff9e64"},
	"ksh":        {"Korn Shell", "#c678dd"},
	"tcsh":       {"tcsh", "#e06c9f"},
}

// Name is what the strip calls a shell of an id, whose program is command:
// the name the table gives it, or the name of its program.
func Name(id, command string) string {
	if k, ok := known[id]; ok {
		return k.Name
	}
	if i := strings.LastIndexAny(command, `\/`); i >= 0 {
		command = command[i+1:]
	}
	command = strings.TrimSuffix(strings.TrimSuffix(command, ".exe"), ".EXE")
	return command
}

// Color is the dot a shell of an id wears, "" for none.
func Color(id string) string {
	if strings.HasPrefix(id, "wsl:") {
		return "#e95420"
	}
	return known[id].Color
}

// Resolve is the profiles the window offers: the shells found on this machine
// in the order sys found them, each replaced by the settings' profile of the
// same id, so renaming or recoloring a discovered profile keeps it, and hidden
// ones stay out of the way; and then the profiles the settings name that no
// shell matched. The shells come first because they are what the machine
// already has, and a profile someone adds goes after them: adding one should
// not push down the shells they already know.
func Resolve(configured []store.Profile, found []Found) []Profile {
	named := make(map[string]store.Profile, len(configured))
	for _, p := range configured {
		if p.ID != "" {
			named[p.ID] = p
		}
	}
	profiles := make([]Profile, 0, len(configured)+len(found))
	used := make(map[string]bool, len(named))
	for _, f := range found {
		if p, ok := named[f.ID]; ok {
			used[f.ID] = true
			profiles = append(profiles, fromStore(p))
			continue
		}
		profiles = append(profiles, Profile{
			ID:         f.ID,
			Name:       Name(f.ID, f.Command),
			Command:    f.Command,
			Args:       f.Args,
			Color:      Color(f.ID),
			Icon:       GuessIcon(f.Command),
			Discovered: true,
		})
	}
	seen := make(map[string]bool, len(profiles))
	for _, p := range profiles {
		seen[p.ID] = true
	}
	for _, p := range configured {
		if p.ID == "" || used[p.ID] || seen[p.ID] {
			continue
		}
		profiles = append(profiles, fromStore(p))
	}
	return profiles
}

// fromStore is a profile the settings hold, as the window runs it.
func fromStore(p store.Profile) Profile {
	return Profile{
		ID:         p.ID,
		Name:       p.Name,
		Command:    p.Command,
		Args:       p.Args,
		Dir:        p.Dir,
		Color:      p.Color,
		Icon:       IconFor(p),
		Hidden:     p.Hidden,
		Discovered: p.Discovered,
	}
}

// Default is the profile a new tab runs: the one with id while the window
// shows it, else the device, which is the shell of this machine.
//
// Falling back to the first profile in the list instead would make "the
// default" a position in a list, which anyone can change by hiding a profile:
// following the device is a fact about the machine, not about the order.
func Default(profiles []Profile, id string, device Profile) Profile {
	if id != "" {
		for _, p := range profiles {
			if p.ID == id && !p.Hidden {
				return p
			}
		}
	}
	return device
}

// Device is the shell of this machine that a new tab runs when the settings
// name none: the first one sys found, as it was found.
//
// It takes nothing from the settings' profiles, not even their names: a profile
// someone hid is one they chose not to offer, so its name, its color and its
// command line are not the machine's shell either. A machine with no shell at
// all leaves the zero profile, which has the terminal look for one itself.
func Device(found []Found) Profile {
	if len(found) == 0 {
		return Profile{}
	}
	f := found[0]
	return Profile{
		ID:         f.ID,
		Name:       Name(f.ID, f.Command),
		Command:    f.Command,
		Args:       f.Args,
		Color:      Color(f.ID),
		Icon:       GuessIcon(f.Command),
		Discovered: true,
	}
}

// CommandLine is the program and its arguments, as a terminal starts it;
// nil runs the user's shell.
func (p Profile) CommandLine() []string {
	if p.Command == "" {
		return nil
	}
	return append([]string{p.Command}, p.Args...)
}

// Reset is a profile the window found, as the machine has it now: the program
// and its arguments from the probe, the name and the dot from the table, and
// not hidden, because the shell is there. The start directory stays: where a
// tab begins is a choice the person made, and the machine knows nothing about
// it.
func Reset(p store.Profile, f Found) store.Profile {
	p.Command, p.Args = f.Command, append([]string(nil), f.Args...)
	p.Name, p.Color = Name(f.ID, f.Command), Color(f.ID)
	p.Hidden, p.Discovered = false, true
	return p
}

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
