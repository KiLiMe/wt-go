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
	r := Run(args, strings.NewReader(""), &o, &e)
	return o.String(), e.String(), r.Handled, r.Code
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

// A report of a Windows machine may spell the account in another case than the
// environment holds it, and it still names that machine: the match folds case
// where the filesystem does, and the text around it keeps its own.
func TestRedactIgnoresCase(t *testing.T) {
	withFold(t, true)

	tests := []struct {
		name  string
		pairs [][2]string
		in    string
		want  string
	}{
		{
			name:  "the account in another case",
			pairs: [][2]string{{`C:\Users\Foo`, `%USERPROFILE%`}},
			in:    `c:\users\FOO\settings.json`,
			want:  `%USERPROFILE%\settings.json`,
		},
		{
			name:  "a mixed case that matches nothing case-sensitively",
			pairs: [][2]string{{`C:\Users\Foo\AppData\Local`, `%LOCALAPPDATA%`}, {`C:\Users\Foo`, `%USERPROFILE%`}},
			in:    `c:\USERS\foo\appdata\LOCAL\wt-go\wt-go.exe`,
			want:  `%LOCALAPPDATA%\wt-go\wt-go.exe`,
		},
		{
			name:  "the longer path still wins over the home it sits under",
			pairs: [][2]string{{`C:\Users\Foo\AppData\Local`, `%LOCALAPPDATA%`}, {`C:\Users\Foo`, `%USERPROFILE%`}},
			in:    `C:\users\foo\APPDATA\local\wt-go\wt-go.log`,
			want:  `%LOCALAPPDATA%\wt-go\wt-go.log`,
		},
		{
			name:  "every occurrence, not only the first",
			pairs: [][2]string{{`C:\Users\Foo`, `%USERPROFILE%`}},
			in:    `C:\Users\Foo\a; c:\users\foo\b; C:\USERS\FOO\c`,
			want:  `%USERPROFILE%\a; %USERPROFILE%\b; %USERPROFILE%\c`,
		},
		{
			name:  "a run of matches, one after the other",
			pairs: [][2]string{{`C:\Users\Foo`, `~`}},
			in:    `c:\users\fooC:\USERS\FOO`,
			want:  `~~`,
		},
		{
			name:  "the text around the path keeps its case",
			pairs: [][2]string{{`C:\Users\Foo`, `%USERPROFILE%`}},
			in:    `E102 Shell.Failed at C:\USERS\foo: not found`,
			want:  `E102 Shell.Failed at %USERPROFILE%: not found`,
		},
		{
			name:  "a path that is not there is left alone",
			pairs: [][2]string{{`C:\Users\Foo`, `%USERPROFILE%`}},
			in:    `C:\Users\someone-else\x`,
			want:  `C:\Users\someone-else\x`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := redact(tt.in, tt.pairs); got != tt.want {
				t.Errorf("redact(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// The whole Redact with the case mixed: no field of the report still names the
// account.
func TestRedactMixedCaseReport(t *testing.T) {
	withFold(t, true)

	pairs := [][2]string{
		{`C:\Users\Foo\AppData\Local`, `%LOCALAPPDATA%`},
		{`C:\Users\Foo`, `%USERPROFILE%`},
	}
	r := Report{
		Executable: `c:\users\foo\APPData\local\wt-go\wt-go.exe`,
		Settings:   SettingsReport{Path: `C:\Users\FOO\settings.json`, Error: `bad field at C:\Users\foo\settings.json`},
		Native:     NativeReport{Path: `c:\USERS\Foo\AppData\Local\mygo\natives\x\ghostty-vt.dll`},
		Profiles:   []ProfileReport{{ID: "pwsh", Command: `C:\USERS\foo\pwsh.exe`}},
		Crashes:    []CrashReport{{Path: `c:\users\fo\crash-1.log`, Fault: `E101 at C:\Users\FOO\x`}},
		Log:        LogReport{Path: `C:\Users\foo\AppData\local\wt-go\wt-go.log`, Fault: `E102 at C:\Users\Foo\x`},
		Checks:     []Check{{Code: "E102", Detail: `no library under c:\users\fOo`}},
	}
	got := r.Redact(pairs)
	want := map[string]string{
		"exe":         `%LOCALAPPDATA%\wt-go\wt-go.exe`,
		"settings":    `%USERPROFILE%\settings.json`,
		"settingsErr": `bad field at %USERPROFILE%\settings.json`,
		"native":      `%LOCALAPPDATA%\mygo\natives\x\ghostty-vt.dll`,
		"command":     `%USERPROFILE%\pwsh.exe`,
		"fault":       `E101 at %USERPROFILE%\x`,
		"log":         `%LOCALAPPDATA%\wt-go\wt-go.log`,
		"logFault":    `E102 at %USERPROFILE%\x`,
		"detail":      `no library under %USERPROFILE%`,
	}
	gotFields := map[string]string{
		"exe":         got.Executable,
		"settings":    got.Settings.Path,
		"settingsErr": got.Settings.Error,
		"native":      got.Native.Path,
		"command":     got.Profiles[0].Command,
		"fault":       got.Crashes[0].Fault,
		"log":         got.Log.Path,
		"logFault":    got.Log.Fault,
		"detail":      got.Checks[0].Detail,
	}
	for what, field := range gotFields {
		if field != want[what] {
			t.Errorf("%s = %q, want %q", what, field, want[what])
		}
	}
	// A path outside the account is left as it was, and the crash is still one.
	if got.Crashes[0].Path != `c:\users\fo\crash-1.log` {
		t.Errorf("crash = %q, want the path outside the account", got.Crashes[0].Path)
	}
}

// Where the filesystem keeps case, so does the match: two names that differ in
// case are two files there, and rewriting one for the other would turn a real
// path into a false one.
func TestRedactCaseSensitive(t *testing.T) {
	withFold(t, false)

	pairs := [][2]string{{`C:\Users\Foo`, `%USERPROFILE%`}}
	if got, want := redact(`c:\users\foo\x`, pairs), `c:\users\foo\x`; got != want {
		t.Errorf("redact(mixed case) = %q, want %q", got, want)
	}
	if got, want := redact(`C:\Users\Foo\x`, pairs), `%USERPROFILE%\x`; got != want {
		t.Errorf("redact(path) = %q, want %q", got, want)
	}
}

// withFold sets whether the match folds case for one test, and puts it back.
func withFold(t *testing.T, fold bool) {
	t.Helper()
	was := foldCase
	foldCase = fold
	t.Cleanup(func() { foldCase = was })
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

func crashCheck(t *testing.T, r Report) Check {
	t.Helper()
	for _, c := range checks(r) {
		if c.Code == fault.Crashed {
			return c
		}
	}
	t.Fatalf("checks() = %+v, want %s", checks(r), fault.Crashed)
	return Check{}
}

// A Go panic leaves no fault code in the log: the crash is reported by its
// path, and the check still fails.
func TestChecksCrashWithoutFault(t *testing.T) {
	path := `C:\Users\me\AppData\Local\wt-go\crash-1.log`
	r := Report{Crashes: []CrashReport{{Path: path}}}
	got := crashCheck(t, r)
	if got.OK {
		t.Errorf("%s = %+v, want a failed check", fault.Crashed, got)
	}
	if got.Detail != path {
		t.Errorf("detail = %q, want %q", got.Detail, path)
	}
}

// Without a crash log there is nothing for the check to answer, so it is not
// asked at all: a passing E302 line would claim the disk was looked at.
func TestChecksNoCrash(t *testing.T) {
	for _, c := range checks(Report{}) {
		if c.Code == fault.Crashed {
			t.Fatalf("%s = %+v, want no crash check without a crash log", fault.Crashed, c)
		}
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
