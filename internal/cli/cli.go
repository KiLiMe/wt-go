package cli

import (
	"fmt"
	"io"

	"wtgo/internal/sys"
)

// Usage is what `wt-go help` prints: the window, and what runs without one.
const Usage = `wt-go — a terminal multiplexer in one window

  wt-go [-d DIR]                   open the window, new tabs starting in DIR
  wt-go diag [--json] [--redact]   what this machine is missing, and why
  wt-go mcp                        answer the Model Context Protocol on stdin/stdout
  wt-go version                    print the build
  wt-go help                       this text

DIR is a path, or current / home as in settings.json; a profile's own dir
still wins. The report and the MCP tools read the same files the window reads,
and every failure they name has a code with the checks and the fix in
docs/故障排查.md.
`

// Result is what the command line asked for. Handled is true when a command
// answered and the window must not open; Code is what it exits with. StartDir
// is the directory -d named for the window's new tabs, "" when it named none.
type Result struct {
	Handled  bool
	Code     int
	StartDir string
}

// Run answers the command line: the commands that do not open a window, and
// the flag that shapes it.
func Run(args []string, in io.Reader, out, errOut io.Writer) Result {
	startDir, rest, bad := windowFlags(args)
	if bad != "" {
		fmt.Fprintf(errOut, "wt-go: %s\n\n%s", bad, Usage)
		return Result{Handled: true, Code: 2}
	}
	if len(rest) == 0 {
		return Result{StartDir: startDir}
	}
	// -d is about the window, and a command answers instead of opening one.
	if startDir != "" {
		fmt.Fprintf(errOut, "wt-go: -d opens the window; it cannot go with %q\n\n%s", rest[0], Usage)
		return Result{Handled: true, Code: 2}
	}
	switch rest[0] {
	case "diag":
		return Result{Handled: true, Code: diag(rest[1:], out, errOut)}
	case "mcp":
		return Result{Handled: true, Code: servemcp(in, out, errOut)}
	case "version":
		r := Collect()
		fmt.Fprintf(out, "wt-go %s %s\n", r.version(), r.Revision)
		return Result{Handled: true}
	case "help", "-h", "--help", "-help":
		fmt.Fprint(out, Usage)
		return Result{Handled: true}
	}
	fmt.Fprintf(errOut, "wt-go: unknown command %q\n\n%s", rest[0], Usage)
	return Result{Handled: true, Code: 2}
}

// windowFlags takes the flags that shape the window off the front of the
// command line: -d DIR, or --dir DIR, is the directory the tabs it opens start
// in. What is left is the command, if any; bad names the first thing it could
// not read.
func windowFlags(args []string) (startDir string, rest []string, bad string) {
	for len(args) > 0 && (args[0] == "-d" || args[0] == "--dir") {
		if len(args) < 2 || args[1] == "" {
			return "", nil, args[0] + " needs a directory"
		}
		startDir, args = args[1], args[2:]
	}
	return startDir, args, ""
}

// diag prints the report: the lines a person reads, or with --json the same
// report as a tool reads it. --redact replaces the account's own paths with
// placeholders, which is what a report pasted into an issue wants.
func diag(args []string, out, errOut io.Writer) int {
	asJSON, asRedacted := false, false
	for _, a := range args {
		switch a {
		case "--json", "-json":
			asJSON = true
		case "--redact", "-redact":
			asRedacted = true
		default:
			fmt.Fprintf(errOut, "wt-go diag: unknown flag %q\n", a)
			return 2
		}
	}
	r := Collect()
	if asRedacted {
		r = r.Redact(sys.Redactions())
	}
	if asJSON {
		text, err := r.JSON()
		if err != nil {
			fmt.Fprintf(errOut, "wt-go diag: %v\n", err)
			return 1
		}
		fmt.Fprintln(out, text)
		return 0
	}
	// The failing checks come last, each with the line its code stands for.
	fmt.Fprint(out, r.Explain())
	return 0
}
