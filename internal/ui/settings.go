package ui

import (
	"strings"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/i18n"
	"wtgo/internal/icons"
	"wtgo/internal/store"
	"wtgo/internal/theme"
	"wtgo/locales"
)

// The pages the settings tab shows in its left column.
const (
	PageStartup     = "startup"
	PageInteraction = "interaction"
	PageAppearance  = "appearance"
	PageSchemes     = "schemes"
	PageActions     = "actions"
	PageAdvanced    = "advanced"
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
		// The pages scroll, so that adding profiles never pushes the row that
		// adds the next one out of reach.
		myui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
			myui.Column(c).Gap(2).Children(func() {
				for _, it := range []struct{ page, label string }{
					{PageStartup, locales.Startup()},
					{PageInteraction, locales.Interaction()},
					{PageAppearance, locales.Appearance()},
					{PageSchemes, locales.ColorSchemes()},
					{PageActions, locales.Actions()},
					{PageAdvanced, locales.Advanced()},
				} {
					navItem(c, st.Page, it.page, it.label, act)
				}
				myui.Text(c, locales.Profiles()).FontSize(11).TextColor(t.TextMuted).Padding(12, 8, 4, 8)
				navItem(c, st.Page, PageProfiles, locales.Defaults(), act)
				for _, p := range st.Profiles {
					navItem(c, st.Page, ProfilePage(p.ID), p.Name, act)
				}
			})
		})
		myui.Column(c).Gap(2).Padding(6, 0, 0, 0).Children(func() {
			navAction(c, icons.Add.SVG, locales.AddProfile(), act.AddProfile)
			navCheck(c, st.Hidden, locales.ShowHiddenProfiles(), act.ShowHiddenProfiles)
		})
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

// navAction is a row under the pages that does something instead of opening a
// page, so it sits where the list ends rather than on the page it acts on.
func navAction(c *myui.Context, icon *myui.SVG, label string, onClick func()) {
	t := c.Theme()
	el := myui.Row(c).Height(30).Padding(0, 6).Gap(8).AlignItems(myui.Center).Radius(6).
		Cursor(myui.CursorPointer).OnClick(onClick)
	if el.Hovered() {
		el.Background(t.SurfaceHover)
	}
	el.Children(func() {
		myui.Icon(c, icon).Size(14, 14).TextColor(t.TextMuted)
		myui.Text(c, label).FontSize(13).SingleLine()
	})
}

// navCheck is a switch that holds in the column, under the pages.
func navCheck(c *myui.Context, on bool, label string, toggle func()) {
	if myui.Checkbox(c, &on, label).FontSize(12).Changed() {
		toggle()
	}
}

// footer is what saves the page: what the draft holds that the file does not,
// and the buttons that write it or drop it.
func footer(c *myui.Context, st Settings, act Actions) myui.Element {
	t := c.Theme()
	return myui.Row(c).Height(48).Padding(0, 16).AlignItems(myui.Center).Gap(8).Background(t.Surface).Children(func() {
		note := st.Note
		if note == "" && st.Dirty {
			note = locales.UnsavedChanges()
		}
		if note != "" {
			myui.Text(c, note).FontSize(11).TextColor(t.TextMuted).SingleLine().MaxWidth(420)
		}
		myui.Spacer(c)
		myui.Button(c, locales.OpenSettingsFolder()).OnClick(act.OpenSettingsFolder)
		myui.Button(c, locales.Discard()).Disabled(!st.Dirty).Opacity(dimmed(st.Dirty)).OnClick(act.DiscardSettings)
		myui.PrimaryButton(c, locales.Save()).Disabled(!st.Dirty).Opacity(dimmed(st.Dirty)).OnClick(act.SaveSettings)
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
	case PageAdvanced:
		advancedPage(c, st, act)
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
	section(c, locales.Startup(), func() {
		field(c, locales.DefaultProfile(), locales.DefaultProfileNote(), func() {
			opts := make([]option, 0, len(st.Known)+1)
			opts = append(opts, option{"", locales.None()})
			for _, p := range st.Known {
				opts = append(opts, option{p.ID, p.Name})
			}
			choice(c, &s.DefaultProfile, opts)
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
		field(c, locales.StartDirectory(), locales.StartDirectoryNote(), func() {
			myui.TextInput(c, &s.StartDirectory).Placeholder(store.DefaultHomeHint())
		})
		field(c, locales.StartSize(), locales.StartSizeNote(), func() {
			myui.Row(c).AlignItems(myui.Center).Gap(8).Children(func() {
				count(c, &s.StartSize.Columns, 0, 400)
				myui.Text(c, "\u00d7").TextColor(c.Theme().TextMuted)
				count(c, &s.StartSize.Rows, 0, 200)
			})
		})
		field(c, locales.CloseOnLastTab(), locales.CloseOnLastTabNote(), func() {
			myui.Switch(c, &s.CloseOnLastTab)
		})
	})
}

// interactionPage is how the terminal answers the mouse and the clipboard.
func interactionPage(c *myui.Context, st Settings, act Actions) {
	s := st.Draft
	section(c, locales.Interaction(), func() {
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
		field(c, locales.PasteMultiline(), locales.PasteMultilineNote(), func() {
			myui.Switch(c, &s.PasteMultilineConfirm)
		})
		field(c, locales.ConfirmClose(), locales.ConfirmCloseNote(), func() {
			choice(c, &s.ConfirmClose, []option{
				{string(store.ConfirmRunning), locales.WhileProgramRuns()},
				{string(store.ConfirmAlways), locales.Always()},
				{string(store.ConfirmNever), locales.Never()},
			})
		})
		field(c, locales.PanesPerTab(), locales.PanesPerTabNote(), func() {
			count(c, &s.MaxPanes, 0, 32)
		})
	})
}

// appearancePage is how the window looks.
func appearancePage(c *myui.Context, st Settings, act Actions) {
	s := st.Draft
	section(c, locales.Appearance(), func() {
		field(c, locales.Theme(), locales.ThemeNote(), func() {
			options := []option{
				{"system", locales.FollowSystem()},
				{"light", locales.Light()},
				{"dark", locales.Dark()},
			}
			for _, name := range theme.UIThemes() {
				options = append(options, option{name, name})
			}
			choice(c, &s.Theme, options)
		})
		field(c, locales.BackgroundOpacity(), locales.BackgroundOpacityNote(), func() {
			ratio(c, &s.BackgroundOpacity, 0.3, 1, 0.05)
		})
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
		field(c, locales.AlwaysOnTop(), locales.AlwaysOnTopNote(), func() {
			myui.Switch(c, &s.AlwaysOnTop)
		})
		field(c, locales.TitleFollowsPane(), locales.TitleFollowsPaneNote(), func() {
			myui.Switch(c, &s.TitleFromActiveTerminal)
		})
	})
}

// profilesPage is the settings a profile starts from, and the profiles.
// fontOptions are the font setting's choices: the generic family every system
// resolves to a monospace font, then the families installed, each called by
// its own name.
func fontOptions(families []string) []option {
	out := make([]option, 0, len(families)+1)
	out = append(out, option{"monospace", "monospace"})
	for _, f := range families {
		out = append(out, option{f, f})
	}
	return out
}

func profilesPage(c *myui.Context, st Settings, act Actions) {
	s := st.Draft
	t := c.Theme()
	section(c, locales.Defaults(), func() {
		field(c, locales.Font(), locales.FontNote(), func() {
			choice(c, &s.Font.Family, fontOptions(st.FontFamilies))
		})
		field(c, locales.FontSize(), locales.FontSizeNote(), func() {
			measure(c, &s.Font.Size, 6, 48, 1)
		})
		field(c, locales.LineHeight(), locales.LineHeightNote(), func() {
			measure(c, &s.Font.LineHeight, 0.8, 2.4, 0.1)
		})
		field(c, locales.Cursor(), "", func() {
			myui.Row(c).AlignItems(myui.Center).Gap(12).Children(func() {
				choice(c, &s.CursorShape, []option{
					{"bar", locales.Bar()},
					{"block", locales.Block()},
					{"underline", locales.Underline()},
				})
				myui.Switch(c, &s.CursorBlink)
				myui.Text(c, locales.Blink()).TextColor(t.TextMuted)
			})
		})
		field(c, locales.Padding(), locales.PaddingNote(), func() {
			myui.Row(c).AlignItems(myui.Center).Gap(8).Children(func() {
				measure(c, &s.Padding.X, 0, 64, 1)
				myui.Text(c, "\u00d7").TextColor(t.TextMuted)
				measure(c, &s.Padding.Y, 0, 64, 1)
			})
		})
		field(c, locales.Scrollback(), locales.ScrollbackNote(), func() {
			count(c, &s.ScrollbackMB, 0, 1024)
		})
	})
	section(c, locales.Profiles(), func() {
		for _, p := range st.Profiles {
			row(c, func() {
				myui.Row(c).Height(32).AlignItems(myui.Center).Gap(8).Children(func() {
					name := myui.Text(c, p.Name).Grow(1)
					if p.Hidden {
						name.TextColor(t.TextMuted)
					}
					if p.Discovered {
						myui.Text(c, locales.BuiltIn()).FontSize(11).TextColor(t.TextMuted).
							Background(t.Surface).Padding(0, 6).Radius(4)
					}
					myui.Button(c, locales.Duplicate()).OnClick(func() { act.DuplicateProfile(p.ID) })
					switch {
					case p.Hidden:
						myui.Button(c, locales.Show()).OnClick(func() { act.ShowProfile(p.ID) })
					case p.Discovered:
						// The window finds this shell again on the next start, so
						// the page offers to hide it rather than to remove it.
						myui.Button(c, locales.Hide()).OnClick(func() { act.RemoveProfile(p.ID) })
					default:
						myui.Button(c, locales.Remove()).OnClick(func() { act.RemoveProfile(p.ID) })
					}
				})
			})
		}
	})
}

// profilePage is the profile with an id, as the copy the page edits.
func profilePage(c *myui.Context, st Settings, act Actions, id string) {
	p := st.Profile
	if p == nil || p.ID != id {
		myui.Text(c, locales.SelectProfile()).TextColor(c.Theme().TextMuted)
		return
	}
	section(c, p.Name, func() {
		field(c, locales.Name(), "", func() {
			myui.TextInput(c, &p.Name)
		})
		field(c, locales.CommandLine(), locales.CommandLineNote(), func() {
			myui.TextInput(c, &p.Command).Placeholder(store.DefaultShellHint())
		})
		field(c, locales.Arguments(), locales.ArgumentsNote(), func() {
			myui.TextArea(c, st.Args).Lines(2, 6)
		})
		field(c, locales.StartDirectory(), locales.ProfileStartNote(), func() {
			myui.TextInput(c, &p.Dir).Placeholder(store.DefaultHomeHint())
		})
		field(c, locales.IconColor(), locales.IconColorNote(), func() {
			colorField(c, &p.Color)
		})
		field(c, locales.Hidden(), locales.HiddenNote(), func() {
			myui.Switch(c, &p.Hidden)
		})
	})
	myui.Row(c).Gap(8).Children(func() {
		myui.Button(c, locales.Duplicate()).OnClick(func() { act.DuplicateProfile(p.ID) })
		myui.Button(c, locales.Remove()).OnClick(func() { act.RemoveProfile(p.ID) })
	})
}
