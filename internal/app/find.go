package app

import (
	"github.com/egoist/mygo/plugins/terminal"

	"wtgo/internal/ui"
)

// OpenFind asks the terminal in front for text.
func (a *App) OpenFind() {
	if a.find != nil {
		return
	}
	if a.focusedTerm() == nil {
		return
	}
	a.find = &ui.Find{}
	a.redraw()
}

// Find searches the terminal the bar bound for text, as the bar is typed in.
func (a *App) Find(text string) {
	if term := a.searchTarget(); term != nil {
		term.Search(text)
	}
}

// FindNext picks the match after the one selected.
func (a *App) FindNext() {
	if term := a.searchTarget(); term != nil {
		term.SearchNext()
	}
}

// FindPrevious picks the match before the one selected.
func (a *App) FindPrevious() {
	if term := a.searchTarget(); term != nil {
		term.SearchPrevious()
	}
}

// EndFind closes the bar, and stops the search that selected matches.
func (a *App) EndFind() {
	if a.find == nil {
		return
	}
	a.find = nil
	a.endSearch()
	a.redraw()
}

// endSearch stops the search the bar started, which may be on a terminal the
// tab in front no longer shows.
func (a *App) endSearch() {
	if a.findTerm == nil {
		return
	}
	a.findTerm.EndSearch()
	a.findTerm = nil
}

// searchTarget is the terminal the bar searches: the one it bound, or the one
// in front while it has bound none. It does not rebind: findState does, so the
// search stays on the tab the bar was opened on.
func (a *App) searchTarget() *terminal.Terminal {
	if a.findTerm != nil {
		return a.findTerm
	}
	return a.focusedTerm()
}

// findState is the find bar as the window draws it, with what the terminal in
// front says of the matches.
func (a *App) findState() *ui.Find {
	if a.find == nil {
		return nil
	}
	term := a.focusedTerm()
	if term == nil {
		// The tab in front is not a terminal, as the settings: the bar
		// has nothing to search, and the search it started goes.
		a.find = nil
		a.endSearch()
		return nil
	}
	if term != a.findTerm {
		// The bar follows the tab in front, and stops the search it
		// started on the tab it left.
		a.endSearch()
		a.findTerm = term
	}
	a.find.Status = term.SearchStatus()
	return a.find
}

// focusedTerm is the terminal the keys go to: the focused pane of the tab in
// front, nil while that tab holds none.
func (a *App) focusedTerm() *terminal.Terminal {
	tab, ok := a.panes.Front()
	if !ok || tab.Focus == nil {
		return nil
	}
	return tab.Focus.Term
}
