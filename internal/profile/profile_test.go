package profile

import (
	"testing"

	"wtgo/internal/store"
)

// TestResolve covers the order of the strip: the found shells first, in the
// order sys found them, with the settings' profile of an id in its place, and
// then the profiles the settings name that no shell matched.
func TestResolve(t *testing.T) {
	configured := []store.Profile{
		{ID: "pwsh7", Name: "我的 PowerShell", Color: "#ff0000"},
		{ID: "", Name: "no id"},
		{ID: "mine", Name: "我的", Command: "cmd.exe"},
	}
	found := []Found{
		{ID: "pwsh7", Command: "pwsh.exe"},
		{ID: "cmd", Command: "cmd.exe"},
		{ID: "wsl:Ubuntu", Command: "wsl.exe", Args: []string{"-d", "Ubuntu"}},
	}
	got := Resolve(configured, found)
	if len(got) != 4 {
		t.Fatalf("len = %d, want 4: %+v", len(got), got)
	}
	if got[0].Name != "我的 PowerShell" || got[0].Color != "#ff0000" || got[0].Discovered {
		t.Errorf("configured profile lost its name or color: %+v", got[0])
	}
	if got[1].ID != "cmd" || got[1].Color != "#cccccc" || !got[1].Discovered {
		t.Errorf("cmd = %+v", got[1])
	}
	if got[2].ID != "wsl:Ubuntu" || got[2].Color != "#e95420" {
		t.Errorf("wsl = %+v", got[2])
	}
	if got[3].ID != "mine" || got[3].Name != "我的" {
		t.Errorf("configured profile with no shell = %+v", got[3])
	}
}

// TestDefault covers picking the profile a new tab runs: the named one while
// the window shows it, else the shell of the machine.
func TestDefault(t *testing.T) {
	profiles := []Profile{
		{ID: "a", Hidden: true},
		{ID: "b"},
		{ID: "c"},
	}
	found := []Found{{ID: "b", Command: "/bin/b"}}
	device := Device(found)
	if p := Default(profiles, "c", device); p.ID != "c" {
		t.Errorf("named = %+v", p)
	}
	// The device answers with what sys found, not with the profile of the
	// settings that has the same id: that one may be hidden, or hold another
	// command line.
	for _, id := range []string{"", "a", "missing"} {
		if p := Default(profiles, id, device); p.ID != "b" || p.Command != "/bin/b" {
			t.Errorf("Default(%q) = %+v, want the shell of the machine", id, p)
		}
	}
}

// TestDevice covers the shell of the machine as a profile: the first shell sys
// found, named and colored from the table, and the zero profile when the
// machine has none.
func TestDevice(t *testing.T) {
	if p := Device(nil); p.ID != "" || p.Command != "" || p.Name != "" {
		t.Errorf("Device(nil) = %+v, want the zero profile", p)
	}
	got := Device([]Found{{ID: "cmd", Command: `C:\Windows\System32\cmd.exe`, Args: []string{"/k"}}})
	if got.ID != "cmd" || got.Command == "" || got.Name == "" || got.Color == "" || !got.Discovered || len(got.Args) != 1 {
		t.Errorf("cmd = %+v", got)
	}
}

// TestCommandLine covers a profile that runs nothing, which means the
// user's shell.
func TestCommandLine(t *testing.T) {
	if got := (Profile{}).CommandLine(); got != nil {
		t.Errorf("empty profile = %v, want nil", got)
	}
	got := Profile{Command: "wsl.exe", Args: []string{"-d", "Ubuntu"}}.CommandLine()
	if len(got) != 3 || got[0] != "wsl.exe" || got[2] != "Ubuntu" {
		t.Errorf("command line = %v", got)
	}
}

// TestColor covers the dot a discovered profile wears: a known shell gets its
// own color, and a shell nothing knows gets none.
func TestColor(t *testing.T) {
	tests := []struct {
		id   string
		want string
	}{
		{"pwsh7", "#3b78ff"},
		{"powershell", "#2671be"},
		{"cmd", "#cccccc"},
		{"wsl:Ubuntu", "#e95420"},
		{"bash", "#4eaa25"},
		{"zsh", "#e5c07b"},
		{"fish", "#56b6c2"},
		{"sh", "#b0b0b0"},
		{"nu", "#7c6cf5"},
		{"dash", "#ff9e64"},
		{"ksh", "#c678dd"},
		{"tcsh", "#e06c9f"},
		{"elvish", ""},
	}
	for _, tt := range tests {
		if got := Color(tt.id); got != tt.want {
			t.Errorf("Color(%q) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

// TestName covers what the strip calls a shell: the name the table gives a
// shell it knows, and the name of the program for one it does not.
func TestName(t *testing.T) {
	tests := []struct {
		id      string
		command string
		want    string
	}{
		{"pwsh7", `C:\Program Files\PowerShell\7\pwsh.exe`, "PowerShell 7"},
		{"cmd", `C:\Windows\system32\cmd.exe`, "Command Prompt"},
		{"bash", `/usr/bin/bash`, "Bash"},
		{"elvish", `/usr/local/bin/elvish`, "elvish"},
		{"xonsh", `C:\tools\xonsh.exe`, "xonsh"},
		{"wsl:Ubuntu", `C:\Windows\system32\wsl.exe`, "wsl"},
		{"", "", ""},
	}
	for _, tt := range tests {
		if got := Name(tt.id, tt.command); got != tt.want {
			t.Errorf("Name(%q, %q) = %q, want %q", tt.id, tt.command, got, tt.want)
		}
	}
}

// TestReset covers a profile the window found, put back the way this machine
// has it: the probe's program and arguments, the table's name and dot, and on
// the menu again. What the table does not own, the start directory, stays.
func TestReset(t *testing.T) {
	p := store.Profile{
		ID:         "pwsh7",
		Name:       "我的 pwsh",
		Command:    `C:\gone\pwsh.exe`,
		Args:       []string{"-NoLogo"},
		Dir:        `D:\work`,
		Color:      "#123456",
		Hidden:     true,
		Discovered: true,
	}
	found := Found{ID: "pwsh7", Command: `C:\pwsh.exe`, Args: []string{"-NoLogo", "-NoProfile"}}
	got := Reset(p, found)

	if got.Name != "PowerShell 7" || got.Color != "#3b78ff" {
		t.Errorf("name, color = %q, %q, want the table's", got.Name, got.Color)
	}
	if got.Command != `C:\pwsh.exe` || len(got.Args) != 2 || got.Args[1] != "-NoProfile" {
		t.Errorf("command line = %q %q, want the one the probe found", got.Command, got.Args)
	}
	if got.Hidden || !got.Discovered {
		t.Errorf("hidden, discovered = %v, %v, want a found shell on the menu", got.Hidden, got.Discovered)
	}
	if got.Dir != `D:\work` {
		t.Errorf("dir = %q, want the directory the person chose", got.Dir)
	}
	// The arguments belong to the profile, not to the found shell the caller
	// keeps: a change to one must not reach the other.
	got.Args[0] = "-Changed"
	if found.Args[0] != "-NoLogo" {
		t.Errorf("found args = %q, want the caller's own", found.Args)
	}
}
