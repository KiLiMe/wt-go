package ui

import (
	"testing"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/store"
	"wtgo/locales"
)

// advancedActions is the window as the advanced page sees it: it records the
// questions asked and the folders opened.
type advancedActions struct {
	Actions
	asked  []string
	opened []string
}

func (a *advancedActions) AskAdvanced(kind string) { a.asked = append(a.asked, kind) }
func (a *advancedActions) OpenSettingsFolder()     { a.opened = append(a.opened, "settings") }
func (a *advancedActions) OpenProgramFolder()      { a.opened = append(a.opened, "program") }

// TestAdvancedPage covers the page: both buttons that change the file ask
// before they do it, and the folder buttons open the folders.
func TestAdvancedPage(t *testing.T) {
	s := store.Default()
	act := &advancedActions{}
	tt := myui.NewTester(func(c *myui.Context) {
		SettingsTab(c, Settings{Draft: &s, Page: PageAdvanced}, act)
	}, 900, 700)
	tt.Frame()

	if err := tt.Click(locales.Probe()); err != nil {
		t.Fatalf("no button probes: %v", err)
	}
	if err := tt.Click(locales.DeleteAndRestart()); err != nil {
		t.Fatalf("no button resets: %v", err)
	}
	want := []string{AskRediscover, AskReset}
	if len(act.asked) != len(want) {
		t.Fatalf("asked = %v, want %v", act.asked, want)
	}
	for i, k := range want {
		if act.asked[i] != k {
			t.Fatalf("asked = %v, want %v", act.asked, want)
		}
	}
	if err := tt.Click(locales.Open()); err != nil {
		t.Fatalf("no folder button: %v", err)
	}
	if len(act.opened) != 1 || act.opened[0] != "settings" {
		t.Fatalf("opened = %v, want the settings folder", act.opened)
	}
}
