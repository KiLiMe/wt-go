package app

import (
	"testing"

	"github.com/egoist/mygo/plugins/terminal"
)

// TestTerminalOptions covers the settings that name a terminal's behavior as
// words: each word is a choice on the settings page, and an unknown one leaves
// the default rather than asking the terminal for something it does not have.
func TestTerminalOptions(t *testing.T) {
	if got := scrollMultiplier(0); got.Precision != 1 || got.Discrete != 3 {
		t.Errorf("scrollMultiplier(0) = %+v, want the default", got)
	}
	if got := scrollMultiplier(2); got.Precision != 2 || got.Discrete != 6 {
		t.Errorf("scrollMultiplier(2) = %+v, want a notch to scroll twice as far", got)
	}
	for word, want := range map[string]terminal.PaddingColor{
		"background": terminal.PaddingBackground,
		"extend":     terminal.PaddingExtend,
		"always":     terminal.PaddingExtendAlways,
		"nonsense":   terminal.PaddingBackground,
	} {
		if got := paddingColor(word); got != want {
			t.Errorf("paddingColor(%q) = %v, want %v", word, got, want)
		}
	}
	for word, want := range map[string]terminal.ClipboardAccess{
		"deny":     terminal.ClipboardDeny,
		"allow":    terminal.ClipboardAllow,
		"ask":      terminal.ClipboardAsk,
		"nonsense": terminal.ClipboardDeny,
	} {
		if got := clipboardRead(word); got != want {
			t.Errorf("clipboardRead(%q) = %v, want %v", word, got, want)
		}
	}
}
