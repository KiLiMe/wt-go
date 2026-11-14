package app

import (
	"sync"
	"testing"
	"time"
	"wtgo/internal/store"
	"wtgo/internal/sys"
)

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
