package app

import (
	"strings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/transfer"

	"wtgo/internal/i18n"
	"wtgo/internal/pane"
	"wtgo/internal/store"
	"wtgo/internal/sys"
	"wtgo/internal/ui"
	"wtgo/locales"
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
	q.No = locales.Cancel()
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
		Title: locales.ConfirmPasteTitle(),
		Text:  i18n.F("confirmPasteText", map[string]any{"count": lines}),
		Yes:   locales.Paste(),
	}, func() { term.PasteData(d) }, nil)
	return true
}

// askCloseSettings asks whether to close the settings tab, which drops the
// changes it holds, and closes it when the answer is yes.
func (a *App) askCloseSettings(tab *pane.Tab) {
	a.ask(ui.Confirm{
		Title: locales.ConfirmCloseSettingsTitle(),
		Text:  locales.ConfirmCloseSettingsText(),
		Yes:   locales.Close(),
	}, func() { a.closeTab(tab) }, nil)
}

// askCloseTab asks whether to close a tab whose program still runs, and closes
// it when the answer is yes. It says what still runs, which is a shell at its
// prompt only when the settings ask about every tab.
func (a *App) askCloseTab(tab *pane.Tab, act tabActivity) {
	a.ask(ui.Confirm{
		Title: locales.ConfirmCloseTabTitle(),
		Text:  a.closeText(act, "tabRunningText", "tabUnsureText", "tabAlwaysText"),
		Yes:   locales.Close(),
	}, func() { a.closeTab(tab) }, nil)
}

// askCloseOtherTabs asks whether to close the other tabs whose programs still
// run, and closes them when the answer is yes.
func (a *App) askCloseOtherTabs(keep *pane.Tab, act tabActivity) {
	a.ask(ui.Confirm{
		Title: locales.ConfirmCloseOthersTitle(),
		Text:  a.closeText(act, "othersRunningText", "othersUnsureText", "othersAlwaysText"),
		Yes:   locales.Close(),
	}, func() { a.closeOtherTabs(keep) }, nil)
}

// closeText is what the question says: which programs still run, or, when the
// settings ask about every tab, what closing it ends. What the window cannot
// tell is said as that, and not as a program it found. The three are the keys
// of the messages it may say.
func (a *App) closeText(act tabActivity, running, unsure, always string) string {
	switch {
	case a.settings.ConfirmClose == store.ConfirmAlways:
		return i18n.T(always)
	case act.act == sys.Unknown:
		return i18n.T(unsure)
	case len(act.programs) > 0:
		// Naming the programs is what makes the question answerable: a
		// shell at its prompt is not, and a program the user forgot is.
		return i18n.F("closeRunning", map[string]any{"programs": strings.Join(act.programs, ", ")})
	}
	return i18n.T(running)
}
