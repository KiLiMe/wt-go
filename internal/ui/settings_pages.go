package ui

import (
	myui "github.com/egoist/mygo/ui"
	"wtgo/internal/i18n"
	"wtgo/internal/store"
	"wtgo/locales"
)

// generalPage is what the window starts up with, how its tabs behave, and what
// it asks before it closes: the questions live with the things they close.
func generalPage(c *myui.Context, st Settings, act Actions) {
	s := st.Draft
	section(c, locales.GeneralStartup(), func() {
		field(c, locales.DefaultProfile(), locales.DefaultProfileNote(), func() {
			opts := make([]option, 0, len(st.Known)+1)
			// The choice a person reads names the shell it runs, the way the
			// strip's menu names it.
			opts = append(opts, option{"", deviceLabel(st.Device)})
			for _, p := range st.Known {
				opts = append(opts, option{p.ID, p.Name})
			}
			// A profile the file names but no longer offers runs the shell of the
			// machine anyway, so the select says following the device rather than
			// a choice that is gone. The value the select edits is the page's
			// copy, so that drawing the page never changes the draft: only a
			// click names a profile for the window to write down.
			offered := s.DefaultProfile
			if offered != "" && !offers(st.Known, offered) {
				offered = ""
			}
			opts[0].Value = offered
			picked := offered
			choice(c, &picked, opts)
			if picked != offered {
				act.SetDefaultProfile(picked)
			}
		})
		field(c, locales.Language(), "", func() {
			langs := i18n.Languages()
			opts := make([]option, 0, len(langs)+1)
			opts = append(opts, option{"", locales.SystemLanguage()})
			for _, tag := range langs {
				opts = append(opts, option{tag, i18n.Autonym(tag)})
			}
			choice(c, &s.Language, opts)
		})
		fieldBlock(c, locales.StartDirectory(), locales.StartDirectoryNote(), func() {
			myui.TextInput(c, &s.StartDirectory).Placeholder(store.DefaultHomeHint())
		})
	})
	section(c, locales.GeneralWindow(), func() {
		field(c, locales.NewTabPosition(), locales.NewTabPositionNote(), func() {
			choice(c, &s.NewTabPosition, []option{
				{"last", locales.AfterLastTab()},
				{"first", locales.BeforeFirstTab()},
			})
		})
		field(c, locales.TabWidth(), locales.TabWidthNote(), func() {
			choice(c, &s.TabWidth, []option{
				{"compact", locales.AsWideAsName()},
				{"equal", locales.AllTheSame()},
			})
		})
		field(c, locales.TitleFollowsPane(), locales.TitleFollowsPaneNote(), func() {
			myui.Switch(c, &s.TitleFromActiveTerminal)
		})
		field(c, locales.AlwaysOnTop(), locales.AlwaysOnTopNote(), func() {
			myui.Switch(c, &s.AlwaysOnTop)
		})
		field(c, locales.PanesPerTab(), locales.PanesPerTabNote(), func() {
			count(c, &s.MaxPanes, 0, 32)
		})
	})
	section(c, locales.GeneralClose(), func() {
		field(c, locales.CloseOnLastTab(), locales.CloseOnLastTabNote(), func() {
			myui.Switch(c, &s.CloseOnLastTab)
		})
		field(c, locales.ConfirmClose(), locales.ConfirmCloseNote(), func() {
			choice(c, &s.ConfirmClose, []option{
				{string(store.ConfirmRunning), locales.WhileProgramRuns()},
				{string(store.ConfirmAlways), locales.Always()},
				{string(store.ConfirmNever), locales.Never()},
			})
		})
		// It asks about the way the window goes, which closes the tab above it
		// too, so it follows that switch.
		field(c, locales.ConfirmCloseWindow(), locales.ConfirmCloseWindowNote(), func() {
			myui.Switch(c, &s.ConfirmCloseWindow)
		})
	})
}

// offers reports whether a list of profiles holds an id.
func offers(profiles []Profile, id string) bool {
	for _, p := range profiles {
		if p.ID == id {
			return true
		}
	}
	return false
}

// interactionPage is how the terminal answers the mouse and the clipboard,
// and how much it keeps to look back through.
func interactionPage(c *myui.Context, st Settings, act Actions) {
	s := st.Draft
	section(c, locales.InteractionMouse(), func() {
		field(c, locales.CopyOnSelect(), locales.CopyOnSelectNote(), func() {
			myui.Switch(c, &s.CopyOnSelect)
		})
		field(c, locales.RightClick(), locales.RightClickNote(), func() {
			choice(c, &s.RightClick, []option{
				{"copyOrPaste", locales.CopyOrPaste()},
				{"paste", locales.Paste()},
				{"menu", locales.Menu()},
				{"nothing", locales.Nothing()},
			})
		})
		field(c, locales.RightClickCopyPastePriority(), locales.RightClickCopyPastePriorityNote(), func() {
			myui.Switch(c, &s.RightClickCopyPastePriority)
		})
		field(c, locales.Links(), locales.LinksNote(), func() {
			myui.Switch(c, &s.Links.Open)
		})
		field(c, locales.ScrollMultiplier(), locales.ScrollMultiplierNote(), func() {
			measure(c, &s.ScrollMultiplier, 0.5, 5, 0.5)
		})
	})
	section(c, locales.InteractionClipboard(), func() {
		field(c, locales.ClipboardRead(), locales.ClipboardReadNote(), func() {
			choice(c, &s.ClipboardRead, []option{
				{"deny", locales.Deny()},
				{"allow", locales.Allow()},
				{"ask", locales.Ask()},
			})
		})
		field(c, locales.ClipboardWrite(), locales.ClipboardWriteNote(), func() {
			choice(c, &s.ClipboardWrite, []option{
				{"allow", locales.Allow()},
				{"deny", locales.Deny()},
			})
		})
		field(c, locales.PasteMultiline(), locales.PasteMultilineNote(), func() {
			myui.Switch(c, &s.PasteMultilineConfirm)
		})
	})
	section(c, locales.InteractionBuffer(), func() {
		field(c, locales.Scrollback(), locales.ScrollbackNote(), func() {
			count(c, &s.ScrollbackMB, 0, 1024)
		})
	})
}
