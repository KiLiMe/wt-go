package ui

// Boundary: the strip's measurements, and the room a window needs for it. A
// number here is a contract with the desktop as much as with the drawing: the
// files that draw read them, none of them sets one. The window's own buttons
// are window_buttons.go, and what room each takes is set here.

// TitleBarHeight is the height of the tab strip, which is the title bar: the
// window keeps it clear of its window controls, and the strip runs under
// them. It is Windows Terminal's 30 DIP, so that a window beside one keeps
// the two bars level and its pane starts where the other one's does.
const TitleBarHeight = 30

// stripGutter is the room the strip keeps between the window's left edge and
// the first tab, for the same reason.
const stripGutter float32 = 8

// nameBold is the weight a bold name is drawn in, as myui's Bold names it.
const nameBold = 700

// tabChrome is the room a tab takes around its name: the padding on both sides,
// the mark before the name and the button that closes it, with the gaps between
// them.
const tabChrome float32 = 68

// The room a tab takes: Windows Terminal's widest, which a tab with a long name
// and room to spare reaches, and the narrowest that still holds the mark, the
// close button and a name -- below it a tab says nothing, and the strip scrolls
// instead of squeezing the tabs further.
const (
	minTabWidth float32 = tabChrome + minTabName
	maxTabWidth float32 = 240
)

// The room a tab's name may take: as little as a name that is a single
// letter, as much as fits the widest tab.
const (
	minTabName float32 = 48
	maxTabName float32 = maxTabWidth - tabChrome
)

// mouseMiddle is the button that closes the tab it points at, as it does in
// Windows Terminal.
const mouseMiddle = 2

// The room the strip keeps whatever the tabs do: the buttons that move them,
// and the part of the title bar that moves the window. A strip of many tabs
// never takes the last of either.
const (
	stepWidth float32 = 24
	dragRoom  float32 = 56
)

// tabStep is how far the buttons beside the tabs move what the strip shows:
// about one tab, so that a press brings the next tabs into view.
const tabStep float32 = 176

// MinWindowWidth is the narrowest the window may be and still show a tab: the
// room of one tab at its narrowest, the buttons beside the tabs, the room that
// moves the window, the two buttons that scroll the tabs, and the desktop's
// own window buttons -- three of them at the widest a desktop draws them, which
// is Windows' 46 DIP each, so that macOS' traffic lights fit too. Narrower than
// this the strip shows no tab at all, only the buttons that move tabs that are
// not there. It is a variable because stripButtonsWidth is a function.
var MinWindowWidth = stripGutter + minTabWidth + dragRoom +

	stripButtonsWidth() + 2*stepWidth + 3*46

// MinWindowHeight is the shortest the window may be: the strip tall enough for
// the title bar, and a pane below it tall enough for a prompt and the two lines
// above it at the default font and padding -- three rows of 16 DIPs and the
// padding of 8 above and below. Short of it the pane holds no line of output at
// all, and the window has nothing to show.
const MinWindowHeight float32 = TitleBarHeight + 3*16 + 2*8

// The room a window button takes, the gap between two of them, and the room
// the buttons keep from the corner of the window.
const (
	windowButtonWidth float32 = 34
	windowButtonGap   float32 = 2
	windowButtonRoom  float32 = 6
)
