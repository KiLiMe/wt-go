package ui

import (
	myui "github.com/egoist/mygo/ui"
	"strings"
	"wtgo/locales"
)

// The pages the settings tab shows in its left column.
const (
	PageGeneral     = "general"
	PageInteraction = "interaction"
	PageAppearance  = "appearance"
	PageSchemes     = "schemes"
	PageActions     = "actions"
	PageAdvanced    = "advanced"
	PageProfiles    = "profiles"
)

// PageStartup is the name the page of the window's own behavior had before it
// took in the tabs and the closing questions: the window still opens there by
// default, and a page kept across a build would name it, so both names show
// the same page.
const PageStartup = PageGeneral

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
			myui.Column(c).Grow(1).Children(func() {
				banner(c, st)
				myui.Scroll(c).Grow(1).Children(func() {
					myui.Column(c).Padding(24, 24, 16, 24).Gap(24).Children(func() {
						pageIntro(c, st)
						page(c, st, act)
					})
				})
			})
		})
		footer(c, st, act)
	})
}

// navWidth is how wide the column of pages is, and where the page's own room
// starts.
const navWidth = 220

// nav is the column that lists the pages: the box that searches the settings
// over it, and the pages, or what the search finds on them.
func nav(c *myui.Context, st Settings, act Actions) myui.Element {
	t := c.Theme()
	// The column is the page's own color, not a surface of its own: a band of
	// Surface beside the page read as a second window, and the tab in front
	// already carries the background down from the strip.
	return myui.Column(c).Width(navWidth).Padding(12, 8, 12, 8).Gap(2).
		BorderWidth(0, 1, 0, 0).BorderColor(cardLine(t)).Children(func() {
		if st.Search != nil {
			myui.TextInput(c, st.Search).Placeholder(locales.SearchSettings())
		}
		// The column is the pages and nothing else: a profile is not a page of
		// its own any more but a row of the profiles page, which lists them
		// beside the one it edits, so that adding one never grows the column.
		myui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
			myui.Column(c).Gap(2).Children(func() {
				if words := searchSettings(st); words != "" {
					hits := searchSettingsIndex(words)
					if len(hits) == 0 {
						myui.Text(c, locales.NoMatchingSettings()).FontSize(12).
							TextColor(t.TextMuted).Padding(4, 8).Wrap()
					}
					for _, h := range hits {
						settingsHitRow(c, h, act)
					}
					return
				}
				for _, it := range []struct{ page, label string }{
					{PageGeneral, locales.General()},
					{PageInteraction, locales.Interaction()},
					{PageAppearance, locales.Appearance()},
					{PageProfiles, locales.Profiles()},
					{PageActions, locales.Actions()},
					{PageAdvanced, locales.Advanced()},
				} {
					// A profile page is this page with one profile in front, so the
					// column lights that page for it.
					navItem(c, topPage(st.Page), it.page, it.label, act)
				}
			})
		})
	})
}

// topPage is the page of the column a page of the settings belongs to: a
// profile page is the profiles page with the profile's id after it.
func topPage(page string) string {
	if _, ok := ProfileID(page); ok {
		return PageProfiles
	}
	return page
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

// banner is why the last save refused the draft, above the page it refused it
// on: the footer's note scrolls with nothing and cuts a long reason at the
// width of the row, when what is wrong is the thing to correct.
func banner(c *myui.Context, st Settings) {
	if st.Validation == "" {
		return
	}
	t := c.Theme()
	myui.Row(c).Padding(10, 24).Gap(8).AlignItems(myui.Start).
		Background(t.Background.MixOklab(t.Danger, 0.1)).Children(func() {
		myui.Text(c, st.Validation).FontSize(12).TextColor(t.Danger).Wrap()
	})
}

// footer is what saves the page: what the draft holds that the file does not,
// and the buttons that write it or drop it. The settings file itself opens
// from the left, where the row starts, so that the way to edit it by hand is
// the first thing on it.
func footer(c *myui.Context, st Settings, act Actions) myui.Element {
	t := c.Theme()
	return myui.Row(c).Height(48).Padding(0, 16).AlignItems(myui.Center).Gap(8).
		BorderWidth(1, 0, 0, 0).BorderColor(cardLine(t)).Children(func() {
		myui.Button(c, locales.OpenSettingsFile()).OnClick(act.OpenSettingsFile)
		switch {
		case st.Note != "":
			myui.Text(c, st.Note).FontSize(11).TextColor(t.TextMuted).SingleLine().MaxWidth(420)
		case st.Validation != "":
			// The banner carries the reason itself, so the row says what to do
			// about it, which is what a person reads at the button they pressed.
			myui.Text(c, st.SaveRefusedNote()).FontSize(11).TextColor(t.Danger).SingleLine().MaxWidth(420)
		case st.Dirty:
			myui.Text(c, locales.UnsavedChanges()).FontSize(11).TextColor(t.TextMuted).SingleLine().MaxWidth(420)
		}
		myui.Spacer(c)
		myui.Button(c, locales.Discard()).Disabled(!st.Dirty).Opacity(dimmed(st.Dirty)).OnClick(act.DiscardSettings)
		myui.PrimaryButton(c, locales.Save()).Disabled(!st.Dirty).Opacity(dimmed(st.Dirty)).OnClick(act.SaveSettings)
	})
}

// SaveRefusedNote is what the foot of the settings page says once a save
// refused the draft: whatever went wrong is the banner's to name, and this is
// the way out of it.
func (st Settings) SaveRefusedNote() string {
	return locales.SaveRefusedNote()
}

// dimmed is how much of a button that does nothing shows: the theme does not
// paint a disabled button differently, so the footer says it itself.
func dimmed(live bool) float32 {
	if live {
		return 1
	}
	return 0.4
}

// pageIntro is the name of the page in front and what the page is for.
func pageIntro(c *myui.Context, st Settings) {
	switch st.Page {
	case PageGeneral:
		intro(c, locales.General(), locales.GeneralNote())
	case PageInteraction:
		intro(c, locales.Interaction(), locales.InteractionNote())
	case PageAppearance:
		intro(c, locales.Appearance(), locales.AppearanceNote())
	case PageSchemes:
		// The color schemes are a section of the appearance page now, and a
		// page kept across a build still names the page they had of their own:
		// it shows where they went rather than nothing.
		intro(c, locales.Appearance(), locales.AppearanceNote())
	case PageActions:
		intro(c, locales.Actions(), locales.ActionsNote())
	case PageAdvanced:
		intro(c, locales.Advanced(), locales.AdvancedNote())
	case PageProfiles:
		// The profiles are one page with the profile in front named above the
		// fields of that one, not a list of profiles and a page each.
		if st.Profile != nil && st.Page == ProfilePage(st.Profile.ID) {
			intro(c, st.Profile.Name, "")
			return
		}
		intro(c, locales.Profiles(), locales.DefaultsNote())
	default:
		intro(c, locales.Profiles(), locales.DefaultsNote())
	}
}

// page is the fields of the page in front.
func page(c *myui.Context, st Settings, act Actions) {
	if st.Draft == nil {
		return
	}
	switch st.Page {
	case PageGeneral:
		generalPage(c, st, act)
	case PageInteraction:
		interactionPage(c, st, act)
	case PageAppearance:
		appearancePage(c, st, act)
	case PageSchemes:
		// The name the color schemes were a page of their own: what they hold
		// is the last section of the appearance page, so both names show it.
		appearancePage(c, st, act)
	case PageActions:
		actionsPage(c, st, act)
	case PageAdvanced:
		advancedPage(c, st, act)
	case PageProfiles:
		profilesPage(c, st, act)
	default:
		// A page kept across a build that this one does not know, and the
		// profile pages the column used to hold: both show the profiles page.
		profilesPage(c, st, act)
	}
}
