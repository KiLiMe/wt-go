//go:build windows || darwin

package app

import "github.com/egoist/mygo"

// titleBarStyle hides the title bar because mygo draws the window buttons
// itself, which makes the tab strip the title bar.
func titleBarStyle() mygo.TitleBarStyle { return mygo.TitleBarHidden }
