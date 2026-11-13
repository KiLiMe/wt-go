// Package profile is what a tab runs: the profiles the settings name, and
// the ones the machine has. It is pure: sys finds the shells, profile turns
// them into profiles and picks the default.
package profile

import (
	"strings"

	"wtgo/internal/store"
)

// Profile is a program a tab can run, and what the tab strip calls it.
type Profile struct {
	ID         string
	Name       string
	Command    string
	Args       []string
	Dir        string
	Color      string
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
