package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"wtgo/internal/action"
	"wtgo/internal/pane"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/sys"
	"wtgo/internal/ui"
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
	a.SelectSettingsPage(ui.PageStartup)
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
