package store

import (
	"fmt"
	"regexp"
	"strings"
)

// Validate reports the settings that are outside the range the window can
// run with. The ranges are wider than the settings page allows: they keep a
// hand-written file from asking for a font no program can draw or a scrollback
// no machine can hold, not to enforce the page's own limits.
func (s Settings) Validate() error {
	var bad []string
	check := func(name string, ok bool, want string) {
		if !ok {
			bad = append(bad, fmt.Sprintf("%s must be %s", name, want))
		}
	}
	check("font.size", between(float64(s.Font.Size), 4, 200), "between 4 and 200")
	check("font.lineHeight", between(float64(s.Font.LineHeight), 0.5, 4), "between 0.5 and 4")
	check("padding.x", between(float64(s.Padding.X), 0, 256), "between 0 and 256")
	check("padding.y", between(float64(s.Padding.Y), 0, 256), "between 0 and 256")
	check("backgroundOpacity", between(s.BackgroundOpacity, 0, 1), "between 0 and 1")
	check("scrollbackMB", between(float64(s.ScrollbackMB), 0, 1024), "between 0 and 1024")
	check("scrollMultiplier", between(float64(s.ScrollMultiplier), 0.1, 10), "between 0.1 and 10")
	check("minimumContrast", between(float64(s.MinimumContrast), 0, 21), "between 0 and 21")
	check("paddingColor", oneOf(s.PaddingColor, "background", "extend", "always"), `"background", "extend" or "always"`)
	check("clipboardRead", oneOf(s.ClipboardRead, "deny", "allow", "ask"), `"deny", "allow" or "ask"`)
	check("clipboardWrite", oneOf(s.ClipboardWrite, "allow", "deny"), `"allow" or "deny"`)
	check("maxPanes", between(float64(s.MaxPanes), 0, 64), "between 0 and 64")
	check("confirmClose", s.ConfirmClose.Valid(), `"running", "always" or "never"`)
	if len(bad) == 0 {
		return nil
	}
	return fmt.Errorf("settings: %s", strings.Join(bad, "; "))
}

// fallbackUnknown puts back the default for each name the file wrote that no
// setting takes. A word the window does not know is a hand-written mistake, and
// docs/配置参考.md promises it is read as the default rather than refused: the
// ranges in Validate are the other case, a file the window cannot run with. The
// window never writes such a word, so this runs once over a file read from disk
// and leaves an already-clean set exactly as it was.
func (s *Settings) fallbackUnknown() {
	d := Default()
	s.Language = knownLanguage(s.Language)
	s.Theme = knownTheme(s)
	s.StartOn = known(s.StartOn, enumStartOn, d.StartOn)
	s.CursorShape = known(s.CursorShape, enumCursorShapes, d.CursorShape)
	s.PaddingColor = known(s.PaddingColor, enumPaddingColors, d.PaddingColor)
	s.RightClick = known(s.RightClick, enumRightClicks, d.RightClick)
	s.ClipboardRead = known(s.ClipboardRead, enumClipboardReads, d.ClipboardRead)
	s.ClipboardWrite = known(s.ClipboardWrite, enumClipboardWrites, d.ClipboardWrite)
	s.NewTabPosition = known(s.NewTabPosition, enumTabPositions, d.NewTabPosition)
	s.TabWidth = known(s.TabWidth, enumTabWidths, d.TabWidth)
	if !s.ConfirmClose.Valid() {
		s.ConfirmClose = d.ConfirmClose
	}
}

// enumThemes is what "theme" takes besides the name of a scheme: the appearance
// the desktop may answer with.
var enumThemes = []string{"system", "light", "dark"}

// knownTheme is the theme a file names: one of enumThemes, or the name of a
// scheme — the file's own, a built in one, or one of Ghostty's. A scheme name is
// checked against the schemes the file writes, since the rest are resolved in
// internal/theme, which this package may not import. A name neither is, and a
// theme left out of a file, is the default: "" would be a window that draws
// with the default's colors but tells internal/theme nothing about the light
// or dark it is, and the settings page shows the theme it does draw with.
func knownTheme(s *Settings) string {
	if oneOf(s.Theme, enumThemes...) {
		return s.Theme
	}
	for _, sc := range s.Schemes {
		if sc.Name == s.Theme {
			return s.Theme
		}
	}
	return Default().Theme
}

// enumStartOn is what "startOn" takes: the first profile, or a new one. Session
// restore is a stage of its own, see docs/plan.md section 5.1.
var enumStartOn = []string{StartDefault, StartNew}

// enumCursorShapes is what the terminal draws, which is also what
// docs/配置参考.md lists: the boxes of Windows Terminal are not drawn here.
var enumCursorShapes = []string{"bar", "block", "underline"}

var enumPaddingColors = []string{"background", "extend", "always"}

var enumRightClicks = []string{"copyOrPaste", "paste", "copy", "menu", "nothing"}

var enumClipboardReads = []string{"deny", "allow", "ask"}
var enumClipboardWrites = []string{"allow", "deny"}

var enumTabPositions = []string{"last", "first"}
var enumTabWidths = []string{"compact", "equal"}

// known is v when it is one of words, and def for one that is not.
func known(v string, words []string, def string) string {
	if oneOf(v, words...) {
		return v
	}
	return def
}

// knownLanguage is the tag a file may name.
//
// A language with no catalog of its own is kept: a tag is a language, which a
// person may write for a translation this window will have later, and mygo
// shows the source language for it on its own. What is not a tag at all — a
// sentence, a path, a tag with a space in it — is the system's language back.
// The legacy region names are resolved before this runs.
func knownLanguage(tag string) string {
	if tag == "" || bcp47.MatchString(tag) {
		return tag
	}
	return ""
}

// bcp47 is the shape of a language tag, as BCP 47 writes one: "en", "zh-Hans",
// "pt-BR".
var bcp47 = regexp.MustCompile(`^[a-zA-Z]{2,8}(-[a-zA-Z0-9]{1,8})*$`)

// between reports whether v is in the range lo through hi, both included.
func between(v, lo, hi float64) bool {
	return v >= lo && v <= hi
}

// oneOf reports whether v is one of the words a setting is written as.
func oneOf(v string, words ...string) bool {
	for _, w := range words {
		if v == w {
			return true
		}
	}
	return false
}
