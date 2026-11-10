package fault

import (
	"strings"
	"testing"
)

func TestAll(t *testing.T) {
	all := All()
	if len(all) < 8 {
		t.Fatalf("All() = %v, want every code", all)
	}
	seen := map[Code]bool{}
	for i, c := range all {
		if seen[c] {
			t.Errorf("%s is listed twice", c)
		}
		seen[c] = true
		if i > 0 && all[i-1] >= c {
			t.Errorf("All() is not in order: %s before %s", all[i-1], c)
		}
		if c.Short() == "" {
			t.Errorf("%s has no meaning", c)
		}
		if !c.Known() {
			t.Errorf("%s is in All() but not known", c)
		}
	}
}

func TestKnown(t *testing.T) {
	if Code("E999").Known() {
		t.Error("E999 is not a code this program hands out")
	}
	if !BadSettings.Known() {
		t.Error("E201 is a code")
	}
}

func TestLine(t *testing.T) {
	got := Line(BadField, "font.size must be between 4 and 200")
	for _, want := range []string{"E202", "font.size", "between 4 and 200", doc} {
		if !strings.Contains(got, want) {
			t.Errorf("Line() = %q, want it to hold %q", got, want)
		}
	}
	if plain := Line(NoShell, ""); strings.Contains(plain, "()") {
		t.Errorf("Line() = %q, want no empty detail", plain)
	}
}

func TestFind(t *testing.T) {
	lines := []string{
		"2026-11-10T23:00:00+08:00",
		"wt-go: E203 a keybinding names a key wt-go cannot read (keybinding \"ctrl++\"), see docs/故障排查.md",
		"goroutine 1 [running]:",
	}
	got, ok := Find(lines)
	if !ok || !strings.HasPrefix(got, "wt-go: E203") {
		t.Errorf("Find() = %q, %v, want the fault line", got, ok)
	}
	if _, ok := Find([]string{"wt-go: nothing is wrong"}); ok {
		t.Error("Find() found a fault in a line that has none")
	}
	if _, ok := In("panic: boom\nwt-go: E302 the window stopped with a panic, see docs/故障排查.md\n"); !ok {
		t.Error("In() did not find the fault in a crash log")
	}
}
