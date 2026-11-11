//go:build !windows

package sys

import (
	"errors"
	"reflect"
	"testing"
)

// TestShellID covers the id a shell path turns into: the lowercase base name,
// so the same shell found by two paths is one shell.
func TestShellID(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/bin/bash", "bash"},
		{"/usr/local/bin/bash", "bash"},
		{"/usr/bin/zsh", "zsh"},
		{"/usr/bin/fish", "fish"},
		{"/bin/sh", "sh"},
		{"/usr/bin/nu", "nu"},
		{"/usr/bin/dash", "dash"},
		{"/bin/ksh", "ksh"},
		{"/bin/tcsh", "tcsh"},
		{"/usr/bin/elvish", "elvish"},
	}
	for _, tt := range tests {
		if got := shellID(tt.path); got != tt.want {
			t.Errorf("shellID(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

// TestDiscoverShells covers the order of the shells: the user's own shell,
// then the ones /etc/shells names, then /bin/sh, each once, and only when it
// is installed.
func TestDiscoverShells(t *testing.T) {
	tests := []struct {
		name      string
		envShell  string
		shells    string
		readErr   error
		installed map[string]bool
		want      []Shell
	}{
		{
			name:     "user shell first, then the file",
			envShell: "/bin/zsh",
			shells:   "# comment\n\n/bin/bash\n/bin/zsh\n/bin/fish\n",
			installed: map[string]bool{
				"/bin/zsh": true, "/bin/bash": true, "/bin/fish": true, "/bin/sh": true,
			},
			want: []Shell{
				{ID: "zsh", Command: "/bin/zsh"},
				{ID: "bash", Command: "/bin/bash"},
				{ID: "fish", Command: "/bin/fish"},
				{ID: "sh", Command: "/bin/sh"},
			},
		},
		{
			name:     "missing user shell is skipped",
			envShell: "/bin/nope",
			shells:   "/bin/bash\n",
			installed: map[string]bool{
				"/bin/bash": true, "/bin/sh": true,
			},
			want: []Shell{
				{ID: "bash", Command: "/bin/bash"},
				{ID: "sh", Command: "/bin/sh"},
			},
		},
		{
			name:     "same shell by two paths is one shell",
			envShell: "",
			shells:   "/bin/bash\n/usr/local/bin/bash\n",
			installed: map[string]bool{
				"/bin/bash": true, "/usr/local/bin/bash": true, "/bin/sh": true,
			},
			want: []Shell{
				{ID: "bash", Command: "/bin/bash"},
				{ID: "sh", Command: "/bin/sh"},
			},
		},
		{
			name:     "a shell a tab cannot run is left out",
			envShell: "/bin/bash",
			shells:   "/bin/bash\n/usr/bin/git-shell\n/bin/rbash\n",
			installed: map[string]bool{
				"/bin/bash": true, "/usr/bin/git-shell": true, "/bin/rbash": true, "/bin/sh": true,
			},
			want: []Shell{
				{ID: "bash", Command: "/bin/bash"},
				{ID: "sh", Command: "/bin/sh"},
			},
		},
		{
			name:     "powershell is a shell here too",
			envShell: "",
			shells:   "/usr/bin/pwsh\n",
			installed: map[string]bool{
				"/usr/bin/pwsh": true, "/bin/sh": true,
			},
			want: []Shell{
				{ID: "pwsh", Command: "/usr/bin/pwsh"},
				{ID: "sh", Command: "/bin/sh"},
			},
		},
		{
			name:     "an unknown shell keeps its name",
			envShell: "/usr/bin/elvish",
			installed: map[string]bool{
				"/usr/bin/elvish": true, "/bin/sh": true,
			},
			want: []Shell{
				{ID: "elvish", Command: "/usr/bin/elvish"},
				{ID: "sh", Command: "/bin/sh"},
			},
		},
		{
			name:     "no file falls back to sh",
			envShell: "",
			readErr:  errors.New("no /etc/shells"),
			installed: map[string]bool{
				"/bin/sh": true,
			},
			want: []Shell{
				{ID: "sh", Command: "/bin/sh"},
			},
		},
		{
			name:      "nothing installed is no shell",
			envShell:  "",
			readErr:   errors.New("no /etc/shells"),
			installed: map[string]bool{},
			want:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readFile := func(string) ([]byte, error) {
				if tt.readErr != nil {
					return nil, tt.readErr
				}
				return []byte(tt.shells), nil
			}
			lookPath := func(file string) (string, error) {
				if tt.installed[file] {
					return file, nil
				}
				return "", errors.New("not found")
			}
			got := discoverShells(tt.envShell, readFile, lookPath)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Discover = %+v, want %+v", got, tt.want)
			}
		})
	}
}
