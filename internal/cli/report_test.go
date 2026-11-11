package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"wtgo/internal/fault"
)

func run(t *testing.T, args ...string) (out, errOut string, handled bool, code int) {
	t.Helper()
	var o, e bytes.Buffer
	handled, code = Run(args, strings.NewReader(""), &o, &e)
	return o.String(), e.String(), handled, code
}

func TestNoArgumentsOpenTheWindow(t *testing.T) {
	if out, errOut, handled, code := run(t); handled || code != 0 || out != "" || errOut != "" {
		t.Errorf("Run() = %q, %q, %v, %d, want the window", out, errOut, handled, code)
	}
}

func TestVersion(t *testing.T) {
	out, _, handled, code := run(t, "version")
	if !handled || code != 0 || !strings.HasPrefix(out, "wt-go ") {
		t.Errorf("Run(version) = %q, %v, %d", out, handled, code)
	}
}

func TestHelp(t *testing.T) {
	out, _, handled, code := run(t, "help")
	if !handled || code != 0 || out != Usage {
		t.Errorf("Run(help) = %q, %v, %d, want the usage", out, handled, code)
	}
}

func TestUnknownCommand(t *testing.T) {
	out, errOut, handled, code := run(t, "nope")
	if !handled || code != 2 || out != "" || !strings.Contains(errOut, `unknown command "nope"`) {
		t.Errorf("Run(nope) = %q, %q, %v, %d", out, errOut, handled, code)
	}
	if !strings.Contains(errOut, Usage) {
		t.Errorf("stderr = %q, want the usage as well", errOut)
	}
}

// The command test only checks the wiring: what the report says belongs to
// Report.Text, and on a machine without libghostty-vt this one would fail a
// check and print differently.
func TestDiagText(t *testing.T) {
	out, _, handled, code := run(t, "diag")
	if !handled || code != 0 {
		t.Fatalf("Run(diag) = %v, %d", handled, code)
	}
	for _, want := range []string{"wt-go ", "settings:", "native:"} {
		if !strings.Contains(out, want) {
			t.Errorf("diag = %q, want %q", out, want)
		}
	}
}

func TestReportText(t *testing.T) {
	r := Report{
		Version:  "1.0.0",
		Platform: "linux/amd64",
		Checks:   []Check{{Code: fault.NoShell, OK: true, Detail: "1 profile"}},
	}
	var b strings.Builder
	r.Text(&b)
	if got := b.String(); !strings.Contains(got, "checks:   all 1 passed") {
		t.Errorf("Text() = %q, want the passing line", got)
	}

	// A failed check replaces that line: the report lists what is wrong and
	// stops claiming everything is fine.
	r.Checks = append(r.Checks, Check{Code: fault.NativeLib, OK: false, Detail: "no libghostty-vt"})
	b.Reset()
	r.Text(&b)
	got := b.String()
	if !strings.Contains(got, "FAILED    E102 no libghostty-vt") {
		t.Errorf("Text() = %q, want the failed check", got)
	}
	if strings.Contains(got, "checks:") {
		t.Errorf("Text() = %q, want no passing line next to a failure", got)
	}
}

func TestDiagJSON(t *testing.T) {
	out, _, _, code := run(t, "diag", "--json")
	if code != 0 {
		t.Fatalf("code = %d", code)
	}
	var r Report
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatalf("diag --json: %v\n%s", err, out)
	}
	if r.Program != "wt-go" || r.Settings.Path == "" || len(r.Checks) < 5 {
		t.Errorf("report = %+v, want the program, the file and the checks", r)
	}
	for _, c := range r.Checks {
		if !c.Code.Known() {
			t.Errorf("check %q has no code", c.Code)
		}
		if c.Detail == "" {
			t.Errorf("check %s has no detail", c.Code)
		}
	}
}

func TestDiagFlag(t *testing.T) {
	if _, errOut, _, code := run(t, "diag", "-x"); code != 2 || !strings.Contains(errOut, "unknown flag") {
		t.Errorf("Run(diag -x) = %q, %d", errOut, code)
	}
}

func TestRedact(t *testing.T) {
	pairs := [][2]string{
		{`C:\Users\me\AppData\Local`, `%LOCALAPPDATA%`},
		{`C:\Users\me`, `%USERPROFILE%`},
	}
	r := Report{
		Executable: `C:\Users\me\AppData\Local\wt-go\wt-go.exe`,
		Settings:   SettingsReport{Path: `C:\Users\me\settings.json`},
		Native:     NativeReport{Path: `C:\Users\me\AppData\Local\mygo\natives\x\ghostty-vt.dll`},
		Profiles:   []ProfileReport{{ID: "pwsh", Command: `C:\Users\me\pwsh.exe`}},
		Crashes:    []CrashReport{{Path: `C:\Users\me\AppData\Local\wt-go\crash-1.log`}},
		Log:        LogReport{Path: `C:\Users\me\AppData\Local\wt-go\wt-go.log`, Fault: `E102 at C:\Users\me\x`},
		Checks:     []Check{{Code: "E102", Detail: `no library under C:\Users\me`}},
	}
	got := r.Redact(pairs)
	want := map[string]string{
		"exe":      `%LOCALAPPDATA%\wt-go\wt-go.exe`,
		"settings": `%USERPROFILE%\settings.json`,
		"native":   `%LOCALAPPDATA%\mygo\natives\x\ghostty-vt.dll`,
		"command":  `%USERPROFILE%\pwsh.exe`,
		"crash":    `%LOCALAPPDATA%\wt-go\crash-1.log`,
		"log":      `%LOCALAPPDATA%\wt-go\wt-go.log`,
		"fault":    `E102 at %USERPROFILE%\x`,
		"detail":   `no library under %USERPROFILE%`,
	}
	for what, got := range map[string]string{
		"exe":      got.Executable,
		"settings": got.Settings.Path,
		"native":   got.Native.Path,
		"command":  got.Profiles[0].Command,
		"crash":    got.Crashes[0].Path,
		"log":      got.Log.Path,
		"fault":    got.Log.Fault,
		"detail":   got.Checks[0].Detail,
	} {
		if got != want[what] {
			t.Errorf("%s = %q, want %q", what, got, want[what])
		}
	}
	// The report it was given is not rewritten under the caller.
	if r.Executable != `C:\Users\me\AppData\Local\wt-go\wt-go.exe` {
		t.Errorf("Redact() changed the report it was given: %q", r.Executable)
	}
}

func TestDiagRedacted(t *testing.T) {
	out, errOut, _, code := run(t, "diag", "--redact")
	if code != 0 {
		t.Fatalf("Run(diag --redact) = %d, %q", code, errOut)
	}
	if !strings.Contains(out, "settings:") {
		t.Errorf("diag --redact = %q, want the report", out)
	}
}

func TestExplainNamesTheDoc(t *testing.T) {
	r := Report{Checks: []Check{{Code: "E102", Detail: "not found"}}}
	text := r.Explain()
	if !strings.Contains(text, "E102") || !strings.Contains(text, "docs/故障排查.md") {
		t.Errorf("Explain() = %q, want the code and where it is written up", text)
	}
}

func TestTailLines(t *testing.T) {
	text := "a\nb\nc\n"
	if got := tailLines(text, 2); got != "b\nc" {
		t.Errorf("tailLines() = %q", got)
	}
	if got := tailLines(text, 10); got != "a\nb\nc" {
		t.Errorf("tailLines() = %q", got)
	}
}
