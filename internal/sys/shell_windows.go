//go:build windows

package sys

import (
	"os/exec"
	"sort"

	"golang.org/x/sys/windows/registry"
)

// Shell is a shell this machine has, as it is found: no color, no place in
// the settings.
type Shell struct {
	ID      string
	Name    string
	Command string
	Args    []string
}

// Discover is the shells this machine has, in the order a strip shows them:
// PowerShell 7, Windows PowerShell, Command Prompt, then the WSL
// distributions. A shell that is not installed is not listed.
func Discover() []Shell {
	var shells []Shell
	if path, err := exec.LookPath("pwsh.exe"); err == nil {
		shells = append(shells, Shell{ID: "pwsh7", Name: "PowerShell 7", Command: path})
	}
	if path, err := exec.LookPath("powershell.exe"); err == nil {
		shells = append(shells, Shell{ID: "powershell", Name: "Windows PowerShell", Command: path})
	}
	if path, err := exec.LookPath("cmd.exe"); err == nil {
		shells = append(shells, Shell{ID: "cmd", Name: "Command Prompt", Command: path})
	}
	if path, err := exec.LookPath("wsl.exe"); err == nil {
		for _, name := range distributions() {
			shells = append(shells, Shell{
				ID:      "wsl:" + name,
				Name:    name,
				Command: path,
				Args:    []string{"-d", name},
			})
		}
	}
	return shells
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
