package cli

import (
	"fmt"
	"io"
	"strings"
	"wtgo/internal/fault"
)

// Boundary: the report as a person reads it -- the lines of the text form and
// the one block an agent gets. The checks that feed it are report_checks.go.

// Text writes the report as lines a person reads.
func (r Report) Text(w io.Writer) {
	line := func(format string, a ...any) { fmt.Fprintf(w, format+"\n", a...) }
	line("wt-go %s %s %s", r.version(), r.Revision, r.Platform)
	line("exe:      %s", or(r.Executable, "unknown"))
	if r.Portable {
		line("portable: %s", or(r.Data, "unknown"))
	}
	line("settings: %s", or(r.Settings.Path, "not available"))
	if r.Settings.Error != "" {
		line("          %s", r.Settings.Error)
	}
	line("native:   %s", or(r.Native.Path, "not found"))
	if strings.HasPrefix(r.Platform, "windows/") {
		line("conpty:   %s", or(r.ConPTY.Path, "Windows' own console, which drops what a program asks"))
	}
	for _, p := range r.Profiles {
		state := "ok"
		if !p.Available {
			state = "missing"
		}
		line("profile:  %-16s %-40s %s", p.ID, p.Command, state)
	}
	for _, c := range r.Crashes {
		line("crash:    %s", c.Path)
	}
	if r.Log.Path != "" {
		line("log:      %s", r.Log.Path)
	}
	if r.Log.Fault != "" {
		line("last fault: %s", r.Log.Fault)
	}
	if failed := r.Failed(); len(failed) > 0 {
		for _, c := range failed {
			line("FAILED    %s %s", c.Code, c.Detail)
		}
		return
	}
	line("checks:   all %d passed", len(r.Checks))
}

// Explain is the report as one block of text for an agent: the lines, then
// each failed check with what its code means, then where the details are
// written — once for the report, not once per line.
func (r Report) Explain() string {
	var b strings.Builder
	r.Text(&b)
	failed := r.Failed()
	for _, c := range failed {
		fmt.Fprintf(&b, "\n%s\n", fault.Line(c.Code, c.Detail))
	}
	if len(failed) > 0 {
		fmt.Fprintf(&b, "\nsee %s\n", fault.Doc())
	}
	return b.String()
}

func missing(ps []ProfileReport) string {
	var out []string
	for _, p := range ps {
		if !p.Available {
			out = append(out, p.ID+" ("+p.Command+")")
		}
	}
	if len(out) == 0 {
		return ""
	}
	return "not on this machine: " + strings.Join(out, ", ")
}

func count(n int, what string) string {
	if n == 1 {
		return "1 " + what
	}
	return fmt.Sprintf("%d %ss", n, what)
}

func keybindings(n int) string {
	if n == 0 {
		return ""
	}
	return ", " + count(n, "keybinding")
}

func or(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
