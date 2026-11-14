package cli

import (
	"runtime"
	"strings"
	"wtgo/internal/fault"
)

// checks are the questions a diagnostic asks, in the order the window meets
// them.
func checks(r Report) []Check {
	out := make([]Check, 0, 6)
	add := func(c fault.Code, ok bool, detail string) {
		out = append(out, Check{Code: c, OK: ok, Detail: detail})
	}
	dirOK := r.Settings.Path != ""
	add(fault.NoConfigDir, dirOK, r.Settings.Path)
	switch {
	case r.Settings.Error == "":
		add(fault.BadSettings, true, "settings.json parses"+keybindings(r.Settings.Keybindings))
	case strings.Contains(r.Settings.Error, " must be "):
		add(fault.BadField, false, r.Settings.Error)
	default:
		add(fault.BadSettings, false, r.Settings.Error)
	}
	add(fault.NativeLib, r.Native.Found, or(r.Native.Path, "no libghostty-vt in $MYGO_GHOSTTY_VT, next to the exe, or under the mygo natives cache"))
	add(fault.NoShell, len(r.Profiles) > 0, count(len(r.Profiles), "profile"))
	// One profile that cannot start is a tab that opens empty, which the user
	// asked for: the check fails and names it.
	why := missing(r.Profiles)
	if why == "" {
		why = count(len(r.Profiles), "profile") + " can start"
	}
	add(fault.ChildFailed, len(r.Profiles) > 0 && missing(r.Profiles) == "", why)
	if len(r.Crashes) > 0 {
		c := r.Crashes[0]
		// Any crash log is a crash, whether or not it carries a fault code: a
		// plain Go panic has none, so Fault is empty and only the path remains.
		add(fault.Crashed, false, or(c.Fault, c.Path))
	}
	return out
}

// Redact is the report with the machine's own paths replaced by the
// placeholders that name them, so it can be pasted somewhere public: the
// report still says which files matter without saying whose machine it is.
func (r Report) Redact(pairs [][2]string) Report {
	if len(pairs) == 0 {
		return r
	}
	s := func(v string) string { return redact(v, pairs) }
	r.Executable = s(r.Executable)
	r.Settings.Path = s(r.Settings.Path)
	r.Settings.Error = s(r.Settings.Error)
	r.Native.Path = s(r.Native.Path)
	profiles := make([]ProfileReport, len(r.Profiles))
	for i, p := range r.Profiles {
		p.Command = s(p.Command)
		profiles[i] = p
	}
	r.Profiles = profiles
	crashes := make([]CrashReport, len(r.Crashes))
	for i, c := range r.Crashes {
		c.Path, c.Fault = s(c.Path), s(c.Fault)
		crashes[i] = c
	}
	r.Crashes = crashes
	r.Log.Path = s(r.Log.Path)
	r.Log.Fault = s(r.Log.Fault)
	checks := make([]Check, len(r.Checks))
	for i, c := range r.Checks {
		c.Detail = s(c.Detail)
		checks[i] = c
	}
	r.Checks = checks
	return r
}

// redact replaces each path with its placeholder, longest first so a nested
// directory wins over the home it sits under.
func redact(s string, pairs [][2]string) string {
	for _, p := range pairs {
		s = replacePath(s, p[0], p[1])
	}
	return s
}

// foldCase is whether two paths that differ only in case name the same file.
// Windows and macOS say they do; Linux says they do not.
var foldCase = runtime.GOOS == "windows" || runtime.GOOS == "darwin"

// replacePath is ReplaceAll for a path, matching case the way the filesystem
// the report came from does: on Windows the account may sit in the report as
// C:\Users\Foo while %USERPROFILE% holds c:\users\foo, and a report of that
// machine has to be redacted all the same. The text keeps its own case outside
// the match; only the path itself becomes the placeholder.
//
// On a case-sensitive filesystem the match is a plain ReplaceAll, so a string
// that is not a path is never rewritten: two names that differ in case are two
// files there.
func replacePath(s, from, to string) string {
	if from == "" || !foldCase {
		return strings.ReplaceAll(s, from, to)
	}
	// The needle folds to lower case rather than the haystack: EqualFold is not
	// the same test as a lowered byte compare for the handful of characters
	// Unicode folds differently, and a path is bytes the system handed us.
	needle := strings.ToLower(from)
	var b strings.Builder
	for {
		i := strings.Index(strings.ToLower(s), needle)
		if i < 0 {
			b.WriteString(s)
			return b.String()
		}
		b.WriteString(s[:i])
		b.WriteString(to)
		s = s[i+len(needle):]
	}
}

// Failed is the checks that did not pass.
func (r Report) Failed() []Check {
	out := make([]Check, 0, len(r.Checks))
	for _, c := range r.Checks {
		if !c.OK {
			out = append(out, c)
		}
	}
	return out
}
