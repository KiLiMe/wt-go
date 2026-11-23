package ui

import (
	myui "github.com/egoist/mygo/ui"
	"sort"
	"strings"
	"wtgo/locales"
)

// settingHit is one setting the column searches: what the page labels it, the
// setting itself by the name the file gives it, the page it is on, the page's
// name to say where it is, and the words that find it besides its own name --
// the field's key in settings.json, and the word a person may know it by in
// either language.
type settingHit struct {
	name     string
	key      string
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
	// key is the setting the file writes, which is how the page finds the row to
	// mark and scroll to; "" is a setting the page holds as a whole, as its
	// list of keybindings is.
	add := func(p, key, name, words string) {
		out = append(out, settingHit{name: name, key: key, page: p, pageName: page(p), words: words})
	}
	add(PageGeneral, "language", locales.Language(), "language locale i18n 语言 界面")
	add(PageGeneral, "defaultProfile", locales.DefaultProfile(), "profile shell startup default 默认 配置文件 启动")
	add(PageGeneral, "startDirectory", locales.StartDirectory(), "directory cwd folder start 起始目录 目录")
	add(PageGeneral, "newTabPosition", locales.NewTabPosition(), "tab position order 新标签 位置")
	add(PageGeneral, "tabWidth", locales.TabWidth(), "tab width size 标签 宽度")
	add(PageGeneral, "titleFromActiveTerminal", locales.TitleFollowsPane(), "title rename terminal 标题 跟随")
	add(PageGeneral, "alwaysOnTop", locales.AlwaysOnTop(), "window always on top 置顶 窗口")
	add(PageGeneral, "maxPanes", locales.PanesPerTab(), "pane split max count 窗格 上限")
	add(PageGeneral, "closeOnLastTab", locales.CloseOnLastTab(), "close tab window 关闭 标签")
	add(PageGeneral, "confirmClose", locales.ConfirmClose(), "confirm close ask 确认 关闭")
	add(PageGeneral, "confirmCloseWindow", locales.ConfirmCloseWindow(), "confirm window close ask 确认 关闭 窗口")

	add(PageAppearance, "theme", locales.Theme(), "theme dark light system 主题 界面")
	add(PageAppearance, "backgroundOpacity", locales.BackgroundOpacity(), "opacity translucent transparency background 透明度 背景")
	add(PageAppearance, "font", locales.Font(), "font family typeface 字体")
	add(PageAppearance, "font.size", locales.FontSize(), "font size 字号 字体")
	add(PageAppearance, "font.lineHeight", locales.LineHeight(), "line height spacing 行高")
	add(PageAppearance, "padding", locales.Padding(), "padding space 内边距")
	add(PageAppearance, "paddingColor", locales.PaddingColor(), "padding color 内边距 颜色")
	add(PageAppearance, "cursorShape", locales.Cursor(), "cursor shape blink 光标")
	add(PageAppearance, "pointerText", locales.Pointer(), "pointer cursor arrow ibeam text 指针 箭头 光标")
	add(PageAppearance, "minimumContrast", locales.MinimumContrast(), "contrast 对比度")
	add(PageAppearance, "colorScheme", locales.Scheme(), "scheme color colours palette theme 配色 方案 颜色")

	add(PageInteraction, "copyOnSelect", locales.CopyOnSelect(), "copy select mouse clipboard 复制 选中 鼠标")
	add(PageInteraction, "rightClick", locales.RightClick(), "right click mouse menu paste 右键 鼠标")
	add(PageInteraction, "rightClickCopyPastePriority", locales.RightClickCopyPastePriority(), "right click mouse tracking copy paste 右键 鼠标 上报 复制 粘贴")
	add(PageInteraction, "links", locales.Links(), "link click mouse open path url hyperlink 链接 打开 路径 网址 超链接")
	add(PageInteraction, "linksProgram", locales.LinksProgram(), "link hyperlink osc8 program set 链接 超链接 程序 设置")
	add(PageInteraction, "linksURL", locales.LinksURL(), "link url http mailto printed text 链接 网址 打印 文本")
	add(PageInteraction, "linksPath", locales.LinksPath(), "link path file guess line number 链接 路径 文件 行号")
	add(PageInteraction, "linksHover", locales.LinksHover(), "link ctrl hover underline pointer hand 链接 悬停 下划线 手型 指针")
	add(PageInteraction, "scrollMultiplier", locales.ScrollMultiplier(), "scroll wheel mouse speed 滚轮 滚动")
	add(PageInteraction, "clipboardRead", locales.ClipboardRead(), "clipboard read paste osc52 剪贴板 读取 粘贴")
	add(PageInteraction, "clipboardWrite", locales.ClipboardWrite(), "clipboard write copy osc52 剪贴板 写入 复制")
	add(PageInteraction, "pasteMultilineConfirm", locales.PasteMultiline(), "paste multiline confirm clipboard 粘贴 多行")
	add(PageInteraction, "scrollbackMB", locales.Scrollback(), "scrollback buffer history lines 回滚 缓冲")

	add(PageProfiles, "", locales.Profiles(), "profile shell command 配置文件")
	add(PageProfiles, "name", locales.Name(), "name rename profile 名称")
	add(PageProfiles, "command", locales.CommandLine(), "command line exec shell 命令行")
	add(PageProfiles, "args", locales.Arguments(), "arguments args 参数")
	add(PageProfiles, "dir", locales.StartDirectory(), "directory cwd profile 起始目录")
	add(PageProfiles, "color", locales.IconColor(), "color icon 图标 颜色")
	add(PageProfiles, "hidden", locales.Hidden(), "hide hidden 隐藏")

	add(PageActions, keySearchActions, locales.KeySearch(), "keybinding shortcut key hotkey 快捷键 键位")
	add(PageActions, keyRestoreDefaults, locales.RestoreDefaults(), "reset keybindings restore default 恢复 重置 默认")
	add(PageActions, keyOnlyChanged, locales.OnlyChanged(0), "changed modified 已修改")

	add(PageAdvanced, "probe", locales.ProbeAgain(), "probe rediscover shell detect 重新探测 探测")
	add(PageAdvanced, "settingsFolder", locales.SettingsFolder(), "settings folder file path 设置目录")
	add(PageAdvanced, "programFolder", locales.ProgramFolder(), "program folder path 程序目录")
	add(PageAdvanced, "reset", locales.Reset(), "reset delete settings 重置 删除")
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
// is called, the page that holds it under it, and a click that opens that page
// and marks the setting on it.
func settingsHitRow(c *myui.Context, h settingHit, act Actions) {
	t := c.Theme()
	el := myui.Column(c).Padding(4, 8).Gap(1).Radius(6).
		Cursor(myui.CursorPointer).OnClick(func() { act.RevealSetting(h.page, h.key) })
	if el.Hovered() {
		el.Background(t.SurfaceHover)
	}
	el.Children(func() {
		myui.Text(c, h.name).FontSize(13).SingleLine().Ellipsis("…")
		myui.Text(c, h.pageName).FontSize(11).TextColor(t.TextMuted).SingleLine().Ellipsis("…")
	})
}
