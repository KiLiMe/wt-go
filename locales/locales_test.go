package locales_test

import (
	"encoding/json"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"wtgo/internal/action"
	"wtgo/internal/i18n"
	"wtgo/locales"
)

// source is the language the code is written in, which every other language is
// a translation of.
const source = "en"

// translated are the languages of the catalogs, other than the source one.
var translated = []string{"zh-Hans"}

// catalog is a language's file: the messages by key, and the metadata that says
// what their placeholders are by "@" key.
type catalog map[string]any

// read reads a language's file.
func read(t *testing.T, lang string) catalog {
	t.Helper()
	b, err := os.ReadFile(lang + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var c catalog
	if err := json.Unmarshal(b, &c); err != nil {
		t.Fatalf("%s.json: %v", lang, err)
	}
	return c
}

// messages are the keys a file holds a message for, the metadata aside.
func (c catalog) messages() []string {
	var out []string
	for k, v := range c {
		if _, ok := v.(string); ok && !strings.HasPrefix(k, "@") {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

// placeholders are the names the file declares a message's placeholders with.
// ICU's own keywords, as "plural" and "other", are not among them: the file
// says those in the message.
func (c catalog) placeholders(key string) []string {
	meta, ok := c["@"+key].(map[string]any)
	if !ok {
		return nil
	}
	params, ok := meta["placeholders"].(map[string]any)
	if !ok {
		return nil
	}
	var out []string
	for name := range params {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// used is the placeholders a message's text names, as "{count} found" and
// "{count, plural, one {...} other {...}}".
var used = regexp.MustCompile(`\{([a-z][A-Za-z0-9_]*)[,}]`)

// TestKeyCoverage covers the catalogs: a translation says every message the
// source does, says it with something, and fills the same placeholders, so that
// a message added to the source cannot go untranslated by silence.
func TestKeyCoverage(t *testing.T) {
	want := read(t, source).messages()
	if len(want) == 0 {
		t.Fatal("en.json holds no messages")
	}
	for _, lang := range translated {
		c := read(t, lang)
		have := map[string]bool{}
		for _, key := range c.messages() {
			have[key] = true
			if s, _ := c[key].(string); strings.TrimSpace(s) == "" {
				t.Errorf("%s.json: %q is empty", lang, key)
			}
		}
		for _, key := range want {
			if !have[key] {
				t.Errorf("%s.json: no message for %q", lang, key)
			}
		}
		for _, key := range c.messages() {
			if !contains(want, key) {
				t.Errorf("%s.json: %q is not a message of %s.json", lang, key, source)
			}
		}
	}
}

// TestPlaceholders covers the placeholders: a message that names them declares
// them, and a translation declares the same ones as the source.
func TestPlaceholders(t *testing.T) {
	src := read(t, source)
	for _, key := range src.messages() {
		declared := src.placeholders(key)
		var named []string
		if s, _ := src[key].(string); s != "" {
			for _, m := range used.FindAllStringSubmatch(s, -1) {
				if !contains(named, m[1]) {
					named = append(named, m[1])
				}
			}
		}
		sort.Strings(named)
		if strings.Join(declared, ",") != strings.Join(named, ",") {
			t.Errorf("%q declares %v, its message names %v", key, declared, named)
		}
	}
	for _, lang := range translated {
		c := read(t, lang)
		for _, key := range src.messages() {
			if got, want := c.placeholders(key), src.placeholders(key); strings.Join(got, ",") != strings.Join(want, ",") {
				t.Errorf("%s.json: %q declares %v, the source %v", lang, key, got, want)
			}
		}
	}
}

// TestActionKeys covers the keys the actions name, which no typed accessor
// checks: a title or a group the catalogs lack would show as its key.
func TestActionKeys(t *testing.T) {
	c := read(t, source)
	for _, a := range action.All() {
		for _, key := range []string{a.TitleKey, a.GroupKey} {
			if _, ok := c[key].(string); !ok {
				t.Errorf("%q names %q, which %s.json does not hold", a.ID, key, source)
			}
		}
	}
	for n := 1; n <= action.MaxNumber; n++ {
		for _, id := range []string{action.GotoAction(n), action.ProfileAction(n)} {
			key := actionIDKey(t, id)
			if _, ok := c[key].(string); !ok {
				t.Errorf("%q names %q, which %s.json does not hold", id, key, source)
			}
		}
	}
}

// actionIDKey is the catalog key of an action by its id, as the catalog lists
// it: the numbered ones are in it, one entry each.
func actionIDKey(t *testing.T, id string) string {
	t.Helper()
	a, ok := action.Lookup(id)
	if !ok {
		t.Fatalf("%q is not in the catalog", id)
	}
	return a.TitleKey
}

// TestLanguages covers what the window offers the languages by: the source
// language first, which it falls back to, a name of its own for each, and a name
// of its own to show it in the language list.
func TestLanguages(t *testing.T) {
	langs := locales.Catalog.Languages()
	if len(langs) < 2 || langs[0] != source {
		t.Fatalf("languages = %v, want %s first", langs, source)
	}
	for _, lang := range langs {
		if name := locales.In(lang).Language(); name == "" {
			t.Errorf("no name for %q in %q", lang, lang)
		}
		if name := i18n.Autonym(lang); name == "" {
			t.Errorf("no name of its own for %q", lang)
		}
	}
}

// contains tells whether list holds s.
func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
