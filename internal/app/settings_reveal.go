package app

import (
	"time"

	"wtgo/internal/ui"
)

// RevealSetting shows a page of the settings with the setting the search found
// marked, so that the row a hit names is the row the page shows: the page
// scrolls to it on its first frame and the mark fades from there. Going through
// SelectSettingsPage keeps the search box empty, which is what a second search
// needs, and the page otherwise opens as its own button opens it.
func (a *App) RevealSetting(page, key string) {
	if a.settingsTab == nil {
		return
	}
	a.SelectSettingsPage(page)
	a.reveal = ui.Reveal{Key: key, At: time.Now()}
	a.redraw()
}

// revealFrame ages the mark over the setting the search found, after the frame
// that drew it: that frame is the first one, and the ones after it are the
// fade. The mark goes with the settings tab, and with the search that put it
// there: a page opened without searching has nothing to mark.
func (a *App) revealFrame() {
	if a.reveal.Key == "" {
		return
	}
	if a.settingsTab == nil {
		a.reveal = ui.Reveal{}
		return
	}
	a.reveal.Frame++
	if time.Since(a.reveal.At) > ui.RevealFade {
		a.reveal = ui.Reveal{}
	}
}
