//go:build !windows

package sys

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Shell is a shell this machine has.
type Shell struct {
	ID      string
	Name    string
	Command string
	Args    []string
}

// shellsFile lists the shells a POSIX system offers, one path a line.
const shellsFile = "/etc/shells"

// fallbackShell is the shell a POSIX system is required to have.
const fallbackShell = "/bin/sh"

// shellNames is what the strip calls a shell it knows, by the id of its
// program. A shell it does not know keeps its own name.
var shellNames = map[string]string{
	"bash": "Bash",
	"zsh":  "Zsh",
	"fish": "Fish",
	"sh":   "sh",
	"nu":   "Nushell",
	"dash": "dash",
	"ksh":  "Korn Shell",
	"tcsh": "tcsh",
	"pwsh": "PowerShell 7",
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
		shells = append(shells, Shell{ID: id, Name: shellName(id), Command: command})
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

// knownShell reports whether path is a shell a tab can be: /etc/shells also
// lists the ones a person may not start interactively, like git-shell and
// rbash, and a tab that runs one is a tab that does nothing.
func knownShell(path string) bool {
	_, ok := shellNames[shellID(path)]
	return ok
}

// shellID is the stable name of a shell program: the lowercase base name of
// its path, so /bin/bash and /usr/local/bin/bash are the same shell.
func shellID(path string) string {
	return strings.ToLower(filepath.Base(path))
}

// shellName is what the strip calls a shell.
func shellName(id string) string {
	if name, ok := shellNames[id]; ok {
		return name
	}
	return id
}
