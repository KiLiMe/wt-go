// Package i18n names the messages the window says. The words live in
// ../locales, a file per language, and this package is the way the packages
// that may not import them say a message by the key it is filed under.
package i18n

import "testing"

// TestT covers a message in the source language, which is what the window
// says until the settings name another: a key no catalog has comes back as
// itself, rather than as nothing, and the numbered ones say their number.
func TestT(t *testing.T) {
	if got := T("closeTab"); got != "Close Tab" {
		t.Errorf("closeTab = %q", got)
	}
	if got := T("notAWord"); got != "notAWord" {
		t.Errorf("missing key = %q, want the key", got)
	}
	if got := F("switchToTab", map[string]any{"n": 3}); got != "Switch to Tab 3" {
		t.Errorf("switchToTab = %q", got)
	}
}

// TestHas covers the check the tests make of the keys the actions name, which
// no typed accessor checks.
func TestHas(t *testing.T) {
	if !Has("closeTab") {
		t.Error("Has(closeTab) = false")
	}
	if Has("notAWord") {
		t.Error("Has(notAWord) = true")
	}
}

// TestValid covers the tags a settings file may name, and the ones it may not,
// which the window shows the system's language for instead of panicking.
func TestValid(t *testing.T) {
	for _, tag := range []string{"", "en", "zh-Hans", "fr-FR", "zh-CN", "en_US"} {
		if !Valid(tag) {
			t.Errorf("Valid(%q) = false", tag)
		}
	}
	for _, tag := range []string{"zh CN", " not a tag"} {
		if Valid(tag) {
			t.Errorf("Valid(%q) = true", tag)
		}
	}
}

// TestAutonym covers the names the settings offer the languages by.
func TestAutonym(t *testing.T) {
	if got := Autonym("zh-Hans"); got != "简体中文" {
		t.Errorf("Autonym(zh-Hans) = %q", got)
	}
	if got := Autonym("en"); got != "English" {
		t.Errorf("Autonym(en) = %q", got)
	}
}
