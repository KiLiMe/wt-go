package sys

import (
	"os"
	"testing"
)

// TestExpandPath covers the placeholders a path in the settings may write:
// the home directory, and the variables of the machine.
func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("no home directory: %v", err)
	}
	t.Setenv("WTGO_TEST_VAR", "value")
	t.Setenv("WTGO_TEST_EMPTY", "")
	tests := []struct {
		path string
		want string
	}{
		{"", ""},
		{"plain/path", "plain/path"},
		{"/abs/path", "/abs/path"},
		{`C:\Users\me`, `C:\Users\me`},
		{"~", home},
		{"~/src", home + "/src"},
		{`~\src`, home + `\src`},
		{"~other", "~other"},
		{"%WTGO_TEST_VAR%", "value"},
		{"$WTGO_TEST_VAR", "value"},
		{"${WTGO_TEST_VAR}", "value"},
		{"/a/%WTGO_TEST_VAR%/b", "/a/value/b"},
		{"/a/$WTGO_TEST_VAR/b", "/a/value/b"},
		{"a${WTGO_TEST_VAR}b", "avalueb"},
		{"a$WTGO_TEST_EMPTY/b", "a/b"},
		{"%WTGO_TEST_UNSET%", "%WTGO_TEST_UNSET%"},
		{"$WTGO_TEST_UNSET", "$WTGO_TEST_UNSET"},
		{"${WTGO_TEST_UNSET}", "${WTGO_TEST_UNSET}"},
		{"a/%WTGO_TEST_UNSET%/b", "a/%WTGO_TEST_UNSET%/b"},
		{"100%", "100%"},
		{"%%", "%%"},
		{"$", "$"},
		{"${}", "${}"},
	}
	for _, tt := range tests {
		if got := ExpandPath(tt.path); got != tt.want {
			t.Errorf("ExpandPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

// TestExpandPathHome covers a path that starts with the home placeholder and
// a variable after it.
func TestExpandPathHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("no home directory: %v", err)
	}
	t.Setenv("WTGO_TEST_VAR", "value")
	if got, want := ExpandPath("~/$WTGO_TEST_VAR"), home+"/value"; got != want {
		t.Errorf("ExpandPath(%q) = %q, want %q", "~/$WTGO_TEST_VAR", got, want)
	}
}
