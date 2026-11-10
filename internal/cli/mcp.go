package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"wtgo/internal/mcp"
	"wtgo/internal/sys"
)

// servemcp runs the MCP server on stdin/stdout until the stream ends. It
// touches no window: an editor or an agent asks about this machine while
// wt-go runs, or while it does not.
func servemcp(in io.Reader, out, errOut io.Writer) int {
	s := Server()
	s.In = in
	s.Out = out
	s.Logf = func(format string, a ...any) { fmt.Fprintf(errOut, "wtgo mcp: "+format+"\n", a...) }
	if err := s.Serve(); err != nil {
		fmt.Fprintf(errOut, "wtgo mcp: %v\n", err)
		return 1
	}
	return 0
}

// Server is the MCP server wt-go offers: the same report `wtgo diag` prints,
// as tools and resources, so an agent reads the checks instead of asking the
// user to copy a window.
func Server() *mcp.Server {
	return &mcp.Server{
		Name:    "wt-go",
		Version: Collect().Version,
		Instructions: `wt-go is a terminal multiplexer for Windows in one window: tabs, split
panes, profiles, and a settings.json under %LOCALAPPDATA%\wt-go.

Call wtgo_diag first: it answers which checks fail on this machine, each with
a code (E1xx before the window is up, E2xx reading settings.json, E3xx while
it runs). Every code is written up in docs/故障排查.md of the wt-go
repository: what it means, how to confirm it, and what to change.`,
		Tools: []mcp.Tool{
			{
				Name:  "wtgo_diag",
				Title: "Diagnose wt-go",
				Description: `The checks wt-go fails on this machine, as text: the build, the paths,
the settings, the shells, and a crash log when there is one. Each failure
comes with its code and what the code means. Start here.`,
				Schema: mcp.ObjectSchema(nil, nil),
				Run: func(map[string]any) (string, error) {
					return Collect().Explain(), nil
				},
			},
			{
				Name:        "wtgo_status",
				Title:       "wt-go status",
				Description: "The same report as wtgo_diag, as JSON: build, settings, native library, shells, profiles, crash logs and the checks.",
				Schema:      mcp.ObjectSchema(nil, nil),
				Run: func(map[string]any) (string, error) {
					return Collect().JSON()
				},
			},
			{
				Name:  "wtgo_settings",
				Title: "wt-go settings",
				Description: `Where settings.json is, whether it parses, and the values a report usually turns
on: language, theme, colour scheme, default profile, keybinding count. The
error is what the window would say instead of using the file.`,
				Schema: mcp.ObjectSchema(nil, nil),
				Run: func(map[string]any) (string, error) {
					r := Collect()
					return mcp.JSON(map[string]any{
						"settings": r.Settings,
						"checks":   r.Checks[:2],
					})
				},
			},
			{
				Name:  "wtgo_profiles",
				Title: "wt-go profiles",
				Description: `The shells this machine has, and the profiles after the configured and the found
ones were merged, with whether each one can actually start: a profile whose
command is missing opens a tab with nothing in it.`,
				Schema: mcp.ObjectSchema(nil, nil),
				Run: func(map[string]any) (string, error) {
					r := Collect()
					return mcp.JSON(map[string]any{
						"shells":   r.Shells,
						"profiles": r.Profiles,
					})
				},
			},
			{
				Name:  "wtgo_log",
				Title: "Recent crash log",
				Description: `The end of a log wt-go wrote when it stopped, newest crash first. Use it after a
window closed by itself: the fault line names the code the doc explains.`,
				Schema: mcp.ObjectSchema(map[string]any{
					"lines": mcp.Prop("number", "how many lines from the end, 1 to 2000 (default 120)"),
					"path":  mcp.Prop("string", "a log path from wtgo_status; the newest crash log when empty"),
				}, nil),
				Run: func(args map[string]any) (string, error) {
					lines := mcp.ArgInt(args, "lines", 120, 1, 2000)
					path := mcp.ArgString(args, "path", "")
					if path == "" {
						paths, err := sys.CrashLogs()
						if err != nil {
							return "", err
						}
						if len(paths) == 0 {
							return "", errors.New("no crash log: wt-go has not stopped with a panic")
						}
						path = paths[0]
					}
					text, cut, err := sys.Tail(path, 1<<20)
					if err != nil {
						return "", err
					}
					kept := tailLines(text, lines)
					if cut {
						kept = "(only the end of the file)\n" + kept
					}
					return path + "\n\n" + kept, nil
				},
			},
		},
		Resources: []mcp.Resource{
			{
				URI:         "wtgo://diag",
				Name:        "diag",
				Description: "The wt-go diagnostic report as text, the way wtgo diag prints it.",
				MimeType:    "text/plain",
				Read:        func() (string, error) { return Collect().Explain(), nil },
			},
			{
				URI:         "wtgo://status",
				Name:        "status",
				Description: "The wt-go status report as JSON.",
				MimeType:    "application/json",
				Read:        func() (string, error) { return Collect().JSON() },
			},
		},
	}
}

// tailLines keeps the last n lines of text, so that a long log does not fill
// an answer an agent has to read.
func tailLines(text string, n int) string {
	text = strings.TrimRight(text, "\n")
	lines := strings.Split(text, "\n")
	if len(lines) <= n {
		return text
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}
