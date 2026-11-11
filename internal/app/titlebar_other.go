//go:build windows || darwin

package app

import "github.com/egoist/mygo"

// titleBarStyle hides the title bar because mygo draws the window buttons
// itself, which makes the tab strip the title bar.
func titleBarStyle() mygo.TitleBarStyle { return mygo.TitleBarHidden }

// hideWindowButtons keeps mygo's window buttons, which follow the system:
// Windows draws its own, with the snap layouts the maximize button shows, and
// macOS keeps the traffic lights.
func hideWindowButtons() bool { return false }
