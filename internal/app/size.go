package app

import (
	"math"

	"github.com/egoist/mygo/plugins/terminal"

	"wtgo/internal/layout"
	"wtgo/internal/store"
	"wtgo/internal/ui"
)

// grow is how much bigger the window gets while it measures a cell: enough to
// gain several cells, which the measuring reads as the size of one.
const grow = 120

// reading is what the terminal reports: how many cells it holds, and the room
// it takes.
type reading struct {
	cols, rows int
	room       layout.Rect
}

// startSize gives the window the grid the settings start it with. A terminal
// reports how many cells fit and not how big one is, and the room a size
// leaves over hides that, so the window grows once, reads the cells it gained,
// and then takes the size that holds the cells asked for.
//
// It waits for the terminal to report the same size in two frames before it
// reads one, since the terminal takes a size a frame after the window does.
func (a *App) startSize() {
	if a.sized || a.win == nil || a.frames < 3 {
		return
	}
	want := a.settings.StartSize
	if want.Columns <= 0 || want.Rows <= 0 {
		a.sized = true
		return
	}
	tab, ok := a.panes.Front()
	if !ok || tab.Focus == nil || tab.Focus.Term == nil {
		return
	}
	now, settled := a.settled(tab.Focus.Term)
	if !settled {
		return
	}
	switch a.measure {
	case 0:
		a.base = now
		w, h := a.win.Size()
		a.settle()
		a.win.SetSize(w+grow, h+grow)
		a.measure = 1
	case 1:
		cols, rows := now.cols-a.base.cols, now.rows-a.base.rows
		if cols <= 0 || rows <= 0 {
			a.sized = true
			return
		}
		a.cellW = (now.room.W - a.base.room.W) / float32(cols)
		a.cellH = (now.room.H - a.base.room.H) / float32(rows)
		a.settle()
		a.setSize(want)
		a.measure = 2
	case 2:
		if now.cols == want.Columns && now.rows == want.Rows {
			a.sized = true
			return
		}
		// The cells a window holds round down, so the size that holds the
		// cells asked for may hold one less.
		w, h := a.win.Size()
		a.settle()
		a.win.SetSize(
			w+int(math.Round(float64(float32(want.Columns-now.cols)*a.cellW))),
			h+int(math.Round(float64(float32(want.Rows-now.rows)*a.cellH))),
		)
		a.measure = 3
	case 3:
		a.sized = true
	}
}

// settled is what the terminal reports, and whether it reported the same in
// the frame before.
func (a *App) settled(term *terminal.Terminal) (reading, bool) {
	cols, rows := term.Size()
	now := reading{cols: cols, rows: rows, room: a.content}
	ok := cols > 0 && rows > 0 && now == a.last
	a.last = now
	return now, ok
}

// settle forgets what the terminal reported, so that it reports twice again
// once the window takes the size asked for.
func (a *App) settle() { a.last = reading{} }

// setSize gives the window the room a grid of cells takes, with the borders
// and the strip it draws around them.
func (a *App) setSize(want store.Size) {
	if a.cellW <= 0 || a.cellH <= 0 {
		a.sized = true
		return
	}
	pad := a.settings.Padding
	room := layout.Rect{
		W: float32(want.Columns)*a.cellW + 2*float32(pad.X),
		H: float32(want.Rows)*a.cellH + 2*float32(pad.Y),
	}
	w, h := a.win.Size()
	clientW, clientH := a.content.W, a.content.H+ui.TitleBarHeight
	a.win.SetSize(
		int(math.Round(float64(room.W+float32(w)-clientW))),
		int(math.Round(float64(room.H+float32(h)-clientH))),
	)
}
