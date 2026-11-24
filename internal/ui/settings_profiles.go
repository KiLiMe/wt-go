package ui

import (
	myui "github.com/egoist/mygo/ui"
	"strings"
	"wtgo/internal/icons"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/locales"
)

// profilesPage is the settings a profile starts from, and the profiles: the
// list of profiles on the left and the fields of the one in front on the
// right. How the window and the terminals look is the appearance page's.
func profilesPage(c *myui.Context, st Settings, act Actions) {
	t := c.Theme()
	// The list takes the height of the fields beside it and scrolls with the
	// page when it is longer than they are: a page of twenty profiles grows
	// rather than crops, which is what a height of the card's own did.
	myui.Row(c).Grow(1).MinHeight(0).AlignItems(myui.Stretch).Gap(24).Children(func() {
		myui.Column(c).Width(profileListWidth).Shrink(0).Children(func() {
			profileList(c, st, act)
		})
		myui.Column(c).Grow(1).MinWidth(0).Children(func() {
			if id, ok := profileInFront(st); ok {
				profilePage(c, st, act, id)
				return
			}
			myui.Column(c).Background(cardFace(t)).BorderWidth(1).BorderColor(cardLine(t)).
				Radius(cardRadius).Padding(16).Children(func() {
				myui.Text(c, locales.SelectProfile()).FontSize(12).TextColor(t.TextMuted)
			})
		})
	})
}

// profileListWidth is how wide the list of profiles is: enough for a name and
// the badge beside it, and no more, so that the fields keep the page.
const profileListWidth = 250

// profileInFront is the profile the fields on the right edit. A page that
// names none, which is what a window shows while it has none open, shows the
// first one it lists: a list beside an empty half is a page of nothing.
func profileInFront(st Settings) (string, bool) {
	if id, ok := ProfileID(st.Page); ok && st.Profile != nil && st.Profile.ID == id {
		return id, true
	}
	if listed := listedProfiles(st); len(listed) > 0 {
		return listed[0].ID, true
	}
	return "", false
}

// listedProfiles are the profiles the list shows: what the search keeps of the
// profiles the settings hold, in the order of the file.
func listedProfiles(st Settings) []Profile {
	words := strings.ToLower(strings.TrimSpace(text(st.ProfileFilter)))
	if words == "" {
		return st.Profiles
	}
	out := make([]Profile, 0, len(st.Profiles))
	for _, p := range st.Profiles {
		if strings.Contains(strings.ToLower(p.Name), words) {
			out = append(out, p)
		}
	}
	return out
}

// profileList is the left column of the page: what adds a profile, the words
// the list is searched for, and the profiles themselves, under a scroll of
// their own. Hidden ones stay in the list, marked: the page edits hiding in the
// fields beside it, so the list has to keep the way back to what it hid.
func profileList(c *myui.Context, st Settings, act Actions) {
	t := c.Theme()
	// Grow, not a height of its own: the card takes the height of the fields
	// beside it, and a list taller than they are grows the page instead of
	// being cropped by a card too short to hold it.
	myui.Column(c).Grow(1).MinHeight(0).Background(cardFace(t)).
		BorderWidth(1).BorderColor(cardLine(t)).Radius(cardRadius).Children(func() {
		myui.Column(c).Padding(8).Gap(8).Children(func() {
			myui.Button(c, locales.AddProfile()).Grow(1).OnClick(act.AddProfile)
			if st.ProfileFilter != nil {
				myui.TextInput(c, st.ProfileFilter).Placeholder(locales.ProfileSearch())
			}
		})
		myui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
			myui.Column(c).Padding(0, 0, 6, 0).Children(func() {
				listed := listedProfiles(st)
				if len(listed) == 0 {
					myui.Text(c, locales.ProfileNone()).FontSize(12).TextColor(t.TextMuted).Padding(10, 10)
				}
				for _, p := range listed {
					profileRow(c, st, act, p)
				}
			})
		})
	})
}

// profileRow is one profile of the list: a click opens it in the fields on the
// right, and the row in front says so with a surface of its own.
func profileRow(c *myui.Context, st Settings, act Actions, p Profile) {
	t := c.Theme()
	lit := st.Profile != nil && st.Profile.ID == p.ID
	el := myui.Row(c).Height(30).Padding(0, 10).AlignItems(myui.Center).Gap(6).Radius(6).
		Cursor(myui.CursorPointer).OnClick(func() { act.SelectSettingsPage(ProfilePage(p.ID)) })
	if lit {
		el.Background(t.SurfaceHover)
	} else if el.Hovered() {
		el.Background(t.Surface)
	}
	el.Children(func() {
		// The mark the tab will wear, at the size the tab wears it: the list is
		// where a profile is chosen, and two profiles of the same shell are the
		// case the mark is for.
		markIcon(c, p.Icon, p.Color, 16, true)
		name := myui.Text(c, p.Name).FontSize(13).SingleLine().Shrink(1).MinWidth(0)
		if p.Hidden {
			// A profile the page hides reads as one it does not offer, but still
			// takes a click: that is how it comes back.
			name.TextColor(t.TextMuted)
		} else {
			name.FontWeight(500)
		}
		if !p.Discovered && !p.Hidden {
			return
		}
		// The badges sit at the end of the row rather than behind the name: a
		// column of them reads as a column, and a long name does not push the
		// word it is marked with out of the row.
		myui.Spacer(c)
		if p.Discovered {
			myui.Text(c, locales.BuiltIn()).FontSize(10).TextColor(t.TextMuted).
				Background(t.Surface).Padding(0, 5).Radius(4).Shrink(0)
		}
		if p.Hidden {
			// The list keeps a hidden profile, so it has to say so: the row is
			// otherwise a profile the new tab menu would not offer.
			myui.Text(c, locales.HiddenProfile()).FontSize(10).TextColor(t.TextMuted).
				Background(t.Surface).Padding(0, 5).Radius(4).Shrink(0)
		}
	})
}

// profilePage is the profile with an id, as the copy the page edits. The
// profile the window found runs a shell it knows the path of, so its command is
// not the page's to change.
func profilePage(c *myui.Context, st Settings, act Actions, id string) {
	p := st.Profile
	if p == nil || p.ID != id {
		myui.Text(c, locales.SelectProfile()).TextColor(c.Theme().TextMuted)
		return
	}
	// The window finds this profile again on the next start, so a command it
	// changed here would be a change the page could not keep.
	readonly := p.Discovered
	section(c, "", func() {
		field(c, "name", locales.Name(), "", func() {
			myui.TextInput(c, &p.Name)
		}, st.Reveal)
		if readonly {
			fieldBlock(c, "command", locales.CommandLine(), locales.DiscoveredProfileNote(), func() {
				myui.TextInput(c, &p.Command).ReadOnly(true).Grow(1)
			}, st.Reveal)
		} else {
			fieldBlock(c, "command", locales.CommandLine(), locales.CommandLineNote(), func() {
				myui.TextInput(c, &p.Command).Placeholder(store.DefaultShellHint()).Grow(1)
			}, st.Reveal)
			// Below the field rather than beside it: the note says what
			// emptying the field does, and the hint says why the empty one a
			// person just left is not what runs yet.
			if strings.TrimSpace(p.Command) == "" {
				myui.Column(c).Gap(3).MinWidth(0).Children(func() {
					myui.Text(c, locales.CommandEmptyHint(store.DefaultShellHint())).FontSize(12).TextColor(c.Theme().TextMuted).Wrap()
				})
			}
		}
		fieldBlock(c, "args", locales.Arguments(), locales.ArgumentsNote(), func() {
			myui.TextArea(c, st.Args).Lines(2, 6).Grow(1)
		}, st.Reveal)
		fieldBlock(c, "dir", locales.StartDirectory(), locales.ProfileStartNote(), func() {
			myui.TextInput(c, &p.Dir).Placeholder(store.DefaultHomeHint()).Grow(1)
		}, st.Reveal)
		fieldBlock(c, "icon", locales.Icon(), locales.IconNote(), func() {
			iconField(c, st, act, p)
		}, st.Reveal)
		field(c, "color", locales.IconColor(), locales.IconColorNote(), func() {
			colorField(c, &p.Color)
		}, st.Reveal)
		field(c, "hidden", locales.Hidden(), locales.HiddenNote(), func() {
			myui.Switch(c, &p.Hidden)
		}, st.Reveal)
	})
	// No line over this row: it stands under the card of fields rather than
	// inside it, and a hairline there draws a second edge a few pixels under
	// the card's own, one that runs past the card's rounded corner.
	myui.Row(c).Padding(12, 16).Gap(8).Children(func() {
		myui.Button(c, locales.Duplicate()).OnClick(func() { act.DuplicateProfile(p.ID) })
		if p.Discovered {
			// The machine is the one thing that can say what this profile is now,
			// so the button that asks it sits here rather than on a page about
			// every profile at once (the advanced page's probe).
			myui.Button(c, locales.ResetProfile()).OnClick(func() { act.ResetProfile(p.ID) })
		}
		// Hiding is the whole of what a removal can do to a shell the window
		// finds again on the next start, so that profile is offered the hiding
		// alone rather than a delete that would not hold.
		myui.Button(c, locales.Hidden()).OnClick(func() { act.HideProfile(p.ID) })
		if !p.Discovered {
			myui.Button(c, locales.Remove()).OnClick(func() { act.RemoveProfile(p.ID) })
		}
	})
}

// iconField is the mark of the profile the page edits: the built-in shapes,
// the automatic one first, and a box for one character of the person's own.
//
// The squares are drawn at the size a tab has, so that the page shows what the
// strip will show: a preview at the size of the icon's own file would promise
// detail that 16 DIP has no room for. A preset paints its own face, which is
// why the color beside it only holds for the other two kinds of mark.
func iconField(c *myui.Context, st Settings, act Actions, p *store.Profile) {
	t := c.Theme()
	glyph := ""
	if st.Glyph != nil {
		glyph = profile.Glyph(*st.Glyph)
	}
	// Which square is lit: a character in the box is a mark of its own, so
	// none of them is.
	chosen := func(key string) bool { return glyph == "" && profile.IconKey(p.Icon) == key }
	chip := func(key, label string, ic *icons.Icon, picture bool) {
		el := myui.Box(c).Size(24, 24).Shrink(0).Radius(6).Center().
			BorderWidth(1).BorderColor(t.Border).Cursor(myui.CursorPointer).
			Label(label).Tooltip(label).OnClick(func() { act.SetProfileIcon(key) })
		if chosen(key) {
			el.BorderWidth(2).BorderColor(t.Accent)
		}
		el.Children(func() {
			if picture {
				// A preset brings its own colors; Icon would flatten them
				// into the text color, which is how a row of white squares
				// shipped once (ADR-76).
				myui.Image(c, ic.SVG).Size(16, 16)
				return
			}
			myui.Icon(c, ic.SVG).Size(16, 16).TextColor(t.TextMuted)
		})
	}
	// Wrapped, because twelve shapes and a box are wider than the column
	// the field is written in, and a row of chips that runs past the card
	// reads as a broken card rather than as a long list.
	myui.Row(c).Wrap().Gap(6).AlignItems(myui.Center).Children(func() {
		auto := profile.GuessIcon(p.Command)
		ic := icons.Preset(auto)
		picture := ic != nil
		if ic == nil {
			ic = icons.Terminal
		}
		chip("", locales.IconAuto(), ic, picture)
		for _, pr := range profile.Presets() {
			chip(pr.Key, pr.Shape+" \u00b7 "+pr.Name, icons.Preset(pr.Key), true)
		}
		if st.Glyph != nil {
			myui.TextInput(c, st.Glyph).Width(88).Placeholder(locales.IconGlyph())
		}
	})
}
