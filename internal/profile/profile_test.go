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
		{ID: "pwsh7", Name: "PowerShell 7", Command: "pwsh.exe"},
		{ID: "cmd", Name: "命令提示符", Command: "cmd.exe"},
		{ID: "wsl:Ubuntu", Name: "Ubuntu", Command: "wsl.exe", Args: []string{"-d", "Ubuntu"}},
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

// TestDefault covers picking the profile a new tab runs.
func TestDefault(t *testing.T) {
	profiles := []Profile{
		{ID: "a", Hidden: true},
		{ID: "b"},
		{ID: "c"},
	}
	if p, ok := Default(profiles, "c"); !ok || p.ID != "c" {
		t.Errorf("named = %+v, %v", p, ok)
	}
	if p, ok := Default(profiles, "a"); !ok || p.ID != "b" {
		t.Errorf("hidden named = %+v, %v", p, ok)
	}
	if p, ok := Default(profiles, "missing"); !ok || p.ID != "b" {
		t.Errorf("missing named = %+v, %v", p, ok)
	}
	if _, ok := Default([]Profile{{ID: "a", Hidden: true}}, ""); ok {
		t.Error("picked a hidden profile")
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
