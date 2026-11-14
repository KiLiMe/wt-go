package app

import (
	"strings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/transfer"

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
// the dialog is its no, which the button says. Whatever other overlay was open
// goes, and the question is the one the answers answer: a callback that arrives
// after another overlay took its place must not act.
func (a *App) ask(q ui.Confirm, yes, no func()) {
	q.Open = true
	q.No = locales.Cancel()
	a.openOverlay(overlayConfirm)
	a.confirm = &ask{Confirm: q, yes: yes, no: no}
	a.redraw()
}

// ConfirmYes does what the question asks.
func (a *App) ConfirmYes() {
	q := a.confirm
	a.closeOverlay(overlayConfirm)
	if q != nil && q.yes != nil {
		q.yes()
	}
	a.redraw()
}

// ConfirmNo drops the question.
func (a *App) ConfirmNo() {
	q := a.confirm
	a.closeOverlay(overlayConfirm)
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
		Text:  locales.ConfirmPasteText(lines),
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
		Text:  a.closeText(act, closeTabTexts),
		Yes:   locales.Close(),
	}, func() { a.closeTab(tab) }, nil)
}

// askCloseOtherTabs asks whether to close the other tabs whose programs still
// run, and closes them when the answer is yes.
func (a *App) askCloseOtherTabs(keep *pane.Tab, act tabActivity) {
	a.ask(ui.Confirm{
		Title: locales.ConfirmCloseOthersTitle(),
		Text:  a.closeText(act, closeOthersTexts),
		Yes:   locales.Close(),
	}, func() { a.closeOtherTabs(keep) }, nil)
}

// askCloseWindow asks whether the window may close, which takes the tabs it
// holds and the shells they run with it, and closes it when the answer is yes.
func (a *App) askCloseWindow() {
	a.ask(ui.Confirm{
		Title: locales.ConfirmCloseWindowTitle(),
		Text:  locales.ConfirmCloseWindowText(len(a.panes.Tabs)),
		Yes:   locales.Close(),
	}, a.closeWindow, nil)
}

// closeTexts are what a closing question says for each of the three answers
// it may give: what still runs, what the window cannot tell, and what closing
// ends when the settings ask about every tab.
type closeTexts struct {
	running func() string
	unsure  func() string
	always  func() string
}

var (
	closeTabTexts = closeTexts{
		running: locales.TabRunningText,
		unsure:  locales.TabUnsureText,
		always:  locales.TabAlwaysText,
	}
	closeOthersTexts = closeTexts{
		running: locales.OthersRunningText,
		unsure:  locales.OthersUnsureText,
		always:  locales.OthersAlwaysText,
	}
)

// closeText is what the question says: which programs still run, or, when the
// settings ask about every tab, what closing it ends. What the window cannot
// tell is said as that, and not as a program it found.
func (a *App) closeText(act tabActivity, texts closeTexts) string {
	switch {
	case a.settings.ConfirmClose == store.ConfirmAlways:
		return texts.always()
	case act.act == sys.Unknown:
		return texts.unsure()
	case len(act.programs) > 0:
		// Naming the programs is what makes the question answerable: a
		// shell at its prompt is not, and a program the user forgot is.
		return locales.CloseRunning(strings.Join(act.programs, ", "))
	}
	return texts.running()
}
