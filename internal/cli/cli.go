package cli

import (
	"fmt"
	"io"
)

// Usage is what `wt-go help` prints: the window, and what runs without one.
const Usage = `wt-go — a terminal multiplexer in one window

  wt-go                     open the window
  wt-go diag [--json]       what this machine is missing, and why
  wt-go mcp                 answer the Model Context Protocol on stdin/stdout
  wt-go version             print the build
  wt-go help                this text

The report and the MCP tools read the same files the window reads, and every
failure they name has a code with the checks and the fix in docs/故障排查.md.
`

// Run answers the commands that do not open a window. handled is false when
// the arguments name nothing here, and the window opens as usual.
func Run(args []string, in io.Reader, out, errOut io.Writer) (handled bool, code int) {
	if len(args) == 0 {
		return false, 0
	}
	switch args[0] {
	case "diag":
		return true, diag(args[1:], out, errOut)
	case "mcp":
		return true, servemcp(in, out, errOut)
	case "version":
		r := Collect()
		fmt.Fprintf(out, "wt-go %s %s\n", r.Version, r.Revision)
		return true, 0
	case "help", "-h", "--help", "-help":
		fmt.Fprint(out, Usage)
		return true, 0
	}
	fmt.Fprintf(errOut, "wt-go: unknown command %q\n\n%s", args[0], Usage)
	return true, 2
}

// diag prints the report: the lines a person reads, or with --json the same
// report as a tool reads it.
func diag(args []string, out, errOut io.Writer) int {
	asJSON := false
	for _, a := range args {
		switch a {
		case "--json", "-json":
			asJSON = true
		default:
			fmt.Fprintf(errOut, "wt-go diag: unknown flag %q\n", a)
			return 2
		}
	}
	r := Collect()
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
