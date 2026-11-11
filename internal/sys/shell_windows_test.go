//go:build windows

package sys

import (
	"errors"
	"reflect"
	"testing"
)

// TestDiscover covers what the strip lists: the shells the candidate table
// names, in that order, each only when PATH answers for it; Bash unless that
// Bash is the one Windows keeps for the WSL distributions; and the
// distributions last.
func TestDiscover(t *testing.T) {
	tests := []struct {
		name    string
		path    map[string]string
		distros []string
		want    []Shell
	}{
		{
			name: "the candidates' order, whatever the order PATH answers in",
			path: map[string]string{
				"pwsh.exe":       `C:\Program Files\PowerShell\7\pwsh.exe`,
				"powershell.exe": `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`,
				"cmd.exe":        `C:\Windows\system32\cmd.exe`,
			},
			want: []Shell{
				{ID: "pwsh7", Command: `C:\Program Files\PowerShell\7\pwsh.exe`},
				{ID: "powershell", Command: `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`},
				{ID: "cmd", Command: `C:\Windows\system32\cmd.exe`},
			},
		},
		{
			name: "a shell PATH does not answer for is not listed",
			path: map[string]string{"cmd.exe": `C:\Windows\system32\cmd.exe`},
			want: []Shell{{ID: "cmd", Command: `C:\Windows\system32\cmd.exe`}},
		},
		{
			name: "the bash.exe of Windows runs the default distribution, not a Bash",
			path: map[string]string{"bash.exe": `C:\Windows\System32\bash.exe`},
		},
		{
			name: "other bash.exe are one",
			path: map[string]string{
				"bash.exe": `C:\Program Files\Git\bin\bash.exe`,
				"cmd.exe":  `C:\Windows\system32\cmd.exe`,
			},
			want: []Shell{
				{ID: "bash", Command: `C:\Program Files\Git\bin\bash.exe`},
				{ID: "cmd", Command: `C:\Windows\system32\cmd.exe`},
			},
		},
		{
			name: "the distributions come last, each with its own name in the arguments",
			path: map[string]string{
				"cmd.exe": `C:\Windows\system32\cmd.exe`,
				"wsl.exe": `C:\Windows\system32\wsl.exe`,
			},
			distros: []string{"Debian", "Ubuntu"},
			want: []Shell{
				{ID: "cmd", Command: `C:\Windows\system32\cmd.exe`},
				{ID: "wsl:Debian", Command: `C:\Windows\system32\wsl.exe`, Args: []string{"-d", "Debian"}},
				{ID: "wsl:Ubuntu", Command: `C:\Windows\system32\wsl.exe`, Args: []string{"-d", "Ubuntu"}},
			},
		},
		{
			name:    "no wsl.exe lists no distribution",
			path:    map[string]string{"cmd.exe": `C:\Windows\system32\cmd.exe`},
			distros: []string{"Ubuntu"},
			want:    []Shell{{ID: "cmd", Command: `C:\Windows\system32\cmd.exe`}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookPath := func(file string) (string, error) {
				if path, ok := tt.path[file]; ok {
					return path, nil
				}
				return "", errors.New("not found")
			}
			if got := discover(lookPath, tt.distros); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("discover = %+v, want %+v", got, tt.want)
			}
		})
	}
}
