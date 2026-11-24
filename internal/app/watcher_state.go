package app

import (
	"bytes"
	"sync"

	"wtgo/internal/sys"
)

// Boundary: what the watcher remembers of settings.json -- the content it last
// handed over, the reason it says the file is wrong, and the questions that
// decide whether a read is worth dispatching. The loop is watcher.go.

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
