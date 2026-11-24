package store

import (
	"regexp"
)

// Boundary: the words a file may say for each setting, and the small pure
// questions asked of them -- is this one of them, is that language a tag, is
// this number in range. Which settings are in range is validate.go.

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
