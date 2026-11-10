package i18n

import (
	"os"
	"regexp"
	"testing"
)

// TestWords checks that every word the code asks for has a translation: the
// window in Chinese falls back to English for the ones it has not.
func TestWords(t *testing.T) {
	re := regexp.MustCompile(`i18n\.T\("([^"]*)"\)`)
	dirs := []string{"../app", "../ui"}
	missing := map[string]bool{}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if e.IsDir() || len(e.Name()) < 4 || e.Name()[len(e.Name())-3:] != ".go" {
				continue
			}
			b, err := os.ReadFile(dir + "/" + e.Name())
			if err != nil {
				t.Fatal(err)
			}
			for _, m := range re.FindAllStringSubmatch(string(b), -1) {
				if _, ok := zhCN[m[1]]; !ok {
					missing[m[1]] = true
				}
			}
		}
	}
	for word := range missing {
		t.Errorf("no Chinese for %q", word)
	}
}
