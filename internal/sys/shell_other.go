//go:build !windows

package sys

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Shell is a shell this machine has, as it is found: no name, no color, no
// place in the settings.
type Shell struct {
	ID      string
	Command string
	Args    []string
}

// shellsFile lists the shells a POSIX system offers, one path a line.
const shellsFile = "/etc/shells"

// fallbackShell is the shell a POSIX system is required to have.
const fallbackShell = "/bin/sh"

// runnable are the shells a tab can run, by the id of their program:
// /etc/shells also lists the ones a person may not start interactively, like
// git-shell and rbash, and a tab that runs one is a tab that does nothing.
// What the strip calls them, and the dot they wear, is profile's business.
var runnable = map[string]bool{
	"bash": true,
	"zsh":  true,
	"fish": true,
	"sh":   true,
	"nu":   true,
	"dash": true,
	"ksh":  true,
	"tcsh": true,
	"pwsh": true,
}

// Discover is the shells this machine has, in the order a strip shows them:
// the user's own shell, the ones /etc/shells lists that a tab can run, then
// /bin/sh. A shell that is not installed is not listed.
func Discover() []Shell {
	return discoverShells(os.Getenv("SHELL"), os.ReadFile, exec.LookPath)
}

// discoverShells is Discover with the machine's answers handed in, so a test
// can say what /etc/shells holds and what is installed.
func discoverShells(envShell string, readFile func(string) ([]byte, error), lookPath func(string) (string, error)) []Shell {
	var shells []Shell
	seen := make(map[string]bool)
	add := func(path string) {
		if path == "" {
			return
		}
		command, err := lookPath(path)
		if err != nil {
			return
		}
		id := shellID(command)
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		shells = append(shells, Shell{ID: id, Command: command})
	}

	add(envShell)
	if data, err := readFile(shellsFile); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") || !knownShell(line) {
				continue
			}
			add(line)
		}
	}
	add(fallbackShell)
	return shells
}

// knownShell reports whether path is a shell a tab can be.
func knownShell(path string) bool {
	return runnable[shellID(path)]
}

// shellID is the stable name of a shell program: the lowercase base name of
// its path, so /bin/bash and /usr/local/bin/bash are the same shell.
func shellID(path string) string {
	return strings.ToLower(filepath.Base(path))
}
