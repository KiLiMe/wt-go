package app

import "wtgo/internal/ui"

// overlayKind says which of the window's overlays is open, or that none is:
// the command palette, the bar that asks the terminal, the dialog that renames
// a tab, the question before something the window cannot take back, or the
// dialog that says what the program is.
//
// The window shows one of them at a time. Whatever opens goes through
// openOverlay, which closes the one before it, so the kind of the one open is
// enough to say which, and the keys go to that one alone.
type overlayKind int

const (
	overlayNone overlayKind = iota
	// overlayRename is the dialog that renames a tab.
	overlayRename
	// overlayConfirm is the question the window asks before it does
	// something it cannot take back.
	overlayConfirm
	// overlayAbout is the dialog that says what the program is.
	overlayAbout
	// overlayPalette is the command palette.
	overlayPalette
	// overlayFind is the bar that asks the terminal for text. It is not
	// modal, so it does not take the keyboard away from the pane: it takes
	// the keys the pane would have taken, which is what makes Escape close
	// the bar rather than reach the program in front.
	overlayFind
)

// top is the kind of the overlay the window shows, overlayNone while it shows
// none. The overlays are exclusive, so there is exactly one, and the keys go to
// it alone.
func (a *App) top() overlayKind {
	switch {
	case a.palette != nil:
		return overlayPalette
	case a.find != nil:
		return overlayFind
	case a.renaming != nil:
		return overlayRename
	case a.about != nil:
		return overlayAbout
	case a.confirm != nil:
		return overlayConfirm
	}
	return overlayNone
}

// openOverlay shows the overlay top, closing the one before it: a window that
// asked a question and then opened the palette would otherwise have two
// dialogs and no way to say which one a key answers. What the dialog holds
// stays with the caller, which sets it right after. The same goes for a
// question: it comes with its answers, and ask binds it over what openOverlay
// closes, so openOverlay does not open one.
func (a *App) openOverlay(top overlayKind) {
	a.closeOpen()
	switch top {
	case overlayRename:
		a.renaming = &ui.Rename{Open: true}
	case overlayPalette:
		a.palette = &ui.Palette{Open: true}
	case overlayAbout:
		a.about = &ui.About{Open: true}
	case overlayFind:
		a.find = &ui.Find{}
	}
}

// closeOverlay closes the overlay kind, and closes nothing while that one is not
// the one showing: a dialog's own callback can arrive after another overlay
// took its place, and must not close the one that did.
func (a *App) closeOverlay(kind overlayKind) {
	if a.top() == kind {
		a.closeOpen()
	}
}

// closeOpen closes the overlay the window shows, which is what opening another
// does before it shows.
func (a *App) closeOpen() {
	if a.find != nil {
		// The search the bar started goes with the bar, whatever shows in
		// its place: a terminal left highlighting matches is one the next
		// bar opens on without being asked.
		a.endSearch()
	}
	switch a.top() {
	case overlayPalette:
		a.palette = nil
	case overlayFind:
		a.find = nil
	case overlayRename:
		a.renaming = nil
	case overlayAbout:
		a.about = nil
	case overlayConfirm:
		a.confirm = nil
	}
}

// closeOverlays closes every overlay of the window, which is what going to
// another view of it does: the settings tab is not a terminal, so a bar that
// asks the terminal in front for text has nothing left to ask once it shows.
func (a *App) closeOverlays() {
	a.closeOpen()
}

// dismissed is whether the frame closed one of the overlays it was drawing,
// which the dialog sets as the user dismisses it. The frame then drops what was
// closed before anything acts, so that a dismissal alone does not open another
// overlay: a closed dialog leaves the keys to the window again, and a key meant
// for the window would go to whatever was under the dialog.
func (a *App) dismissed() bool {
	return a.renaming != nil && !a.renaming.Open ||
		a.confirm != nil && !a.confirm.Open ||
		a.about != nil && !a.about.Open ||
		a.palette != nil && !a.palette.Open
}

// dropClosed forgets the overlays the last frame closed.
func (a *App) dropClosed() {
	c := a.confirm
	switch {
	case a.renaming != nil && !a.renaming.Open:
		a.renaming = nil
	case c != nil && !c.Open:
		a.confirm = nil
	case a.about != nil && !a.about.Open:
		a.about = nil
	case a.palette != nil && !a.palette.Open:
		a.palette = nil
	}
}

// run is the action a key names, as the overlay showing decides it: an overlay
// that holds the keyboard keeps the window's actions from the keys until it
// closes, so the tabs behind a question or a dialog do not move while it is
// there. The bar that asks the terminal only takes the keys the pane would
// have taken, so the window's keys still work over it.
func (a *App) run(id string) {
	switch a.top() {
	case overlayNone, overlayFind:
		a.act(id)
	}
}
