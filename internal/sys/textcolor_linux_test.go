//go:build linux

package sys

import (
	"strings"
	"testing"
)

// TestTextColorRules keeps the rules narrow: every font loses its subpixel
// order, and nothing else about a font changes.
func TestTextColorRules(t *testing.T) {
	if got := strings.Count(textColorRules, `name="rgba"`); got != 2 {
		t.Errorf("%d rgba edits, want 2", got)
	}
	if got := strings.Count(textColorRules, "<const>none</const>"); got != 2 {
		t.Errorf("%d rgba values, want 2", got)
	}
	for _, want := range []string{`target="pattern"`, `target="font"`} {
		if !strings.Contains(textColorRules, want) {
			t.Errorf("no edit for %s", want)
		}
	}
}
