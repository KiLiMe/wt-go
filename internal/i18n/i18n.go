// Package i18n names the messages the window says. The words live in
// ../locales, a file per language, and this package is the way the packages
// that may not import them say a message by the key it is filed under.
package i18n

import (
	mygoi18n "github.com/egoist/mygo/i18n"

	"wtgo/locales"
)

// T is the message of key in the language the window is in. A key no
// catalog has comes back as itself: a missing message reads as its key
// rather than as nothing.
func T(key string) string { return locales.Catalog.Format(Locale(), key, nil) }

// F is the message of key with the placeholders of args filled in: the
// actions that carry a number, as "Switch to Tab 3", say their titles
// through it.
func F(key string, args map[string]any) string { return locales.Catalog.Format(Locale(), key, args) }

// Has tells whether the catalogs hold key: the tests check the keys the
// actions name, which no typed accessor checks.
func Has(key string) bool {
	_, _, ok := locales.Catalog.Message("", key)
	return ok
}

// Valid tells whether tag names a language, or is "" for the system's: a
// hand-edited settings file may say anything, and the window shows the
// system's language rather than refusing to start.
func Valid(tag string) bool { return tag == "" || mygoi18n.Normalize(tag) != "" }

// Locale is the language the window is in, "" for English, the source
// language of the catalogs, until the settings name another.
func Locale() string { return mygoi18n.Locale() }

// Languages are the languages the catalogs are translated to.
func Languages() []string { return locales.Catalog.Languages() }

// Autonym is the name of locale in its own language, as "简体中文", for the
// settings to offer it by.
func Autonym(locale string) string { return mygoi18n.Autonym(locale) }
