package app

import (
	"errors"
	"testing"

	"wtgo/internal/ui"
)

// TestConsoleOf covers where the about dialog says the terminals' console comes
// from: the dll beside the program, the one MYGO_CONPTY names, or the console
// the machine comes with when the program ships none.
func TestConsoleOf(t *testing.T) {
	none := errors.New("sys: no conpty.dll beside the executable")
	cases := []struct {
		name    string
		dll     string
		fromEnv bool
		err     error
		want    consoleSource
	}{
		{"beside the program", `C:\wt-go\conpty\conpty.dll`, false, nil, consoleSource{ui.TerminalApp, `C:\wt-go\conpty\conpty.dll`}},
		{"from the environment", `C:\build\conpty.dll`, true, nil, consoleSource{ui.TerminalEnv, `C:\build\conpty.dll`}},
		{"the machine's own", "", false, none, consoleSource{kind: ui.TerminalInherent}},
	}
	for _, tc := range cases {
		if got := consoleOf(tc.dll, tc.fromEnv, tc.err); got != tc.want {
			t.Errorf("%s: consoleOf(%q, %v, %v) = %+v, want %+v",
				tc.name, tc.dll, tc.fromEnv, tc.err, got, tc.want)
		}
	}
}
