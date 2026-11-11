//go:build windows

package sys

import (
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// Shell is a shell this machine has, as it is found: no name, no color, no
// place in the settings.
type Shell struct {
	ID      string
	Command string
	Args    []string
}

// candidate is a shell a tab can run, and the programs that are it.
type candidate struct {
	ID   string
	Exes []string
	Args []string
}

// candidates are the shells Windows offers, in the order a strip shows them:
// the ones a person is most likely to want first.
//
// A shell that is not on PATH is not listed. Where a program is usually
// installed is a guess, and a guess that is wrong on the machine that needs it
// most, where the program sits elsewhere: a profile that is missing is a
// profile a person can add, one that runs the wrong program is not.
var candidates = []candidate{
	{ID: "pwsh7", Exes: []string{"pwsh.exe"}},
	{ID: "powershell", Exes: []string{"powershell.exe"}},
	{ID: "bash", Exes: []string{"bash.exe"}},
	{ID: "cmd", Exes: []string{"cmd.exe"}},
	{ID: "nu", Exes: []string{"nu.exe"}},
}

// Discover is the shells this machine has, in the order a strip shows them:
// PowerShell 7, Windows PowerShell, Bash, Command Prompt, Nushell, then the
// WSL distributions. A shell that is not installed is not listed.
func Discover() []Shell {
	return discover(exec.LookPath, distributions())
}

// discover is Discover with the machine's answers handed in, so that a test
// can say what PATH holds and which distributions there are.
func discover(lookPath func(string) (string, error), distros []string) []Shell {
	var shells []Shell
	for _, c := range candidates {
		for _, exe := range c.Exes {
			path, err := lookPath(exe)
			if err != nil || wslStub(path) {
				continue
			}
			shells = append(shells, Shell{ID: c.ID, Command: path, Args: c.Args})
			break
		}
	}
	wsl, err := lookPath("wsl.exe")
	if err != nil {
		return shells
	}
	for _, name := range distros {
		shells = append(shells, Shell{ID: "wsl:" + name, Command: wsl, Args: []string{"-d", name}})
	}
	return shells
}

// wslStub reports whether path is the bash.exe Windows keeps for the WSL
// distributions: it runs the default distribution, which the strip lists by
// name already, so listing it as a Bash of its own would be the same shell
// under two names.
func wslStub(path string) bool {
	return strings.EqualFold(filepath.Base(path), "bash.exe") &&
		strings.EqualFold(filepath.Base(filepath.Dir(path)), "system32")
}

// distributions is the WSL distributions the registry lists, in name order.
func distributions() []string {
	key, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Lxss`,
		registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
	if err != nil {
		return nil
	}
	defer key.Close()
	keys, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return nil
	}
	var found []string
	for _, name := range keys {
		sub, err := registry.OpenKey(key, name, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		distro, _, err := sub.GetStringValue("DistributionName")
		sub.Close()
		if err != nil || distro == "" {
			continue
		}
		found = append(found, distro)
	}
	sort.Strings(found)
	return found
}
