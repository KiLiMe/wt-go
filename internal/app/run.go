package app

import (
	"fmt"
	"github.com/egoist/mygo"
	myui "github.com/egoist/mygo/ui"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"wtgo/internal/cli"
	"wtgo/internal/fault"
	"wtgo/internal/i18n"
	"wtgo/internal/sys"
	"wtgo/internal/theme"
	"wtgo/internal/ui"
)

// Run opens the window and runs until it closes.
func Run() (err error) {
	// The command line: the commands that answer without a window (`wt-go diag`,
	// `wt-go mcp`), and -d, which says where the window's new tabs start.
	command := cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	if command.Handled {
		if command.Code != 0 {
			os.Exit(command.Code)
		}
		return nil
	}
	// A window started from Explorer has no console: what it says goes to
	// wt-go.log as well, so that `wt-go diag` can read it back later.
	if f, ferr := sys.OpenLog(); ferr == nil {
		// The file comes first: a window started from Explorer has no console,
		// and writing to one that is not there must not lose the line.
		log.SetOutput(io.MultiWriter(f, console{os.Stderr}))
	}
	defer func() {
		if r := recover(); r != nil {
			reportCrash(r)
			panic(r)
		}
	}()
	if err := sys.HoldChildren(); err != nil {
		log.Printf("wt-go: %v", err)
	}
	mygo.App.WhenReady(func() {
		// Before the first layout: a glyph drawn after it is already the wrong
		// color, and the fix only reaches glyphs drawn later.
		if !sys.FixTextColor() {
			log.Print(fault.Line(fault.NoFontconfig, ""))
		}
		a := newApp(command.StartDir)
		a.appearance()
		// Before the window: the store is read when the first window with a
		// StateKey is created, and the window opens at the geometry it holds.
		mygo.App.SetWindowStateStore(windowState{})
		a.win = mygo.NewWindow(mygo.WindowOptions{
			Title:             "wt-go",
			Width:             960,
			Height:            600,
			MinWidth:          int(ui.MinWindowWidth),
			MinHeight:         int(ui.MinWindowHeight),
			TitleBarStyle:     titleBarStyle(),
			TitleBarHeight:    ui.TitleBarHeight,
			HideWindowButtons: hideWindowButtons(),
			StateKey:          "main",
			BackgroundColor:   theme.Hex(a.background()),
			Opacity:           opacity(a.settings.BackgroundOpacity),
			Content:           myui.View(a.view),
		})
		a.win.OnClosed(a.onClosed)
		a.win.OnClose(a.onClose)
		a.watchWindowButtons()
		a.openFirst()
		if a.settingsNote() != "" {
			a.OpenSettings()
		}
		go a.watch(a.watchStop)
	})
	return mygo.App.Run()
}

// faultCode is the code a settings error answers for: a field out of range
// is a different fix from a file that is not JSON at all.
func faultCode(err error) fault.Code {
	if strings.Contains(err.Error(), " must be ") {
		return fault.BadField
	}
	return fault.BadSettings
}

// reportCrash writes what stopped the program next to the settings, so that a
// window that goes away leaves something to read.
func reportCrash(r any) {
	path, err := sys.WriteCrashLog(fmt.Sprint(r), debug.Stack())
	if err != nil {
		log.Printf("wt-go: stopped: %v (%v)", r, err)
		return
	}
	log.Printf("wt-go: stopped: %v (written to %s)", r, path)
}

// locale makes the window say its messages in the language the settings name,
// "" for the system's. A file written by hand may hold a tag that names no
// language, which MyGo panics on: the window shows the system's instead of
// refusing to start.
func locale(tag string) {
	if !i18n.Valid(tag) {
		tag = ""
	}
	mygo.App.SetLocale(tag)
}

// onClosed closes the terminals the window was showing.
func (a *App) onClosed() {
	a.closed = true
	if a.watchStop != nil {
		close(a.watchStop)
		a.watchStop = nil
	}
	for _, tab := range a.panes.Tabs {
		closeTerminals(tab)
	}
	a.panes.Tabs = nil
	a.settingsTab = nil
}
