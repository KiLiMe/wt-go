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

// Found is a shell sys found on this machine, before profile gives it a
// place among the ones the settings name.
type Found struct {
	ID      string
	Name    string
	Command string
	Args    []string
}

// Resolve is the profiles the window offers: the shells found on this machine
// in the order sys found them, each replaced by the settings' profile of the
// same id, so renaming or recoloring a discovered profile keeps it; and then
// the profiles the settings name that no shell matched. The shells come first
// because a settings file that names no default profile starts the first
// profile in the list, and a profile someone adds should not take over the
// tabs they did not ask it for.
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
			Name:       f.Name,
			Command:    f.Command,
			Args:       f.Args,
			Color:      colorOf(f.ID),
			Discovered: true,
		})
	}
	for _, p := range configured {
		if p.ID == "" || used[p.ID] {
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

// Default picks the profile a new tab runs when nothing names one: the one
// with id when it is there and shown, else the first shown one.
func Default(profiles []Profile, id string) (Profile, bool) {
	if id != "" {
		for _, p := range profiles {
			if p.ID == id && !p.Hidden {
				return p, true
			}
		}
	}
	for _, p := range profiles {
		if !p.Hidden {
			return p, true
		}
	}
	return Profile{}, false
}

// CommandLine is the program and its arguments, as a terminal starts it;
// nil runs the user's shell.
func (p Profile) CommandLine() []string {
	if p.Command == "" {
		return nil
	}
	return append([]string{p.Command}, p.Args...)
}

// colorOf is the dot a profile of a known kind wears, as Windows Terminal
// gives its shells a color.
func colorOf(id string) string {
	switch {
	case id == "pwsh7":
		return "#3b78ff"
	case id == "powershell":
		return "#2671be"
	case id == "cmd":
		return "#cccccc"
	case strings.HasPrefix(id, "wsl:"):
		return "#e95420"
	}
	return ""
}
