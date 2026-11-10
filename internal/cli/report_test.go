package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
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

func TestDiagText(t *testing.T) {
	out, _, handled, code := run(t, "diag")
	if !handled || code != 0 {
		t.Fatalf("Run(diag) = %v, %d", handled, code)
	}
	for _, want := range []string{"wt-go ", "settings:", "native:", "checks", "profile:"} {
		if !strings.Contains(out, want) {
			t.Errorf("diag = %q, want %q", out, want)
		}
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
