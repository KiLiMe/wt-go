package ui

import (
	"sort"
	"strings"

	myui "github.com/egoist/mygo/ui"

	"wtgo/internal/i18n"
	"wtgo/internal/icons"
	"wtgo/internal/profile"
	"wtgo/internal/store"
	"wtgo/internal/theme"
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

// settingHit is one setting the column searches: what the page labels it, the
// page it is on, the page's name to say where it is, and the words that find
// it besides its own name -- the field's key in settings.json, and the word a
// person may know it by in either language.
type settingHit struct {
	name     string
	page     string
	pageName string
	words    string
}

// searchSettings is what the settings page searches for, "" for nothing.
// searchSettings is what the box over the column searches by, lowered and
// trimmed so that a stray space or a capital letter does not hide a setting.
func searchSettings(st Settings) string {
	if st.Search == nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(*st.Search))
}

// settingsIndex is every setting the column searches: what it is called, the
// page it is on, and the words that find it. The names are what the pages
// label the fields with, made here rather than at the call of the search, so
// that a language switch renames the index with the pages.
func settingsIndex() []settingHit {
	page := func(p string) string {
		switch p {
		case PageGeneral:
			return locales.General()
		case PageInteraction:
			return locales.Interaction()
		case PageAppearance:
			return locales.Appearance()
		case PageActions:
			return locales.Actions()
		case PageAdvanced:
			return locales.Advanced()
		default:
			return locales.Profiles()
		}
	}
	var out []settingHit
	add := func(p, name, words string) {
		out = append(out, settingHit{name: name, page: p, pageName: page(p), words: words})
	}
	add(PageGeneral, locales.Language(), "language locale i18n 语言 界面")
	add(PageGeneral, locales.DefaultProfile(), "profile shell startup default 默认 配置文件 启动")
	add(PageGeneral, locales.StartDirectory(), "directory cwd folder start 起始目录 目录")
	add(PageGeneral, locales.NewTabPosition(), "tab position order 新标签 位置")
	add(PageGeneral, locales.TabWidth(), "tab width size 标签 宽度")
	add(PageGeneral, locales.TitleFollowsPane(), "title rename terminal 标题 跟随")
	add(PageGeneral, locales.AlwaysOnTop(), "window always on top 置顶 窗口")
	add(PageGeneral, locales.PanesPerTab(), "pane split max count 窗格 上限")
	add(PageGeneral, locales.CloseOnLastTab(), "close tab window 关闭 标签")
	add(PageGeneral, locales.ConfirmClose(), "confirm close ask 确认 关闭")
	add(PageGeneral, locales.ConfirmCloseWindow(), "confirm window close ask 确认 关闭 窗口")

	add(PageAppearance, locales.Theme(), "theme dark light system 主题 界面")
	add(PageAppearance, locales.BackgroundOpacity(), "opacity translucent transparency background 透明度 背景")
	add(PageAppearance, locales.Font(), "font family typeface 字体")
	add(PageAppearance, locales.FontSize(), "font size 字号 字体")
	add(PageAppearance, locales.LineHeight(), "line height spacing 行高")
	add(PageAppearance, locales.Padding(), "padding space 内边距")
	add(PageAppearance, locales.PaddingColor(), "padding color 内边距 颜色")
	add(PageAppearance, locales.Cursor(), "cursor shape blink 光标")
	add(PageAppearance, locales.MinimumContrast(), "contrast 对比度")
	add(PageAppearance, locales.Scheme(), "scheme color colours palette theme 配色 方案 颜色")

	add(PageInteraction, locales.CopyOnSelect(), "copy select mouse clipboard 复制 选中 鼠标")
	add(PageInteraction, locales.RightClick(), "right click mouse menu paste 右键 鼠标")
	add(PageInteraction, locales.ScrollMultiplier(), "scroll wheel mouse speed 滚轮 滚动")
	add(PageInteraction, locales.ClipboardRead(), "clipboard read paste osc52 剪贴板 读取 粘贴")
	add(PageInteraction, locales.ClipboardWrite(), "clipboard write copy osc52 剪贴板 写入 复制")
	add(PageInteraction, locales.PasteMultiline(), "paste multiline confirm clipboard 粘贴 多行")
	add(PageInteraction, locales.Scrollback(), "scrollback buffer history lines 回滚 缓冲")

	add(PageProfiles, locales.Profiles(), "profile shell command 配置文件")
	add(PageProfiles, locales.Name(), "name rename profile 名称")
	add(PageProfiles, locales.CommandLine(), "command line exec shell 命令行")
	add(PageProfiles, locales.Arguments(), "arguments args 参数")
	add(PageProfiles, locales.StartDirectory(), "directory cwd profile 起始目录")
	add(PageProfiles, locales.IconColor(), "color icon 图标 颜色")
	add(PageProfiles, locales.Hidden(), "hide hidden 隐藏")

	add(PageActions, locales.KeySearch(), "keybinding shortcut key hotkey 快捷键 键位")
	add(PageActions, locales.RestoreDefaults(), "reset keybindings restore default 恢复 重置 默认")
	add(PageActions, locales.OnlyChanged(0), "changed modified 已修改")

	add(PageAdvanced, locales.ProbeAgain(), "probe rediscover shell detect 重新探测 探测")
	add(PageAdvanced, locales.SettingsFolder(), "settings folder file path 设置目录")
	add(PageAdvanced, locales.ProgramFolder(), "program folder path 程序目录")
	add(PageAdvanced, locales.Reset(), "reset delete settings 重置 删除")
	return out
}

// searchSettingsIndex is the settings that answer what is searched for: the
// ones whose name starts with the word, then the ones whose name holds it,
// then the ones a word about them names.
func searchSettingsIndex(words string) []settingHit {
	words = strings.ToLower(strings.TrimSpace(words))
	if words == "" {
		return nil
	}
	type scored struct {
		hit   settingHit
		score int
	}
	var hits []scored
	for _, h := range settingsIndex() {
		name := strings.ToLower(h.name)
		score := -1
		switch {
		case name == words:
			score = 0
		case strings.HasPrefix(name, words):
			score = 10
		case strings.Contains(name, words):
			score = 20
		case strings.Contains(" "+h.words+" ", words):
			score = 30
		}
		if score < 0 {
			continue
		}
		hits = append(hits, scored{h, score})
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].score < hits[j].score })
	out := make([]settingHit, 0, len(hits))
	for _, h := range hits {
		out = append(out, h.hit)
	}
	return out
}

// settingsHitRow is one setting of the column while the search is on: what it
// is called, the page that holds it under it, and a click that opens that page.
func settingsHitRow(c *myui.Context, h settingHit, act Actions) {
	t := c.Theme()
	el := myui.Column(c).Padding(4, 8).Gap(1).Radius(6).
		Cursor(myui.CursorPointer).OnClick(func() { act.SelectSettingsPage(h.page) })
	if el.Hovered() {
		el.Background(t.SurfaceHover)
	}
	el.Children(func() {
		myui.Text(c, h.name).FontSize(13).SingleLine().Ellipsis("…")
		myui.Text(c, h.pageName).FontSize(11).TextColor(t.TextMuted).SingleLine().Ellipsis("…")
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

// appearancePage is how the window and the terminals look: a terminal drawn as
// the draft draws it, held at the head of the page while the face, the
// typesetting, the cursor, and the colors it is drawn with scroll under it.
func appearancePage(c *myui.Context, st Settings, act Actions) {
	s := st.Draft
	pinnedSection(c, locales.AppearancePreview(), func() {
		preview(c, s)
	})
	section(c, locales.AppearanceFace(), func() {
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
	})
	section(c, locales.AppearanceType(), func() {
		field(c, locales.Font(), locales.FontNote(), func() {
			choice(c, &s.Font.Family, fontOptions(st.FontFamilies))
		})
		field(c, locales.FontSize(), locales.FontSizeNote(), func() {
			measure(c, &s.Font.Size, 6, 48, 1)
		})
		field(c, locales.LineHeight(), locales.LineHeightNote(), func() {
			measure(c, &s.Font.LineHeight, 0.8, 2.4, 0.1)
		})
		field(c, locales.Padding(), locales.PaddingNote(), func() {
			myui.Row(c).AlignItems(myui.Center).Gap(8).Children(func() {
				measure(c, &s.Padding.X, 0, 64, 1)
				myui.Text(c, "\u00d7").TextColor(c.Theme().TextMuted)
				measure(c, &s.Padding.Y, 0, 64, 1)
			})
		})
		field(c, locales.PaddingColor(), locales.PaddingColorNote(), func() {
			choice(c, &s.PaddingColor, []option{
				{"background", locales.WindowBackground()},
				{"extend", locales.NearbyCells()},
				{"always", locales.AlwaysCells()},
			})
		})
	})
	section(c, locales.AppearanceCursor(), func() {
		field(c, locales.Cursor(), locales.CursorNote(), func() {
			myui.Row(c).AlignItems(myui.Center).Gap(12).Children(func() {
				choice(c, &s.CursorShape, []option{
					{"bar", locales.Bar()},
					{"block", locales.Block()},
					{"underline", locales.Underline()},
				})
				myui.Switch(c, &s.CursorBlink)
				myui.Text(c, locales.Blink()).TextColor(c.Theme().TextMuted)
			})
		})
		field(c, locales.MinimumContrast(), locales.MinimumContrastNote(), func() {
			measure(c, &s.MinimumContrast, 0, 21, 0.5)
		})
	})
	section(c, locales.ColorSchemes(), func() {
		schemePicker(c, s, st, act)
	})
}

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
				Radius(10).Padding(16).Children(func() {
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
	for _, p := range listedProfiles(st) {
		return p.ID, true
	}
	return "", false
}

// listedProfiles are the profiles the list shows: what the search keeps of the
// profiles the settings hold, in the order of the file.
func listedProfiles(st Settings) []Profile {
	words := strings.ToLower(strings.TrimSpace(searchText(st.ProfileFilter)))
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
		BorderWidth(1).BorderColor(cardLine(t)).Radius(10).Children(func() {
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
		field(c, locales.Name(), "", func() {
			myui.TextInput(c, &p.Name)
		})
		if readonly {
			fieldBlock(c, locales.CommandLine(), locales.DiscoveredProfileNote(), func() {
				myui.TextInput(c, &p.Command).ReadOnly(true)
			})
		} else {
			fieldBlock(c, locales.CommandLine(), locales.CommandLineNote(), func() {
				myui.TextInput(c, &p.Command).Placeholder(store.DefaultShellHint())
			})
			// Below the field rather than beside it: the note says what
			// emptying the field does, and the hint says why the empty one a
			// person just left is not what runs yet.
			if strings.TrimSpace(p.Command) == "" {
				myui.Column(c).Gap(3).MinWidth(0).Children(func() {
					myui.Text(c, locales.CommandEmptyHint(store.DefaultShellHint())).FontSize(12).TextColor(c.Theme().TextMuted).Wrap()
				})
			}
		}
		fieldBlock(c, locales.Arguments(), locales.ArgumentsNote(), func() {
			myui.TextArea(c, st.Args).Lines(2, 6)
		})
		fieldBlock(c, locales.StartDirectory(), locales.ProfileStartNote(), func() {
			myui.TextInput(c, &p.Dir).Placeholder(store.DefaultHomeHint())
		})
		fieldBlock(c, locales.Icon(), locales.IconNote(), func() {
			iconField(c, st, act, p)
		})
		field(c, locales.IconColor(), locales.IconColorNote(), func() {
			colorField(c, &p.Color)
		})
		field(c, locales.Hidden(), locales.HiddenNote(), func() {
			myui.Switch(c, &p.Hidden)
		})
	})
	t := c.Theme()
	myui.Row(c).Padding(12, 16).Gap(8).BorderWidth(1, 0, 0, 0).BorderColor(cardLine(t)).Children(func() {
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
