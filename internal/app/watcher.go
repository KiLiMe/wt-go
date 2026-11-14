package app

import (
	"bytes"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"time"

	"wtgo/internal/store"
	"wtgo/internal/sys"
	"wtgo/locales"
)

// watch takes the settings the file holds when another program writes them,
// once a second, as Windows Terminal does, until stop closes.
func (a *App) watch(stop <-chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			a.poll(a.Do)
		}
	}
}

// poll reads the file once and hands a change to dispatch, which runs it on the
// UI thread. It reads the same file without dispatching anything while the
// content is unchanged. A panic in one read is logged and the watcher goes on:
// a bad setting is not a reason to stop watching the rest.
func (a *App) poll(dispatch func(func())) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("wt-go: settings watcher: %v\n%s", r, debug.Stack())
		}
	}()
	data, ok, err := sys.ReadSettings()
	if err != nil {
		note := locales.ErrSettings(err.Error())
		if a.file.setReadError(note) {
			dispatch(func() { a.setNote(note) })
		}
		return
	}
	if note, had := a.file.clearReadError(); had {
		dispatch(func() { a.clearNote(note) })
	}
	if !ok {
		// poll runs off the UI thread, so whether there is anything to clean up
		// is asked of the file state, which has its own lock, rather than read
		// from badData and configErr the UI thread writes.
		if a.file.hasAny() {
			dispatch(func() { a.forgetBadFile() })
		}
		return
	}
	if !a.file.changed(data) {
		a.file.clearBad()
		return
	}
	if _, err := store.Parse(data); err != nil && !a.file.settledBad(data) {
		return
	}
	dispatch(func() { a.reloadSafely(data) })
}

// reload takes the settings the file holds now, when they are settings: a file
// that is not keeps the ones the window shows, and says why. Noting the content
// as seen is the UI thread's, where poll's read may be overtaken by a save: the
// note is only taken when the file still holds the bytes that were read.
func (a *App) reload(data []byte) {
	if a.savedData != nil && bytes.Equal(data, a.savedData) {
		a.noteSeen(data)
		if a.badData != nil {
			a.badData, a.configErr = nil, ""
			a.setNote("")
		}
		if a.note == locales.SettingsChangedNote() {
			a.setNote("")
		}
		return
	}
	if a.badData != nil && bytes.Equal(data, a.badData) {
		a.noteSeen(data)
		return
	}
	if a.badData != nil {
		a.badData, a.configErr = nil, ""
	}
	if a.dirty() {
		a.file.clearReason()
		a.noteSeen(data)
		a.setNote(locales.SettingsChangedNote())
		return
	}
	settings, err := store.Parse(data)
	if err != nil {
		a.badData = data
		a.file.markReason(data)
		a.noteSeen(data)
		a.reject(data, err)
		return
	}
	a.settings, a.saved, a.savedData, a.badData = settings, settings.Clone(), data, nil
	a.configErr = ""
	a.file.clearReason()
	a.noteSeen(data)
	if a.editing != nil {
		a.editProfile(a.editing.ID)
	}
	a.setNote("")
	a.redraw()
}

// reject keeps what the file held that was not settings, and says on the
// settings page why the window is not showing it. The note is kept apart from
// what the last save said, so switching pages or discarding a draft does not
// clear it.
func (a *App) reject(data []byte, err error) {
	note := configError(data, err)
	if note == a.configErr {
		return
	}
	a.note = ""
	a.configErr = note
	a.redraw()
	if a.win != nil && !a.closed && a.settingsTab == nil {
		a.OpenSettings()
	}
}

// reloadSafely runs reload on the UI thread. A panic is logged and shown on the
// settings page instead of escaping: the watcher must keep running, and a
// reload that failed must be visible.
func (a *App) reloadSafely(data []byte) {
	a.safely(func() { a.reload(data) })
}

// safely runs fn, turning a panic into a note. It is the boundary between a
// background read and the window: nothing a reload does may take the watcher
// down with it.
func (a *App) safely(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("wt-go: settings reload: %v\n%s", r, debug.Stack())
			a.setNote(locales.ErrSettings(fmt.Sprint(r)))
		}
	}()
	fn()
}

// fileState is what the watcher knows of settings.json: the content it has
// already handed to the UI, bad content it is waiting to see twice, the read
// error it last showed, and the content the UI said was not settings. The
// mutex keeps the watcher goroutine and the UI thread from racing on it.
type fileState struct {
	mu      sync.Mutex
	seen    []byte
	bad     []byte
	readErr string
	reason  []byte
}

// changed reports whether data differs from the content the watcher last
// handed to the UI. It does not record data: only content that is dispatched
// is seen.
func (f *fileState) changed(data []byte) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return !bytes.Equal(data, f.seen)
}

// markSeen records the content the watcher handed to the UI.
func (f *fileState) markSeen(data []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seen = append([]byte(nil), data...)
	f.bad = nil
}

// forgetSeen forgets the content the watcher handed to the UI, for a file that
// is gone: content written back is a new file, and has to be taken again.
func (f *fileState) forgetSeen() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seen, f.bad = nil, nil
	f.reason = nil
}

// hasAny reports whether there is anything about the file for the UI to drop:
// content, a read error, or a reason the file is not settings. poll asks this
// rather than reading badData and configErr, which only the UI thread writes.
func (f *fileState) hasAny() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.seen != nil || f.readErr != "" || f.reason != nil
}

// markReason records that the settings page is saying the file is not
// settings, and for which content. It is recorded here rather than read from
// the window's own fields so that the UI thread can say a reason is showing
// while the watcher's goroutine is asking.
func (f *fileState) markReason(data []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reason = append([]byte(nil), data...)
}

// clearReason forgets the reason, for a file that is settings again.
func (f *fileState) clearReason() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reason = nil
}

// noteSeen records the content that reload took as seen, but only while the
// file still holds it: the watcher's read may have been overtaken by a save or
// by another edit, and the newer bytes must not be skipped as already shown.
func (a *App) noteSeen(data []byte) {
	onDisk, ok, err := sys.ReadSettings()
	if err != nil || !ok || !bytes.Equal(onDisk, data) {
		return
	}
	a.file.markSeen(data)
}

// clearBad forgets bad content that was seen once but never reported, as when
// an editor restores the file it was writing.
func (f *fileState) clearBad() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.bad = nil
}

// settledBad reports whether data is the same bad content the watcher saw on
// its last read. A file an editor is still writing changes from read to read;
// waiting for the same bytes twice keeps its half-written state from being
// archived once a second.
func (f *fileState) settledBad(data []byte) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !bytes.Equal(data, f.bad) {
		f.bad = append([]byte(nil), data...)
		return false
	}
	return true
}

// setReadError records a read error and reports whether it is new.
func (f *fileState) setReadError(note string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.readErr == note {
		return false
	}
	f.readErr = note
	return true
}

// clearReadError forgets a read error and answers with the note it showed.
func (f *fileState) clearReadError() (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.readErr == "" {
		return "", false
	}
	note := f.readErr
	f.readErr = ""
	return note, true
}
