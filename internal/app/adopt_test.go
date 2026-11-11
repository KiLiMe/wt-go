package app

import (
	"reflect"
	"testing"

	"wtgo/internal/profile"
	"wtgo/internal/store"
)

// TestAdopt covers the shells the window writes into the settings: a shell the
// file does not name is added as a found profile, a profile a person wrote
// keeps its place and its name, and a found profile whose program is gone is
// hidden rather than dropped.
func TestAdopt(t *testing.T) {
	found := []profile.Found{
		{ID: "pwsh7", Command: `C:\Windows\pwsh.exe`},
		{ID: "cmd", Command: `C:\Windows\system32\cmd.exe`},
	}
	tests := []struct {
		name string
		have []store.Profile
		want []store.Profile
		// missing are the paths the fake machine does not have.
		missing []string
		added   bool
	}{
		{
			name: "an empty file takes them all, in the order found",
			want: []store.Profile{
				{ID: "pwsh7", Name: "PowerShell 7", Command: `C:\Windows\pwsh.exe`, Color: "#3b78ff", Discovered: true},
				{ID: "cmd", Name: "Command Prompt", Command: `C:\Windows\system32\cmd.exe`, Color: "#cccccc", Discovered: true},
			},
			added: true,
		},
		{
			name: "a found profile that is there stays as it was edited",
			have: []store.Profile{
				{ID: "pwsh7", Name: "我的 PowerShell", Command: `C:\pwsh.exe`, Color: "#ff0000", Discovered: true},
			},
			want: []store.Profile{
				{ID: "pwsh7", Name: "我的 PowerShell", Command: `C:\pwsh.exe`, Color: "#ff0000", Discovered: true},
				{ID: "cmd", Name: "Command Prompt", Command: `C:\Windows\system32\cmd.exe`, Color: "#cccccc", Discovered: true},
			},
			added: true,
		},
		{
			name: "the new ones go below the found ones and above the written ones",
			have: []store.Profile{
				{ID: "mine", Name: "我的", Command: "cmd.exe"},
				{ID: "pwsh7", Name: "PowerShell 7", Command: `C:\Windows\pwsh.exe`, Color: "#3b78ff", Discovered: true},
			},
			want: []store.Profile{
				{ID: "mine", Name: "我的", Command: "cmd.exe"},
				{ID: "pwsh7", Name: "PowerShell 7", Command: `C:\Windows\pwsh.exe`, Color: "#3b78ff", Discovered: true},
				{ID: "cmd", Name: "Command Prompt", Command: `C:\Windows\system32\cmd.exe`, Color: "#cccccc", Discovered: true},
			},
			added: true,
		},
		{
			name: "a found profile whose program is gone is hidden",
			have: []store.Profile{
				{ID: "pwsh7", Name: "PowerShell 7", Command: `C:\Windows\pwsh.exe`, Color: "#3b78ff", Discovered: true},
				{ID: "bash", Name: "Bash", Command: `C:\Program Files\Git\bin\bash.exe`, Color: "#4eaa25", Discovered: true},
			},
			want: []store.Profile{
				{ID: "pwsh7", Name: "PowerShell 7", Command: `C:\Windows\pwsh.exe`, Color: "#3b78ff", Discovered: true},
				{ID: "bash", Name: "Bash", Command: `C:\Program Files\Git\bin\bash.exe`, Color: "#4eaa25", Discovered: true, Hidden: true},
				{ID: "cmd", Name: "Command Prompt", Command: `C:\Windows\system32\cmd.exe`, Color: "#cccccc", Discovered: true},
			},
			missing: []string{`C:\Program Files\Git\bin\bash.exe`},
			added:   true,
		},
		{
			name: "a second start changes nothing",
			have: []store.Profile{
				{ID: "pwsh7", Name: "PowerShell 7", Command: `C:\Windows\pwsh.exe`, Color: "#3b78ff", Discovered: true},
				{ID: "cmd", Name: "Command Prompt", Command: `C:\Windows\system32\cmd.exe`, Color: "#cccccc", Discovered: true},
			},
			want: []store.Profile{
				{ID: "pwsh7", Name: "PowerShell 7", Command: `C:\Windows\pwsh.exe`, Color: "#3b78ff", Discovered: true},
				{ID: "cmd", Name: "Command Prompt", Command: `C:\Windows\system32\cmd.exe`, Color: "#cccccc", Discovered: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists := func(path string) bool {
				for _, missing := range tt.missing {
					if path == missing {
						return false
					}
				}
				return true
			}
			s := store.Settings{Profiles: append([]store.Profile(nil), tt.have...)}
			if got := adopt(&s, found, exists); got != tt.added {
				t.Errorf("changed = %v, want %v", got, tt.added)
			}
			if len(s.Profiles) != len(tt.want) {
				t.Fatalf("profiles = %+v, want %+v", s.Profiles, tt.want)
			}
			for i := range tt.want {
				if !reflect.DeepEqual(s.Profiles[i], tt.want[i]) {
					t.Errorf("profile %d = %+v, want %+v", i, s.Profiles[i], tt.want[i])
				}
			}
		})
	}
}
