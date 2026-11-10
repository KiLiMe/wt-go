package ui

import (
	"testing"
)

// commands is what the palette is asked to find.
func commands() []Command {
	return []Command{
		{Action: "tab.new", Title: "New Tab", Group: "Tab", Key: "ctrl+shift+t"},
		{Action: "pane.split.right", Title: "Split Right", Group: "Pane", Key: "alt+shift+right"},
		{Action: "pane.split.down", Title: "Split Down", Group: "Pane", Key: "alt+shift+down"},
		{Action: "palette.open", Title: "Command Palette", Group: "Window", Key: "ctrl+shift+p"},
		{Action: "find.open", Title: "Find", Group: "Find", Key: "ctrl+shift+f"},
	}
}

// titles are the titles of commands, in order.
func titles(cmds []Command) []string {
	out := make([]string, len(cmds))
	for i, c := range cmds {
		out[i] = c.Title
	}
	return out
}

// TestMatch covers what the palette lists: every command for no filter, the
// commands a filter names, the title before the group, and none for what names
// nothing.
func TestMatch(t *testing.T) {
	if got := titles(Match("", commands())); len(got) != 5 || got[0] != "New Tab" {
		t.Errorf("Match with no filter = %v", got)
	}
	if got := titles(Match("   ", commands())); len(got) != 5 {
		t.Errorf("Match of spaces = %v", got)
	}
	for _, tc := range []struct {
		filter string
		want   []string
	}{
		// Letters in order, a title match first.
		{"pal", []string{"Command Palette"}},
		// A word of the group finds the commands that belong to it.
		{"pane", []string{"Split Right", "Split Down"}},
		{"split", []string{"Split Down", "Split Right"}},
		// Words of the action, which the title does not hold.
		{"split down", []string{"Split Down"}},
		{"nothing", nil},
		{"zzz", nil},
	} {
		got := titles(Match(tc.filter, commands()))
		if len(got) != len(tc.want) {
			t.Errorf("Match(%q) = %v, want %v", tc.filter, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("Match(%q) = %v, want %v", tc.filter, got, tc.want)
				break
			}
		}
	}
}

// TestMatchRanksTitlesFirst covers the score: a title that holds what is typed
// comes before a command the word only belongs to.
func TestMatchRanksTitlesFirst(t *testing.T) {
	cmds := []Command{
		{Action: "pane.close", Title: "Close Pane", Group: "Pane"},
		{Action: "window.pane", Title: "Something", Group: "Pane"},
	}
	if got := titles(Match("pane", cmds)); got[0] != "Close Pane" {
		t.Errorf("Match(pane) = %v, want the title first", got)
	}
}

// TestFuzzy covers the scoring of letters in order, which finds a title by the
// letters of its words.
func TestFuzzy(t *testing.T) {
	if fuzzy("nwb", "New Tab") < 0 {
		t.Error("fuzzy(nwb, New Tab) found nothing")
	}
	if fuzzy("xxx", "New Tab") >= 0 {
		t.Error("fuzzy(xxx, New Tab) found something")
	}
	// The start of a word scores more than the middle of one.
	if fuzzy("t", "Tab") <= fuzzy("b", "Tab") {
		t.Error("the start of a word did not score more")
	}
	if fuzzy("", "New Tab") != 0 {
		t.Error("an empty filter did not score 0")
	}
}
