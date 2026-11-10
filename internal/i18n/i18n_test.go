package i18n

import "testing"

// TestSet covers the two languages: Chinese maps the window's words, and
// English, or a language nobody has written yet, is the source itself.
func TestSet(t *testing.T) {
	Set("zh-CN")
	if got := T("Close Tab"); got != "关闭标签" {
		t.Errorf("zh-CN Close Tab = %q", got)
	}
	if got := T("not a word"); got != "not a word" {
		t.Errorf("missing word = %q, want the source", got)
	}
	Set("en-US")
	if got := T("Close Tab"); got != "Close Tab" {
		t.Errorf("en-US Close Tab = %q", got)
	}
	Set("fr-FR")
	if got := T("Close Tab"); got != "Close Tab" {
		t.Errorf("fr-FR Close Tab = %q", got)
	}
}
