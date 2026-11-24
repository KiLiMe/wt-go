package app

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"wtgo/internal/pane"
	"wtgo/internal/store"
	"wtgo/internal/sys"
	"wtgo/internal/ui"
)

// TestSafelyShowsAPanic covers the boundary around a reload: a panic is caught
// and shown instead of escaping the watcher.
func TestSafelyShowsAPanic(t *testing.T) {
	a := &App{settings: store.Default(), saved: store.Default()}
	a.safely(func() { panic("boom") })
	if !strings.Contains(a.note, "boom") {
		t.Errorf("note = %q, want the panic", a.note)
	}
}

// TestSaveKeepsUnknownKeys covers a file a person extended by hand: a switch
// turned in the settings page and saved writes the settings it shows and keeps
// the keys this window does not know, which reading the file and writing it
// back would otherwise drop.
func TestSaveKeepsUnknownKeys(t *testing.T) {
	settingsDir(t)
	file := []byte(`{"theme":"dark","unknownField":7}`)
	if err := sys.WriteSettings(file); err != nil {
		t.Fatal(err)
	}
	a := newApp("")
	a.settings.Theme = "light"
	a.SaveSettings()
	if !strings.Contains(a.note, "Saved") {
		t.Fatalf("note = %q, want the save to have gone through", a.note)
	}
	after, ok, err := sys.ReadSettings()
	if err != nil || !ok {
		t.Fatalf("read: %v, %v", err, ok)
	}
	if !strings.Contains(string(after), `"unknownField": 7`) {
		t.Errorf("the save dropped a key it does not know:\n%s", after)
	}
	if !strings.Contains(string(after), `"theme": "light"`) {
		t.Errorf("the save did not write the theme the page shows:\n%s", after)
	}
}

// TestSaveRejectsExtremes covers a save that would write a setting no window
// can run with: it says why, and writes nothing.
func TestSaveRejectsExtremes(t *testing.T) {
	settingsDir(t)
	a := &App{settings: store.Default(), saved: store.Default(), savedData: []byte("old")}
	a.settings.Font.Size = 0
	a.SaveSettings()
	if !strings.Contains(a.note, "font.size") {
		t.Fatalf("note = %q, want it to name font.size", a.note)
	}
	path, err := sys.SettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("settings file written: %v", err)
	}
}

// TestSettingsNoteShowsTheActionAndTheFileError covers a settings page with
// both to say: what the last action did is not hidden behind the reason the
// file is not settings, and that reason comes back once the action note goes.
func TestSettingsNoteShowsTheActionAndTheFileError(t *testing.T) {
	a := &App{configErr: "settings.json is not valid: boom"}
	if got := a.settingsNote(); got != a.configErr {
		t.Errorf("note = %q, want the file's reason", got)
	}
	a.note = "settings.json: access is denied"
	got := a.settingsNote()
	if !strings.Contains(got, a.note) {
		t.Errorf("note = %q, want the read error", got)
	}
	if !strings.Contains(got, a.configErr) {
		t.Errorf("note = %q, want the file's reason kept", got)
	}
	a.note = ""
	if got := a.settingsNote(); got != a.configErr {
		t.Errorf("note after clearing = %q, want the file's reason back", got)
	}
}

// TestSaveErrorIsShownWithTheFileError covers a save that fails while the file
// is not settings: the page names what the save refused and still says why the
// file is not settings.
func TestSaveErrorIsShownWithTheFileError(t *testing.T) {
	settingsDir(t)
	bad := []byte("{oops")
	if err := sys.WriteSettings(bad); err != nil {
		t.Fatal(err)
	}
	a := &App{settings: store.Default(), saved: store.Default(), savedData: []byte("{}")}
	a.reload(bad)
	if a.configErr == "" {
		t.Fatal("no note about the bad file")
	}
	a.settings.Font.Size = 0
	a.SaveSettings()
	note := a.settingsNote()
	if !strings.Contains(note, "font.size") {
		t.Errorf("note = %q, want the save's reason", note)
	}
	if !strings.Contains(note, "settings.json") {
		t.Errorf("note = %q, want the file's reason kept", note)
	}
}

// TestSaveRefusalKeepsTheDraftAndSaysWhy covers a save the page refuses: the
// reason stays on the page after the next action's note takes the footer, the
// draft keeps every change, and the page it is on stays.
func TestSaveRefusalKeepsTheDraftAndSaysWhy(t *testing.T) {
	settingsDir(t)
	a, baseData := settingsApp(t, store.Default())
	a.settingsTab = &pane.Tab{Name: "Settings"}
	a.SelectSettingsPage(ui.PageProfiles)

	a.settings.Font.Size = 0
	a.settings.Padding.X = 300
	a.SaveSettings()
	if a.validation == "" {
		t.Fatal("the refusal left no reason for the page to show")
	}
	for _, name := range []string{"font.size", "padding.x"} {
		if !strings.Contains(a.validation, name) {
			t.Errorf("reason = %q, want it to name %s", a.validation, name)
		}
	}
	if a.settingsPage != ui.PageProfiles {
		t.Errorf("page = %q, want the one the save was refused on", a.settingsPage)
	}
	if a.settings.Font.Size != 0 || a.settings.Padding.X != 300 {
		t.Errorf("draft = size %v padding %v, want what the page edited", a.settings.Font.Size, a.settings.Padding.X)
	}
	if !a.dirty() {
		t.Error("the refused draft is not unsaved")
	}

	// The footer says what the last action did, and the reason behind it is
	// what the reader still has to correct.
	a.setNote("settings opened")
	if a.validation == "" {
		t.Error("another note took the reason with it")
	}
	st := a.settingsState()
	if st.Validation != a.validation {
		t.Errorf("the page shows %q, want %q", st.Validation, a.validation)
	}

	path, err := sys.SettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	// settingsApp wrote the file, so a second write would be the refused save's.
	onDisk, ok, err := sys.ReadSettings()
	if err != nil || !ok {
		t.Fatalf("read: %v, %v", err, ok)
	}
	if !bytes.Equal(onDisk, baseData) {
		t.Errorf("file = %q, want the content before the refused save at %s", onDisk, path)
	}
	if !bytes.Equal(a.savedData, baseData) {
		t.Error("savedData changed for a save that wrote nothing")
	}

	// Correcting it saves, and the reason goes with the save.
	a.settings.Font.Size = 13
	a.settings.Padding.X = 8
	a.SaveSettings()
	if a.validation != "" {
		t.Errorf("reason = %q, want none after a save that went through", a.validation)
	}
	if a.dirty() {
		t.Error("the save left the draft unsaved")
	}

	// A refusal the person leaves behind takes its reason with it.
	a.settings.Font.Size = 0
	a.SaveSettings()
	if a.validation == "" {
		t.Fatal("the refusal left no reason")
	}
	a.DiscardSettings()
	if a.validation != "" {
		t.Errorf("reason = %q, want none after discarding the draft", a.validation)
	}
}

// TestSaveRefusalIsShownWithTheFileError covers a save the page refuses while
// the footer holds the file's reason: both stay, the page's own reason in
// front of it.
func TestSaveRefusalIsShownWithTheFileError(t *testing.T) {
	settingsDir(t)
	a, _ := settingsApp(t, store.Default())
	a.settings.Font.Size = 0
	a.SaveSettings()
	a.configErr = "settings.json is not valid: boom"
	note := a.settingsNote()
	if !strings.Contains(note, a.validation) || !strings.Contains(note, a.configErr) {
		t.Errorf("note = %q, want the refusal and the file's reason", note)
	}
}
