// Package cli is wt-go without a window: the report `wt-go diag` prints and
// the MCP server `wt-go mcp` runs for an editor or an agent. It holds no OS
// knowledge of its own — every path and file comes from internal/sys — so one
// report answers a terminal, a JSON reader and a tool call alike.
package cli

import (
	"wtgo/internal/fault"
	"wtgo/internal/mcp"
	"wtgo/internal/sys"
)

// Boundary: the report as data -- what a diagnostic found, and the collection
// that fills it from the files on disk. Rendering it for a person is
// report_text.go; naming the build is version.go.

// Report is what wt-go knows about itself from outside: a fresh process
// reading the files the window reads.
type Report struct {
	Program string `json:"program"`
	// The five lines the program says about itself are Self's, embedded rather
	// than repeated, so that a field cannot land in one and not the other.
	Self
	Executable string `json:"executable,omitempty"`
	// Portable says this copy keeps its files beside the executable, and Data
	// is the directory they are in: the two decide what the paths below point
	// to, so a report of one kind is not read as the other's.
	Portable bool            `json:"portable,omitempty"`
	Data     string          `json:"data,omitempty"`
	Settings SettingsReport  `json:"settings"`
	Native   NativeReport    `json:"native"`
	ConPTY   ConPTYReport    `json:"conpty"`
	Shells   []ShellReport   `json:"shells"`
	Profiles []ProfileReport `json:"profiles"`
	Crashes  []CrashReport   `json:"crash_logs"`
	Log      LogReport       `json:"log"`
	Checks   []Check         `json:"checks"`
}

// SettingsReport is the file the settings live in, and what it holds.
type SettingsReport struct {
	Path    string `json:"path"`
	Exists  bool   `json:"exists"`
	Size    int64  `json:"size,omitempty"`
	ModTime string `json:"mod_time,omitempty"`
	// Error is why the window would ignore the file, empty when it parses.
	Error          string `json:"error,omitempty"`
	Language       string `json:"language,omitempty"`
	Theme          string `json:"theme,omitempty"`
	ColorScheme    string `json:"color_scheme,omitempty"`
	DefaultProfile string `json:"default_profile,omitempty"`
	Keybindings    int    `json:"keybindings,omitempty"`
}

// NativeReport is the libghostty-vt the terminal plugin would load.
type NativeReport struct {
	Path  string `json:"path,omitempty"`
	Env   bool   `json:"from_env,omitempty"`
	Found bool   `json:"found"`
}

// ConPTYReport is the conpty.dll the terminals would be made in, empty when
// Windows' own console is: the window still works, and what a program asks the
// terminal for does not come back.
type ConPTYReport struct {
	Path  string `json:"path,omitempty"`
	Env   bool   `json:"from_env,omitempty"`
	Found bool   `json:"found"`
}

// ShellReport is one shell the machine has.
type ShellReport struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Command string `json:"command"`
}

// ProfileReport is a profile after the configured and found ones were merged,
// with whether this machine can run it.
type ProfileReport struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Command   string `json:"command"`
	Source    string `json:"source"`
	Available bool   `json:"available"`
}

// CrashReport is one crash log, newest first.
type CrashReport struct {
	Path  string `json:"path"`
	Fault string `json:"fault,omitempty"`
}

// LogReport is what the window said to its log, and the last fault in it: a
// keybinding that did not parse or a shell that would not start is written
// there once and read back here.
type LogReport struct {
	Path  string `json:"path,omitempty"`
	Fault string `json:"fault,omitempty"`
}

// Check is one thing a diagnostic asks, with the code that answers for it
// when it fails.
type Check struct {
	Code   fault.Code `json:"code"`
	OK     bool       `json:"ok"`
	Detail string     `json:"detail"`
}

// Collect is wt-go as the files on disk say it is.
func Collect() Report {
	r := Report{Program: "wt-go", Self: About()}
	if exe, err := sys.Executable(); err == nil {
		r.Executable = exe
	}
	if dir := sys.DataDir(); dir != "" {
		r.Portable = true
		r.Data = dir
	}
	r.Settings = settingsReport()
	r.Native = nativeReport()
	r.ConPTY = conptyReport()
	r.Shells = shells()
	r.Profiles = profiles(r.Shells, r.Settings.Error == "")
	r.Crashes = crashes()
	r.Log = logReport()
	r.Checks = checks(r)
	return r
}

// JSON is the report as a tool or a script reads it.
func (r Report) JSON() (string, error) { return mcp.JSON(r) }

// fileChecks are what the diagnostic asks about settings.json: where it is, and
// whether it parses. They are picked by code rather than by place, so that the
// order the checks are asked in stays the report's alone.
func (r Report) fileChecks() []Check {
	out := make([]Check, 0, 2)
	for _, c := range r.Checks {
		switch c.Code {
		case fault.NoConfigDir, fault.BadSettings, fault.BadField:
			out = append(out, c)
		}
	}
	return out
}
