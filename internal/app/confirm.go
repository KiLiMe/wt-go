package app

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/transfer"

	"wtgo/internal/i18n"
	"wtgo/internal/pane"
	"wtgo/internal/store"
	"wtgo/internal/ui"
)

// ask is a question the window asks, and what its answers do.
type ask struct {
	ui.Confirm
	yes func()
	no  func()
}

// ask asks a question, doing what its answers say when they come. Dismissing
// the dialog is its no, which the button says.
func (a *App) ask(q ui.Confirm, yes, no func()) {
	q.Open = true
	q.No = i18n.T("Cancel")
	a.confirm = &ask{Confirm: q, yes: yes, no: no}
	a.redraw()
}

// ConfirmYes does what the question asks.
func (a *App) ConfirmYes() {
	q := a.confirm
	a.confirm = nil
	if q != nil && q.yes != nil {
		q.yes()
	}
	a.redraw()
}

// ConfirmNo drops the question.
func (a *App) ConfirmNo() {
	q := a.confirm
	a.confirm = nil
	if q != nil && q.no != nil {
		q.no()
	}
	a.redraw()
}

// askPaste asks whether to paste what holds many lines, and pastes it when the
// answer is yes. It reports whether it took the paste, which keeps the
// terminal from pasting it as it comes.
func (a *App) askPaste(term *terminal.Terminal, d transfer.Data) bool {
	if term == nil || !a.settings.PasteMultilineConfirm {
		return false
	}
	text, err := d.Read(transfer.Text)
	if err != nil {
		return false
	}
	lines := strings.Count(strings.TrimRight(string(text), "\r\n"), "\n") + 1
	if lines < 2 {
		return false
	}
	a.ask(ui.Confirm{
		Title: i18n.T("Paste many lines?"),
		Text:  fmt.Sprintf(i18n.T("This pastes %d lines into the pane."), lines),
		Yes:   i18n.T("Paste"),
	}, func() { term.PasteData(d) }, nil)
	return true
}

// askCloseSettings asks whether to close the settings tab, which drops the
// changes it holds, and closes it when the answer is yes.
func (a *App) askCloseSettings(tab *pane.Tab) {
	a.ask(ui.Confirm{
		Title: i18n.T("Close the settings?"),
		Text:  i18n.T("The settings hold changes you have not saved."),
		Yes:   i18n.T("Close"),
	}, func() { a.closeTab(tab) }, nil)
}

// askCloseTab asks whether to close a tab whose program still runs, and closes
// it when the answer is yes. It says what still runs, which is a shell at its
// prompt only when the settings ask about every tab.
func (a *App) askCloseTab(tab *pane.Tab) {
	a.ask(ui.Confirm{
		Title: i18n.T("Close the tab?"),
		Text:  a.closeText("The program in this tab is still running.", "Closing the tab ends the shell it runs."),
		Yes:   i18n.T("Close"),
	}, func() { a.closeTab(tab) }, nil)
}

// askCloseOtherTabs asks whether to close the other tabs whose programs still
// run, and closes them when the answer is yes.
func (a *App) askCloseOtherTabs(keep *pane.Tab) {
	a.ask(ui.Confirm{
		Title: i18n.T("Close the other tabs?"),
		Text:  a.closeText("Programs in the other tabs are still running.", "Closing the other tabs ends the shells they run."),
		Yes:   i18n.T("Close"),
	}, func() { a.closeOtherTabs(keep) }, nil)
}

// closeText is what the question says: what still runs, or, when the settings
// ask about every tab, what closing it ends.
func (a *App) closeText(running, always string) string {
	if a.settings.ConfirmClose == store.ConfirmAlways {
		return i18n.T(always)
	}
	return i18n.T(running)
}
