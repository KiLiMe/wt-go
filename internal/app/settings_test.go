package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"wtgo/internal/action"
	"wtgo/internal/pane"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/sys"
	"wtgo/internal/ui"
	"wtgo/locales"
)

// settingsDir points the settings path at a directory of this test's own.
func settingsDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("LOCALAPPDATA", dir)
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)
	return dir
}

// TestLoadKeepsTheOriginal covers a file that is not settings: load answers
// with no settings, so a caller cannot mistake defaults for what the file said,
// and with the bytes the file held.
func TestLoadKeepsTheOriginal(t *testing.T) {
	settingsDir(t)
	bad := []byte("{oops")
	if err := sys.WriteSettings(bad); err != nil {
		t.Fatal(err)
	}
	settings, data, _, err := load()
	if err == nil {
		t.Fatal("loaded a file that is not settings")
	}
	if !store.Equal(settings, store.Settings{}) {
		t.Errorf("settings = %+v, want none: a rejected file is not partly applied", settings)
	}
	if !bytes.Equal(data, bad) {
		t.Errorf("data = %q, want the original", data)
	}
	onDisk, ok, err := sys.ReadSettings()
	if err != nil || !ok {
		t.Fatalf("read: %v, %v", err, ok)
	}
	if !bytes.Equal(onDisk, bad) {
		t.Errorf("file = %q, want the original", onDisk)
	}
}

// TestNewAppKeepsABadFile covers the window opening on a bad file: it shows
// the defaults, says why on the settings page, and keeps the original.
func TestNewAppKeepsABadFile(t *testing.T) {
	settingsDir(t)
	bad := []byte("{oops")
	if err := sys.WriteSettings(bad); err != nil {
		t.Fatal(err)
	}
	a := newApp("")
	if a.configErr == "" {
		t.Fatal("no note about the bad file")
	}
	if !strings.Contains(a.configErr, "settings.json") || !strings.Contains(a.configErr, "bad-") {
		t.Errorf("configErr = %q, want the file and where it was kept", a.configErr)
	}
	if !bytes.Equal(a.badData, bad) {
		t.Errorf("badData = %q, want the original", a.badData)
	}
	if !store.Equal(a.saved, store.Default()) {
		t.Error("saved settings are not the defaults")
	}
	dir, err := sys.Dir()
	if err != nil {
		t.Fatal(err)
	}
	archives, err := filepath.Glob(filepath.Join(dir, "settings.json.bad-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(archives) != 1 {
		t.Errorf("archives = %v, want one", archives)
	}
}

// TestReloadKeepsABadFileOnce covers the watcher seeing the same bad file
// every second: it reports it once, and keeps one copy.
func TestReloadKeepsABadFileOnce(t *testing.T) {
	settingsDir(t)
	a := &App{settings: store.Default(), saved: store.Default(), savedData: []byte("{}")}
	bad := []byte("{oops")
	a.reload(bad)
	if a.configErr == "" {
		t.Fatal("no note about the bad file")
	}
	if !bytes.Equal(a.badData, bad) {
		t.Errorf("badData = %q, want the original", a.badData)
	}
	a.reload(bad)
	dir, err := sys.Dir()
	if err != nil {
		t.Fatal(err)
	}
	archives, err := filepath.Glob(filepath.Join(dir, "settings.json.bad-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(archives) != 1 {
		t.Errorf("archives = %v, want one", archives)
	}
}

// TestRefreshForgetsABadFileThatIsGone covers a file that was not settings and
// then was deleted: reading it again says nothing about it, so the settings
// page does not keep explaining a file that is no longer there.
func TestRefreshForgetsABadFileThatIsGone(t *testing.T) {
	settingsDir(t)
	bad := []byte("{oops")
	if err := sys.WriteSettings(bad); err != nil {
		t.Fatal(err)
	}
	a := &App{settings: store.Default(), saved: store.Default(), savedData: []byte("{}")}
	a.reload(bad)
	if a.configErr == "" || a.badData == nil {
		t.Fatal("the bad file was not reported")
	}

	path, err := sys.SettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	a.refreshSettings()
	if a.configErr != "" {
		t.Errorf("configErr = %q, want none for a file that is gone", a.configErr)
	}
	if a.badData != nil {
		t.Errorf("badData = %q, want none for a file that is gone", a.badData)
	}
	if note := a.settingsNote(); note != "" {
		t.Errorf("note = %q, want none", note)
	}
}

// TestPollForgetsABadFileThatIsGone covers the watcher seeing the bad file
// deleted: it clears the reason once, and goes on reading without dispatching
// anything while there is no file to read.
func TestPollForgetsABadFileThatIsGone(t *testing.T) {
	settingsDir(t)
	bad := []byte("{oops")
	if err := sys.WriteSettings(bad); err != nil {
		t.Fatal(err)
	}
	a := &App{settings: store.Default(), saved: store.Default(), savedData: []byte("{}")}
	a.reload(bad)
	if a.configErr == "" || a.badData == nil {
		t.Fatal("the bad file was not reported")
	}

	path, err := sys.SettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	calls := 0
	dispatch := func(fn func()) {
		calls++
		fn()
	}
	a.poll(dispatch)
	if calls != 1 {
		t.Fatalf("dispatched %d times, want one", calls)
	}
	if a.configErr != "" {
		t.Errorf("configErr = %q, want none for a file that is gone", a.configErr)
	}
	if a.badData != nil {
		t.Errorf("badData = %q, want none for a file that is gone", a.badData)
	}

	a.poll(dispatch)
	if calls != 1 {
		t.Errorf("dispatched %d times for a file that stays gone, want one", calls)
	}
}

// settingsApp is a window whose file already holds the settings.
func settingsApp(t *testing.T, s store.Settings) (*App, []byte) {
	t.Helper()
	data, err := store.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if err := sys.WriteSettings(data); err != nil {
		t.Fatal(err)
	}
	a := &App{settings: s.Clone(), saved: s.Clone(), savedData: data}
	a.file.markSeen(data)
	return a, data
}

// TestPollAppliesExternalChangeOnce covers the watcher taking a file another
// program wrote, and not waking the window again while it stays the same.
func TestPollAppliesExternalChangeOnce(t *testing.T) {
	settingsDir(t)
	base := store.Default()
	a, _ := settingsApp(t, base)

	next := base.Clone()
	next.Theme = "light"
	nextData, err := store.Marshal(next)
	if err != nil {
		t.Fatal(err)
	}
	if err := sys.WriteSettings(nextData); err != nil {
		t.Fatal(err)
	}

	calls := 0
	dispatch := func(fn func()) {
		calls++
		fn()
	}
	a.poll(dispatch)
	if calls != 1 {
		t.Fatalf("dispatched %d times, want one", calls)
	}
	if a.settings.Theme != "light" {
		t.Errorf("theme = %q, want the file's light", a.settings.Theme)
	}
	if !bytes.Equal(a.savedData, nextData) {
		t.Error("savedData is not the file the watcher took")
	}

	a.poll(dispatch)
	if calls != 1 {
		t.Errorf("dispatched %d times for an unchanged file, want one", calls)
	}
}

// TestSaveRefusesToOverwriteExternalChange covers a save that would lose a
// valid change another program made after the window read the file.
func TestSaveRefusesToOverwriteExternalChange(t *testing.T) {
	settingsDir(t)
	base := store.Default()
	a, baseData := settingsApp(t, base)

	next := base.Clone()
	next.Language = "en"
	nextData, err := store.Marshal(next)
	if err != nil {
		t.Fatal(err)
	}
	if err := sys.WriteSettings(nextData); err != nil {
		t.Fatal(err)
	}

	a.settings.Font.Size = 14
	a.SaveSettings()

	onDisk, ok, err := sys.ReadSettings()
	if err != nil || !ok {
		t.Fatalf("read: %v, %v", err, ok)
	}
	if !bytes.Equal(onDisk, nextData) {
		t.Errorf("file = %q, want the external change", onDisk)
	}
	if !bytes.Equal(a.savedData, baseData) {
		t.Error("savedData changed for a save that wrote nothing")
	}
	if !strings.Contains(a.settingsNote(), "settings.json") {
		t.Errorf("note = %q, want the conflict", a.settingsNote())
	}
}

// TestDirtyDraftKeepsExternalChange covers a draft in the way of a valid
// external change: the draft stays, saving is refused, and discarding takes
// the file.
func TestDirtyDraftKeepsExternalChange(t *testing.T) {
	settingsDir(t)
	base := store.Default()
	a, baseData := settingsApp(t, base)
	a.settings.Theme = "light"

	next := base.Clone()
	next.Font.Size = 14
	nextData, err := store.Marshal(next)
	if err != nil {
		t.Fatal(err)
	}
	if err := sys.WriteSettings(nextData); err != nil {
		t.Fatal(err)
	}

	calls := 0
	a.poll(func(fn func()) {
		calls++
		fn()
	})
	if calls != 1 {
		t.Fatalf("dispatched %d times, want one", calls)
	}
	if a.settings.Font.Size == 14 {
		t.Error("the draft took the external change")
	}
	if a.settings.Theme != "light" {
		t.Error("the draft changed")
	}
	if !bytes.Equal(a.savedData, baseData) {
		t.Error("savedData changed while the draft was in the way")
	}
	if !strings.Contains(a.settingsNote(), "settings.json") {
		t.Errorf("note = %q, want the conflict", a.settingsNote())
	}

	a.SaveSettings()
	onDisk, ok, err := sys.ReadSettings()
	if err != nil || !ok {
		t.Fatalf("read: %v, %v", err, ok)
	}
	if !bytes.Equal(onDisk, nextData) {
		t.Errorf("file = %q, want the external change", onDisk)
	}

	a.DiscardSettings()
	if a.settings.Font.Size != 14 {
		t.Errorf("font size = %v, want the file's 14", a.settings.Font.Size)
	}
	if !bytes.Equal(a.savedData, nextData) {
		t.Error("savedData is not the file discarding took")
	}
}

// TestBadFileNoteSurvivesPageSwitchAndDiscard covers the note about a file
// that is not settings: it is why the page shows defaults, so changing pages
// or discarding a draft must not hide it.
func TestBadFileNoteSurvivesPageSwitchAndDiscard(t *testing.T) {
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

	a.settingsTab = &pane.Tab{Name: "Settings"}
	a.SelectSettingsPage(ui.PageGeneral)
	if !strings.Contains(a.settingsNote(), "settings.json") {
		t.Errorf("note after a page switch = %q, want the bad file", a.settingsNote())
	}

	a.DiscardSettings()
	if !strings.Contains(a.settingsNote(), "settings.json") {
		t.Errorf("note after discard = %q, want the bad file", a.settingsNote())
	}
}

// TestReloadKeepsRunningSettingsForOutOfRangeFile covers the whole-file
// decision: a file with one setting outside its range is not half applied, and
// the window keeps the settings it was already running.
func TestReloadKeepsRunningSettingsForOutOfRangeFile(t *testing.T) {
	settingsDir(t)
	base := store.Default()
	base.Font.Size = 14
	a := &App{settings: base.Clone(), saved: base.Clone(), savedData: []byte("old")}
	bad := []byte(`{"language":"en","font":{"size":0}}`)
	a.reload(bad)
	if a.settings.Font.Size != 14 {
		t.Errorf("font size = %v, want the running 14", a.settings.Font.Size)
	}
	if a.settings.Language != "" {
		t.Errorf("language = %q, want the running empty one", a.settings.Language)
	}
	if !strings.Contains(a.configErr, "font.size") {
		t.Errorf("configErr = %q, want it to name font.size", a.configErr)
	}
	if !bytes.Equal(a.badData, bad) {
		t.Error("badData is not the rejected file")
	}
}

// TestPollWaitsForBadBytesToSettle covers an editor still writing: changing
// bad bytes are not kept, and the same bad bytes seen twice are kept once.
func TestPollWaitsForBadBytesToSettle(t *testing.T) {
	settingsDir(t)
	a, _ := settingsApp(t, store.Default())
	calls := 0
	dispatch := func(fn func()) {
		calls++
		fn()
	}

	first := []byte(`{"font":`)
	if err := sys.WriteSettings(first); err != nil {
		t.Fatal(err)
	}
	a.poll(dispatch)
	if calls != 0 {
		t.Fatalf("dispatched %d times for a half-written file, want none", calls)
	}

	second := []byte(`{"font":{"size":`)
	if err := sys.WriteSettings(second); err != nil {
		t.Fatal(err)
	}
	a.poll(dispatch)
	if calls != 0 {
		t.Fatalf("dispatched %d times while the bytes changed, want none", calls)
	}

	if err := sys.WriteSettings(second); err != nil {
		t.Fatal(err)
	}
	a.poll(dispatch)
	if calls != 1 {
		t.Fatalf("dispatched %d times for settled bad bytes, want one", calls)
	}

	dir, err := sys.Dir()
	if err != nil {
		t.Fatal(err)
	}
	archives, err := filepath.Glob(filepath.Join(dir, "settings.json.bad-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(archives) != 1 {
		t.Fatalf("archives = %v, want one", archives)
	}

	a.poll(dispatch)
	if calls != 1 {
		t.Errorf("dispatched %d times for unchanged bad bytes, want one", calls)
	}
	archives, err = filepath.Glob(filepath.Join(dir, "settings.json.bad-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(archives) != 1 {
		t.Errorf("archives = %v, want one after an unchanged read", archives)
	}
}

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

// TestPollShowsAReadErrorOnce covers a file that cannot be read: the read
// error is shown once, an unchanged error does not wake the window again, and
// the note goes when the file can be read again.
func TestPollShowsAReadErrorOnce(t *testing.T) {
	dir := settingsDir(t)
	a, data := settingsApp(t, store.Default())
	path, err := sys.SettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	// A directory where the file should be: ReadFile answers "is a directory"
	// on every platform, which is the read error the watcher has to survive.
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(filepath.Join(dir, "wt-go")) })

	calls := 0
	dispatch := func(fn func()) {
		calls++
		fn()
	}
	a.poll(dispatch)
	if calls != 1 {
		t.Fatalf("dispatched %d times for a read error, want one", calls)
	}
	if a.note == "" {
		t.Fatal("the read error is not shown")
	}
	note := a.note

	a.poll(dispatch)
	if calls != 1 {
		t.Errorf("dispatched %d times for the same read error, want one", calls)
	}
	if a.note != note {
		t.Errorf("note = %q, want the one already shown %q", a.note, note)
	}

	// The file is readable again, holding what the window already shows: the
	// note goes, and the content is not taken a second time.
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := sys.WriteSettings(data); err != nil {
		t.Fatal(err)
	}
	a.poll(dispatch)
	if calls != 2 {
		t.Fatalf("dispatched %d times for a file that reads again, want two", calls)
	}
	if a.note != "" {
		t.Errorf("note = %q, want none once the file reads", a.note)
	}
	if a.configErr != "" || a.badData != nil {
		t.Errorf("configErr = %q badData = %q, want none", a.configErr, a.badData)
	}

	a.poll(dispatch)
	if calls != 2 {
		t.Errorf("dispatched %d times for an unchanged readable file, want two", calls)
	}
}

// TestPollTakesAnEmptiedFileAsBad covers a file that was not settings and was
// then emptied rather than deleted: emptied is a file with no settings in it,
// so the reason names it, and the settings the window runs stay.
func TestPollTakesAnEmptiedFileAsBad(t *testing.T) {
	settingsDir(t)
	bad := []byte("{oops")
	if err := sys.WriteSettings(bad); err != nil {
		t.Fatal(err)
	}
	a := &App{settings: store.Default(), saved: store.Default(), savedData: []byte("{}")}
	a.reload(bad)
	if a.configErr == "" || a.badData == nil {
		t.Fatal("the bad file was not reported")
	}
	running := a.settings.Clone()

	if err := sys.WriteSettings([]byte{}); err != nil {
		t.Fatal(err)
	}

	calls := 0
	dispatch := func(fn func()) {
		calls++
		fn()
	}
	a.poll(dispatch)
	if calls != 1 {
		t.Fatalf("dispatched %d times for an emptied file, want one", calls)
	}
	if a.configErr == "" {
		t.Error("an emptied file is not reported")
	}
	if !store.Equal(a.settings, running) {
		t.Errorf("settings = %+v, want the ones the window runs", a.settings)
	}
	a.poll(dispatch)
	if calls != 1 {
		t.Errorf("dispatched %d times for an emptied file that stays, want one", calls)
	}
}

// TestReloadDoesNotRepeatTheFileItShows covers the watcher being handed the
// bytes the window already runs on: nothing is taken again, the settings page
// is not woken, and running it twice changes nothing.
func TestReloadDoesNotRepeatTheFileItShows(t *testing.T) {
	dir := settingsDir(t)
	s := store.Default()
	s.Theme = "light"
	a, data := settingsApp(t, s)
	a.note = locales.SettingsChangedNote()

	a.reload(data)
	if a.note != "" {
		t.Errorf("note = %q, want the note about a change cleared", a.note)
	}
	if !bytes.Equal(a.savedData, data) {
		t.Error("savedData changed for bytes the window already runs on")
	}
	a.reload(data)
	if a.note != "" || a.configErr != "" {
		t.Errorf("note = %q configErr = %q, want neither on a second pass", a.note, a.configErr)
	}
	archives, err := filepath.Glob(filepath.Join(dir, "wt-go", "settings.json.bad-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(archives) != 0 {
		t.Errorf("archives = %v, want none: the bytes are the settings it shows", archives)
	}
}

// TestPollSkipsTheBytesItAlreadyTook covers the watcher reading the file the
// window wrote itself: the content is what it shows, so nothing is dispatched.
func TestPollSkipsTheBytesItAlreadyTook(t *testing.T) {
	settingsDir(t)
	a, _ := settingsApp(t, store.Default())
	calls := 0
	a.poll(func(fn func()) {
		calls++
		fn()
	})
	if calls != 0 {
		t.Errorf("dispatched %d times for the file it already took, want none", calls)
	}
}

// TestPollWaitsForASaveOvertakenByAnEdit covers reload being handed bytes the
// disk no longer holds: they are not recorded as seen, so the next read takes
// the newer file instead of skipping it as already shown.
func TestPollWaitsForASaveOvertakenByAnEdit(t *testing.T) {
	settingsDir(t)
	a, _ := settingsApp(t, store.Default())

	read := []byte(`{"theme":"light"}`)
	onDisk := []byte(`{"theme":"dark"}`)
	if err := sys.WriteSettings(onDisk); err != nil {
		t.Fatal(err)
	}
	a.reload(read)
	if !a.file.changed(onDisk) {
		t.Error("the file the disk holds was recorded as seen")
	}

	calls := 0
	a.poll(func(fn func()) {
		calls++
		fn()
	})
	if calls != 1 {
		t.Fatalf("dispatched %d times for the newer file, want one", calls)
	}
	if a.settings.Theme != "dark" {
		t.Errorf("theme = %q, want the file's dark", a.settings.Theme)
	}
}

// TestWatchTakesAnExternalChange covers the goroutine itself: the ticker is
// what picks the change up, and dispatch is how it reaches the window.
func TestWatchTakesAnExternalChange(t *testing.T) {
	if testing.Short() {
		t.Skip("waits for the watcher's tick")
	}
	settingsDir(t)
	base := store.Default()
	a, _ := settingsApp(t, base)

	next := base.Clone()
	next.Theme = "light"
	nextData, err := store.Marshal(next)
	if err != nil {
		t.Fatal(err)
	}
	if err := sys.WriteSettings(nextData); err != nil {
		t.Fatal(err)
	}

	// The window is off the UI thread here: dispatch runs the change the
	// watcher hands over, which is what the ticker's the tick reaches.
	done := make(chan struct{}, 4)
	var once sync.Once
	dispatch := func(fn func()) {
		fn()
		once.Do(func() { done <- struct{}{} })
	}
	stop := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		for {
			select {
			case <-stop:
				return
			case <-time.After(10 * time.Millisecond):
				a.poll(dispatch)
			}
		}
	}()
	defer func() {
		close(stop)
		<-finished
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the watcher did not take the change")
	}
	if a.settings.Theme != "light" {
		t.Errorf("theme = %q, want the file's light", a.settings.Theme)
	}
}

// TestFileStateTracksTheWatcher covers the record the watcher keeps without a
// window: content is seen only when said so, bad content settles after the
// same bytes twice, and a file that is gone leaves nothing behind.
func TestFileStateTracksTheWatcher(t *testing.T) {
	var f fileState
	if f.hasAny() {
		t.Error("a fresh state has something to drop")
	}
	if f.changed(nil) {
		t.Error("nothing was handed to the window, so nothing is unchanged")
	}

	data := []byte(`{"theme":"dark"}`)
	f.markSeen(data)
	if f.changed(data) {
		t.Error("the content it handed to the window counts as changed")
	}
	if !f.changed([]byte(`{"theme":"light"}`)) {
		t.Error("other content counts as unchanged")
	}
	if !f.hasAny() {
		t.Error("content it handed to the window is not remembered")
	}

	// Bad content is kept once it has been read the same way twice.
	bad := []byte(`{"font":`)
	if f.settledBad(bad) {
		t.Error("bad content settled on its first read")
	}
	if !f.settledBad(bad) {
		t.Error("the same bad content did not settle")
	}
	f.clearBad()
	if f.settledBad(bad) {
		t.Error("cleared bad content settled again")
	}

	// A reason for the file is something to drop until it is cleared.
	f.markReason(bad)
	if !f.hasAny() {
		t.Error("a reason is not remembered")
	}
	f.clearReason()
	f.forgetSeen()
	if f.hasAny() {
		t.Error("forgetting the file kept something about it")
	}
	take := []byte("x")
	f.markSeen(take)
	take[0] = 'y'
	if f.changed([]byte("x")) {
		t.Error("the state shares its bytes with the caller's slice")
	}
}

// TestFileStateReadErrorIsNewOnce covers the read error: the same error is not
// new a second time, a different one is, and clearing answers with the note
// the window showed.
func TestFileStateReadErrorIsNewOnce(t *testing.T) {
	var f fileState
	if _, had := f.clearReadError(); had {
		t.Error("a fresh state had a read error to clear")
	}
	if !f.setReadError("first") {
		t.Error("the first read error is not new")
	}
	if f.setReadError("first") {
		t.Error("the same read error is new twice")
	}
	if !f.setReadError("second") {
		t.Error("another read error is not new")
	}
	if !f.hasAny() {
		t.Error("a read error is not remembered")
	}
	note, had := f.clearReadError()
	if !had || note != "second" {
		t.Errorf("clearReadError = %q, %v, want the error it showed", note, had)
	}
	if f.hasAny() {
		t.Error("the cleared read error is still remembered")
	}
	if f.setReadError("") {
		t.Error("no error at all counts as a new one")
	}
	if _, had := f.clearReadError(); had {
		t.Error("clearing no error answered with one")
	}
}

// TestWatchTurnsAPanicInTheWindowIntoANote covers the boundary the watcher
// calls through: dispatch itself panics, and poll neither lets it out nor stops
// reading the file.
func TestWatchTurnsAPanicInTheWindowIntoANote(t *testing.T) {
	settingsDir(t)
	a, _ := settingsApp(t, store.Default())

	next := store.Default()
	next.Theme = "light"
	nextData, err := store.Marshal(next)
	if err != nil {
		t.Fatal(err)
	}
	if err := sys.WriteSettings(nextData); err != nil {
		t.Fatal(err)
	}
	a.poll(func(fn func()) { panic("window is gone") })
	a.reload(nextData)
	if a.settings.Theme != "light" {
		t.Errorf("theme = %q, want the file's light", a.settings.Theme)
	}
}

// TestWatchStops covers the watcher ending when the window closes, rather than
// living on after it.
func TestWatchStops(t *testing.T) {
	a := &App{}
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		a.watch(stop)
		close(done)
	}()
	close(stop)
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("the watcher is still running")
	}
}

// TestSetKeybinding covers recording a key: it replaces the one the action
// had, binds one the defaults do not, and removing the action drops every
// keybinding it has.
func TestSetKeybinding(t *testing.T) {
	settingsDir(t)
	a, _ := settingsApp(t, store.Default())
	n := len(a.settings.Keybindings)

	a.RecordKeybinding(action.TabClose)
	if a.recording != action.TabClose {
		t.Fatalf("recording = %q, want %q", a.recording, action.TabClose)
	}
	a.RecordKeybinding(action.TabClose)
	if a.recording != "" {
		t.Fatalf("recording = %q, want none", a.recording)
	}

	a.SetKeybinding(action.TabClose, "ctrl+alt+w")
	if got := keyOf(a.settings.Keybindings, action.TabClose); got != "ctrl+alt+w" {
		t.Errorf("key of %q = %q, want %q", action.TabClose, got, "ctrl+alt+w")
	}
	if len(a.settings.Keybindings) != n+1 {
		t.Errorf("the settings hold %d keybindings, want %d", len(a.settings.Keybindings), n+1)
	}
	if a.recording != "" {
		t.Errorf("recording = %q, want none", a.recording)
	}

	a.SetKeybinding(action.TabClose, "ctrl+alt+e")
	if got := keyOf(a.settings.Keybindings, action.TabClose); got != "ctrl+alt+e" {
		t.Errorf("key of %q = %q, want %q", action.TabClose, got, "ctrl+alt+e")
	}
	if len(a.settings.Keybindings) != n+1 {
		t.Errorf("the settings hold %d keybindings, want %d", len(a.settings.Keybindings), n+1)
	}

	a.RemoveKeybinding(action.TabClose)
	if got := keyOf(a.settings.Keybindings, action.TabClose); got != "" {
		t.Errorf("key of %q = %q, want none", action.TabClose, got)
	}
	if len(a.settings.Keybindings) != n {
		t.Errorf("the settings hold %d keybindings, want the %d they started with", len(a.settings.Keybindings), n)
	}

	// An action the defaults bind twice loses both keys.
	a.RemoveKeybinding(action.TabNew)
	for _, kb := range a.settings.Keybindings {
		if kb.Action == action.TabNew {
			t.Errorf("%q still runs on %q", action.TabNew, kb.Key)
		}
	}
}

// TestResetKeybinding covers going back to the catalog's keys for one action:
// the keys it starts on come back where the file had it, the key the file gave
// it goes, and the actions around it stay where they were.
func TestResetKeybinding(t *testing.T) {
	settingsDir(t)
	a, _ := settingsApp(t, store.Default())
	a.SetKeybinding(action.TabNew, "f2")
	a.SetKeybinding(action.TabClose, "f3")
	a.ResetKeybinding(action.TabNew)

	var keys []string
	for _, kb := range a.settings.Keybindings {
		if kb.Action == action.TabNew {
			keys = append(keys, kb.Key)
		}
	}
	want := store.DefaultKeybindings()
	var wants []string
	for _, kb := range want {
		if kb.Action == action.TabNew {
			wants = append(wants, kb.Key)
		}
	}
	if len(keys) != len(wants) {
		t.Fatalf("%q runs on %v, want its defaults %v", action.TabNew, keys, wants)
	}
	for i := range wants {
		if keys[i] != wants[i] {
			t.Fatalf("%q runs on %v, want its defaults %v", action.TabNew, keys, wants)
		}
	}
	if got := keyOf(a.settings.Keybindings, action.TabClose); got != "f3" {
		t.Errorf("key of %q = %q, want f3: resetting one action left another alone", action.TabClose, got)
	}
}

// TestSetKeybindingDropsTheOtherKeys covers an action the defaults bind twice:
// recording a key leaves it running on that key alone, so that the key it keep
// invisibly does not still run it.
func TestSetKeybindingDropsTheOtherKeys(t *testing.T) {
	settingsDir(t)
	a, _ := settingsApp(t, store.Default())
	a.SetKeybinding(action.TabNew, "f2")
	var keys []string
	for _, kb := range a.settings.Keybindings {
		if kb.Action == action.TabNew {
			keys = append(keys, kb.Key)
		}
	}
	if len(keys) != 1 || keys[0] != "f2" {
		t.Fatalf("%q runs on %v, want f2 alone", action.TabNew, keys)
	}
}

// TestKnownProfilesSkipsHidden covers the list the settings offer: a hidden
// profile runs the system's shell, so the page does not offer it.
func TestKnownProfilesSkipsHidden(t *testing.T) {
	ps := []profile.Profile{{ID: "a", Name: "A"}, {ID: "b", Name: "B", Hidden: true}}
	got := knownProfiles(ps)
	if len(got) != 1 || got[0].ID != "a" {
		t.Fatalf("knownProfiles = %+v, want A alone", got)
	}
}

// TestSchemeBaseName covers the name a copy starts from: the marks earlier
// copies left on it come off, and a name that only looks like one keeps its
// own.
func TestSchemeBaseName(t *testing.T) {
	tests := []struct{ name, want string }{
		{"Campbell", "Campbell"},
		{"Campbell (custom)", "Campbell"},
		{"Campbell (custom 2)", "Campbell"},
		{"Campbell (custom 12)", "Campbell"},
		{"Campbell (custom) (custom)", "Campbell"},
		{"Campbell (custom) (custom 3)", "Campbell"},
		{"Campbell (custom) (custom) (custom)", "Campbell"},
		{"(custom)", "(custom)"},
		{"Paper (custom look)", "Paper (custom look)"},
		{"Paper (custom x)", "Paper (custom x)"},
	}
	for _, tt := range tests {
		if got := schemeBaseName(tt.name); got != tt.want {
			t.Errorf("schemeBaseName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

// TestCopySchemeNamesCopies covers a copy: it is named after the scheme behind
// the copy, so copying copies does not grow a chain of marks, and a name the
// settings already write is numbered.
func TestCopySchemeNamesCopies(t *testing.T) {
	settingsDir(t)
	s := store.Default()
	s.Schemes = []store.Scheme{own("Campbell (custom)")}
	a, _ := settingsApp(t, s)
	a.settings.ColorScheme = "Campbell"

	a.CopyScheme()
	want := "Campbell (custom 2)"
	if a.settings.ColorScheme != want {
		t.Fatalf("copy = %q, want %q", a.settings.ColorScheme, want)
	}
	if !a.schemeNamed(want) {
		t.Errorf("the settings do not write %q", want)
	}

	a.CopyScheme()
	want = "Campbell (custom 3)"
	if a.settings.ColorScheme != want {
		t.Fatalf("copy of a copy = %q, want %q", a.settings.ColorScheme, want)
	}
	if len(a.settings.Schemes) != 3 {
		t.Errorf("the settings write %d schemes, want three", len(a.settings.Schemes))
	}
}

// TestRemoveScheme covers removing a scheme the settings write: it goes, the
// terminal takes the scheme it was a copy of, the defaults' when that is gone
// too, and a name the settings do not write asks for nothing.
func TestRemoveScheme(t *testing.T) {
	settingsDir(t)
	s := store.Default()
	s.Schemes = []store.Scheme{own("Paper"), own("Paper (custom)")}
	s.ColorScheme = "Paper (custom)"
	a, _ := settingsApp(t, s)

	a.RemoveScheme("Dracula")
	if a.note != "" {
		t.Errorf("note = %q, want none for a scheme the settings do not write", a.note)
	}
	if !a.schemeNamed("Paper") || !a.schemeNamed("Paper (custom)") {
		t.Fatal("a scheme the settings do not write took one of theirs")
	}

	a.RemoveScheme("Paper (custom)")
	if a.schemeNamed("Paper (custom)") {
		t.Error("the scheme is still there")
	}
	if a.settings.ColorScheme != "Paper" {
		t.Errorf("scheme = %q, want the %q it copied", a.settings.ColorScheme, "Paper")
	}
	if a.note == "" {
		t.Error("no note about the removal")
	}

	a.RemoveScheme("Paper")
	if want := store.Default().ColorScheme; a.settings.ColorScheme != want {
		t.Errorf("scheme with every copy gone = %q, want %q", a.settings.ColorScheme, want)
	}
}

// own is a scheme the settings write, with the few colors the theme needs.
func own(name string) store.Scheme {
	return store.Scheme{
		Name:       name,
		Background: "#101010",
		Foreground: "#eeeeee",
		Palette:    []string{"#000000"},
	}
}

// TestAddProfileDropsTheArgumentsOfTheLastOne covers a new profile opening on
// the page the profile before it had: the lines it held are not this profile's
// arguments, so the page starts empty and nothing is unsaved.
func TestAddProfileDropsTheArgumentsOfTheLastOne(t *testing.T) {
	settingsDir(t)
	s := store.Default()
	s.Profiles = []store.Profile{{ID: "p1", Name: "One", Command: "pwsh", Args: []string{"-NoLogo", "-NoProfile"}}, {ID: "p2", Name: "Two"}}
	a, _ := settingsApp(t, s)
	a.settingsTab = &pane.Tab{Name: "Settings"}

	a.SelectSettingsPage(ui.ProfilePage("p1"))
	if a.editingArgs != "-NoLogo\n-NoProfile" {
		t.Fatalf("editingArgs = %q, want the profile's lines", a.editingArgs)
	}
	a.editingArgs = "-NoLogo\n-NoProfile\n-Draft"
	if !a.dirty() {
		t.Fatal("a draft over the file is not dirty")
	}

	a.AddProfile()
	if a.editingArgs != "" {
		t.Errorf("editingArgs = %q, want none: the last profile's lines are not this one's", a.editingArgs)
	}
	// The profile itself is new, which is unsaved; what the page holds is not.
	a.settings = a.saved.Clone()
	if a.dirty() {
		t.Error("the page holds a line the new profile does not")
	}
}

// TestDuplicateProfileShowsTheCopiedArguments covers a copy opening on the
// page: it shows the arguments it copied, not the ones the page held before.
func TestDuplicateProfileShowsTheCopiedArguments(t *testing.T) {
	settingsDir(t)
	s := store.Default()
	s.Profiles = []store.Profile{
		{ID: "p1", Name: "One", Command: "pwsh", Args: []string{"-NoLogo"}},
		{ID: "p2", Name: "Two", Command: "bash", Args: []string{"-l", "-i"}},
		{ID: "p3", Name: "Three"},
	}
	a, _ := settingsApp(t, s)

	a.SelectSettingsPage(ui.ProfilePage("p1"))

	a.DuplicateProfile("p2")
	if want := "-l\n-i"; a.editingArgs != want {
		t.Errorf("editingArgs = %q, want %q", a.editingArgs, want)
	}
	// The page opened is the copy's, not the one it was on.
	if a.settingsPage != ui.ProfilePage(a.editing.ID) {
		t.Errorf("page = %q, want the copy's %q", a.settingsPage, ui.ProfilePage(a.editing.ID))
	}
	if a.editing.Name == "Two" {
		t.Error("the page edits the profile itself, not a copy")
	}
	// The copy itself is new, which is unsaved; what the page holds is not.
	a.settings = a.saved.Clone()
	if a.dirty() {
		t.Error("the page holds a line the copy does not")
	}
}

// TestDiscardSettingsLeavesAnUnsavedProfile covers discarding a draft with a
// profile the file does not hold: the profile goes, and the page it was on
// goes with it, rather than editing nothing.
func TestDiscardSettingsLeavesAnUnsavedProfile(t *testing.T) {
	settingsDir(t)
	a, _ := settingsApp(t, store.Default())
	before := len(a.settings.Profiles)

	a.AddProfile()
	if a.editing == nil {
		t.Fatal("the new profile is not open on the page")
	}

	a.DiscardSettings()
	if a.editing != nil {
		t.Error("a profile the file does not hold is still open")
	}
	if a.settingsPage != ui.PageProfiles {
		t.Errorf("page = %q, want %q", a.settingsPage, ui.PageProfiles)
	}
	if len(a.settings.Profiles) != before {
		t.Errorf("the settings hold %d profiles, want the file's %d", len(a.settings.Profiles), before)
	}
	if a.dirty() {
		t.Error("discarding left a draft")
	}
}

// keyOf is the key an action runs on, "" when it runs on none.
func keyOf(all []store.Keybinding, action string) string {
	for _, kb := range all {
		if kb.Action == action {
			return kb.Key
		}
	}
	return ""
}

// The page asks for the schemes once per settings revision: building them
// sorts 617 names, which a frame should not do twice.
func TestSchemeListIsBuiltOnce(t *testing.T) {
	a := &App{settings: store.Default()}
	first := a.schemeList()
	if len(first) == 0 {
		t.Fatal("no schemes")
	}
	if second := a.schemeList(); &second[0] != &first[0] {
		t.Fatal("the list was rebuilt for the same settings")
	}
}

// The schemes the settings page offers come from the theme package, which
// sorts the 617 names it is built with; once per frame is a cost the page
// should not pay, so this pins what one call buys.
func BenchmarkUISchemes(b *testing.B) {
	s := store.Default()
	b.ReportAllocs()
	for b.Loop() {
		uiSchemes(s)
	}
}

// TestProfileIcon covers the mark the profile page picks: a shape is written
// to the file as it is, a character typed in the box is written as the one
// character it is, and an emptied box is a profile that wears no character of
// its own. The page edits a copy of the profile, so the file changes on save.
func TestProfileIcon(t *testing.T) {
	settingsDir(t)
	a := newApp("")
	a.settings.Profiles = []store.Profile{{
		ID:         "pwsh7",
		Name:       "PowerShell 7",
		Command:    `C:\Windows\pwsh.exe`,
		Discovered: true,
	}}
	a.profiles = profile.Resolve(a.settings.Profiles, a.found)
	a.editProfile("pwsh7")

	a.SetProfileIcon("bash:arch")
	if got := a.editing.Icon; got != "bash:arch" {
		t.Errorf("icon = %q, want the shape the page picked", got)
	}
	if a.editingGlyph != "" {
		t.Errorf("the character box holds %q, want it cleared by a shape", a.editingGlyph)
	}
	if !a.dirty() {
		t.Error("the page does not know it changed the mark")
	}
	a.SaveSettings()
	if got := a.settings.Profiles[0].Icon; got != "bash:arch" {
		t.Errorf("saved icon = %q, want the shape the page picked", got)
	}

	a.editProfile("pwsh7")
	if a.editingGlyph != "" {
		t.Error("the character box opens on text, want it empty for a shape")
	}
	a.editingGlyph = "\U0001F92C"
	a.SaveSettings()
	if got := a.settings.Profiles[0].Icon; got != "\U0001F92C" {
		t.Errorf("saved icon = %q, want the character the page typed", got)
	}

	a.editProfile("pwsh7")
	if a.editingGlyph != "\U0001F92C" {
		t.Errorf("the box opens on %q, want the character the file holds", a.editingGlyph)
	}
	a.editingGlyph = "\U0001F92C\U0001F92C"
	a.SaveSettings()
	if got := a.settings.Profiles[0].Icon; got != "\U0001F92C" {
		t.Errorf("saved icon = %q, want the one character of the two typed", got)
	}

	// An emptied box is a profile that wears no character: it keeps a shape it
	// holds, and goes back to its command's mark otherwise.
	a.SetProfileIcon("cmd:matrix")
	a.SaveSettings()
	if got := a.settings.Profiles[0].Icon; got != "cmd:matrix" {
		t.Errorf("saved icon = %q, want the shape the page picked", got)
	}
	a.editProfile("pwsh7")
	a.editingGlyph = ""
	a.SaveSettings()
	if got := a.settings.Profiles[0].Icon; got != "cmd:matrix" {
		t.Errorf("saved icon = %q, want the shape an empty box keeps", got)
	}
	a.SetProfileIcon("")
	a.SaveSettings()
	if got := a.settings.Profiles[0].Icon; got != "" {
		t.Errorf("saved icon = %q, want the command's mark", got)
	}
}
