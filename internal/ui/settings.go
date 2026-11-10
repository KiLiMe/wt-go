package ui

import (
	"strings"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/i18n"
	"wtgo/internal/store"
)

// The pages the settings tab shows in its left column.
const (
	PageStartup     = "startup"
	PageInteraction = "interaction"
	PageAppearance  = "appearance"
	PageSchemes     = "schemes"
	PageActions     = "actions"
	PageProfiles    = "profiles"
)

// profilePrefix starts the id of the page that edits a profile.
const profilePrefix = "profile:"

// ProfilePage is the id of the page that edits the profile with an id.
func ProfilePage(id string) string { return profilePrefix + id }

// ProfileID is the profile a page edits, when it is a profile page.
func ProfileID(page string) (string, bool) { return strings.CutPrefix(page, profilePrefix) }

// SettingsTab is the tab that edits the settings: the pages in a column on the
// left, the fields of the page in front, and what saving them does under them.
func SettingsTab(c *myui.Context, st Settings, act Actions) myui.Element {
	return myui.Column(c).Fill().Children(func() {
		myui.Row(c).Grow(1).AlignItems(myui.Stretch).Children(func() {
			nav(c, st, act)
			myui.Scroll(c).Grow(1).Children(func() {
				myui.Column(c).Padding(24, 24, 16, 24).Gap(24).Children(func() {
					page(c, st, act)
				})
			})
		})
		footer(c, st, act)
	})
}

// nav is the column that lists the pages.
func nav(c *myui.Context, st Settings, act Actions) myui.Element {
	t := c.Theme()
	return myui.Column(c).Width(220).Padding(12, 8, 12, 8).Gap(2).Background(t.Surface).Children(func() {
		for _, it := range []struct{ page, label string }{
			{PageStartup, i18n.T("Startup")},
			{PageInteraction, i18n.T("Interaction")},
			{PageAppearance, i18n.T("Appearance")},
			{PageSchemes, i18n.T("Color schemes")},
			{PageActions, i18n.T("Actions")},
		} {
			navItem(c, st.Page, it.page, it.label, act)
		}
		myui.Text(c, i18n.T("Profiles")).FontSize(11).TextColor(t.TextMuted).Padding(12, 8, 4, 8)
		navItem(c, st.Page, PageProfiles, i18n.T("Defaults"), act)
		for _, p := range st.Profiles {
			navItem(c, st.Page, ProfilePage(p.ID), p.Name, act)
		}
	})
}

// navItem is one page in the column, lit while it is the page in front.
func navItem(c *myui.Context, page, target, label string, act Actions) {
	t := c.Theme()
	el := myui.Box(c).Height(30).Padding(0, 8).Justify(myui.Center).Radius(6).
		Cursor(myui.CursorPointer).OnClick(func() { act.SelectSettingsPage(target) })
	if page == target {
		el.Background(t.SurfaceHover)
	}
	el.Children(func() {
		myui.Text(c, label).SingleLine()
	})
}

// footer is what saves the page: what the draft holds that the file does not,
// and the buttons that write it or drop it.
func footer(c *myui.Context, st Settings, act Actions) myui.Element {
	t := c.Theme()
	return myui.Row(c).Height(48).Padding(0, 16).AlignItems(myui.Center).Gap(8).Background(t.Surface).Children(func() {
		note := st.Note
		if note == "" && st.Dirty {
			note = i18n.T("Unsaved changes")
		}
		if note != "" {
			myui.Text(c, note).FontSize(11).TextColor(t.TextMuted).SingleLine().MaxWidth(420)
		}
		myui.Spacer(c)
		myui.Button(c, i18n.T("Open settings.json")).OnClick(act.OpenSettingsFile)
		myui.Button(c, i18n.T("Discard")).Disabled(!st.Dirty).Opacity(dimmed(st.Dirty)).OnClick(act.DiscardSettings)
		myui.PrimaryButton(c, i18n.T("Save")).Disabled(!st.Dirty).Opacity(dimmed(st.Dirty)).OnClick(act.SaveSettings)
	})
}

// dimmed is how much of a button that does nothing shows: the theme does not
// paint a disabled button differently, so the footer says it itself.
func dimmed(live bool) float32 {
	if live {
		return 1
	}
	return 0.4
}

// page is the fields of the page in front.
func page(c *myui.Context, st Settings, act Actions) {
	if st.Draft == nil {
		return
	}
	switch st.Page {
	case PageStartup:
		startupPage(c, st, act)
	case PageInteraction:
		interactionPage(c, st, act)
	case PageAppearance:
		appearancePage(c, st, act)
	case PageSchemes:
		schemesPage(c, st, act)
	case PageActions:
		actionsPage(c, st, act)
	case PageProfiles:
		profilesPage(c, st, act)
	default:
		if id, ok := strings.CutPrefix(st.Page, profilePrefix); ok {
			profilePage(c, st, act, id)
			return
		}
		profilesPage(c, st, act)
	}
}

// startupPage is what the window does when it opens.
func startupPage(c *myui.Context, st Settings, act Actions) {
	s := st.Draft
	section(c, i18n.T("Startup"), func() {
		field(c, i18n.T("Default profile"), i18n.T("The profile a new tab runs unless it picks one."), func() {
			opts := make([]option, 0, len(st.Known)+1)
			opts = append(opts, option{"", i18n.T("None")})
			for _, p := range st.Known {
				opts = append(opts, option{p.ID, p.Name})
			}
			choice(c, &s.DefaultProfile, opts)
		})
		field(c, i18n.T("Language"), "", func() {
			choice(c, &s.Language, []option{{"zh-CN", "简体中文"}, {"en-US", "English"}})
		})
		field(c, i18n.T("Start directory"), i18n.T("Where a new tab starts when its profile names no directory."), func() {
			myui.TextInput(c, &s.StartDirectory).Width(320).Placeholder(store.DefaultHomeHint())
		})
		field(c, i18n.T("Start size"), i18n.T("Columns and rows of the window when it opens; 0 keeps the size it opens with."), func() {
			myui.Row(c).AlignItems(myui.Center).Gap(8).Children(func() {
				count(c, &s.StartSize.Columns, 0, 400)
				myui.Text(c, "x").TextColor(c.Theme().TextMuted)
				count(c, &s.StartSize.Rows, 0, 200)
			})
		})
		field(c, i18n.T("Close on last tab"), i18n.T("Close the window when its last tab closes."), func() {
			myui.Switch(c, &s.CloseOnLastTab)
		})
	})
}

// interactionPage is how the terminal answers the mouse and the clipboard.
func interactionPage(c *myui.Context, st Settings, act Actions) {
	s := st.Draft
	section(c, i18n.T("Interaction"), func() {
		field(c, i18n.T("Copy on select"), i18n.T("Copy what a selection takes as it takes it."), func() {
			myui.Switch(c, &s.CopyOnSelect)
		})
		field(c, i18n.T("Right click"), i18n.T("What a right click does."), func() {
			choice(c, &s.RightClick, []option{
				{"copyOrPaste", i18n.T("Copy or paste")},
				{"paste", i18n.T("Paste")},
				{"menu", i18n.T("Menu")},
				{"nothing", i18n.T("Nothing")},
			})
		})
		field(c, i18n.T("Paste multiline confirm"), i18n.T("Ask before a paste that holds more than one line."), func() {
			myui.Switch(c, &s.PasteMultilineConfirm)
		})
		field(c, i18n.T("Confirm close"), i18n.T("Ask before a tab closes while its program still runs."), func() {
			myui.Switch(c, &s.ConfirmClose)
		})
		field(c, i18n.T("Panes per tab"), i18n.T("How many panes a tab splits into; 0 for no limit."), func() {
			count(c, &s.MaxPanes, 0, 32)
		})
	})
}

// appearancePage is how the window looks.
func appearancePage(c *myui.Context, st Settings, act Actions) {
	s := st.Draft
	section(c, i18n.T("Appearance"), func() {
		field(c, i18n.T("Theme"), i18n.T("The light or dark face of the window itself."), func() {
			choice(c, &s.Theme, []option{
				{"system", i18n.T("Follow the system")},
				{"light", i18n.T("Light")},
				{"dark", i18n.T("Dark")},
			})
		})
		field(c, i18n.T("Background opacity"), i18n.T("How much of what is behind the window shows through it."), func() {
			ratio(c, &s.BackgroundOpacity, 0.3, 1, 0.05)
		})
		field(c, i18n.T("New tab position"), i18n.T("Where a new tab goes."), func() {
			choice(c, &s.NewTabPosition, []option{
				{"last", i18n.T("After the last tab")},
				{"first", i18n.T("Before the first tab")},
			})
		})
		field(c, i18n.T("Tab width"), i18n.T("How wide the tabs are."), func() {
			choice(c, &s.TabWidth, []option{
				{"compact", i18n.T("As wide as the name")},
				{"equal", i18n.T("All the same")},
			})
		})
		field(c, i18n.T("Always on top"), i18n.T("Keep the window above the others."), func() {
			myui.Switch(c, &s.AlwaysOnTop)
		})
		field(c, i18n.T("Title follows the pane"), i18n.T("Name the window after the pane with the keyboard."), func() {
			myui.Switch(c, &s.TitleFromActiveTerminal)
		})
	})
}

// profilesPage is the settings a profile starts from, and the profiles.
func profilesPage(c *myui.Context, st Settings, act Actions) {
	s := st.Draft
	t := c.Theme()
	section(c, i18n.T("Defaults"), func() {
		field(c, i18n.T("Color scheme"), i18n.T("The colors a terminal paints with unless its profile names another."), func() {
			choice(c, &s.ColorScheme, schemeOptions(st.Schemes))
		})
		field(c, i18n.T("Font"), "", func() {
			myui.TextInput(c, &s.Font.Family).Width(220).Placeholder(store.DefaultFontFamily())
		})
		field(c, i18n.T("Font size"), i18n.T("How tall the letters of a terminal are, in DIPs."), func() {
			measure(c, &s.Font.Size, 6, 48, 1)
		})
		field(c, i18n.T("Line height"), i18n.T("The height of a line of a terminal, as a multiple of the font size."), func() {
			measure(c, &s.Font.LineHeight, 0.8, 2.4, 0.1)
		})
		field(c, i18n.T("Cursor"), "", func() {
			myui.Row(c).AlignItems(myui.Center).Gap(12).Children(func() {
				choice(c, &s.CursorShape, []option{
					{"bar", i18n.T("Bar")},
					{"block", i18n.T("Block")},
					{"underline", i18n.T("Underline")},
				})
				myui.Switch(c, &s.CursorBlink)
				myui.Text(c, i18n.T("Blink")).FontSize(11).TextColor(t.TextMuted)
			})
		})
		field(c, i18n.T("Padding"), i18n.T("The space inside the edges of a terminal, in DIPs."), func() {
			myui.Row(c).AlignItems(myui.Center).Gap(8).Children(func() {
				myui.Text(c, i18n.T("Across")).FontSize(11).TextColor(t.TextMuted)
				measure(c, &s.Padding.X, 0, 64, 1)
				myui.Text(c, i18n.T("Down")).FontSize(11).TextColor(t.TextMuted)
				measure(c, &s.Padding.Y, 0, 64, 1)
			})
		})
		field(c, i18n.T("Scrollback"), i18n.T("How much a terminal keeps to scroll back through, in MB."), func() {
			count(c, &s.ScrollbackMB, 0, 1024)
		})
	})
	section(c, i18n.T("Profiles"), func() {
		for _, p := range st.Profiles {
			myui.Row(c).Height(32).AlignItems(myui.Center).Gap(8).Children(func() {
				myui.Text(c, p.Name).Grow(1)
				myui.Button(c, i18n.T("Duplicate")).OnClick(func() { act.DuplicateProfile(p.ID) })
				myui.Button(c, i18n.T("Remove")).OnClick(func() { act.RemoveProfile(p.ID) })
			})
		}
		myui.Row(c).Children(func() {
			myui.Button(c, i18n.T("Add profile")).OnClick(act.AddProfile)
		})
	})
}

// profilePage is the profile with an id, as the copy the page edits.
func profilePage(c *myui.Context, st Settings, act Actions, id string) {
	p := st.Profile
	if p == nil || p.ID != id {
		myui.Text(c, i18n.T("Select a profile.")).TextColor(c.Theme().TextMuted)
		return
	}
	section(c, p.Name, func() {
		field(c, i18n.T("Name"), "", func() {
			myui.TextInput(c, &p.Name).Width(320)
		})
		field(c, i18n.T("Command line"), i18n.T("The program the profile runs; empty runs the shell the system starts."), func() {
			myui.TextInput(c, &p.Command).Width(320).Placeholder(store.DefaultShellHint())
		})
		field(c, i18n.T("Arguments"), i18n.T("The arguments of the command, one on each line."), func() {
			myui.TextArea(c, st.Args).Lines(2, 6)
		})
		field(c, i18n.T("Start directory"), i18n.T("Where the profile starts; empty uses the setting."), func() {
			myui.TextInput(c, &p.Dir).Width(320).Placeholder(store.DefaultHomeHint())
		})
		field(c, i18n.T("Icon color"), i18n.T("The color of the dot the tab strip marks the profile with."), func() {
			colorField(c, &p.Color)
		})
		field(c, i18n.T("Hidden"), i18n.T("Leave the profile out of the menu that opens tabs."), func() {
			myui.Switch(c, &p.Hidden)
		})
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.Button(c, i18n.T("Duplicate")).OnClick(func() { act.DuplicateProfile(p.ID) })
		myui.Button(c, i18n.T("Remove")).OnClick(func() { act.RemoveProfile(p.ID) })
	})
}
