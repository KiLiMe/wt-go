//go:build !windows

package store

// DefaultFontFamily is the family a terminal draws with until the settings
// name one: the monospace faces a desktop almost always installs, then the
// generic name, so that a machine with none of them still draws a grid of
// cells.
//
// The names come first because fontconfig scores a generic name by the
// language of the session: with a Han language, "monospace" resolves to the
// CJK monospace, whose Latin letters are drawn for half of a Han cell and are
// a compromise (ADR-87). A family this machine does not have is skipped
// without a word, which is what makes a list of guesses safe.
func DefaultFontFamily() string {
	return "DejaVu Sans Mono, Noto Sans Mono, Hack, Liberation Mono, monospace"
}

// DefaultShellHint is the command line the settings show before a profile
// names one: the name of the shell a POSIX system almost always has, so that
// the example is a program rather than a Windows path.
func DefaultShellHint() string { return "bash" }

// DefaultHomeHint is how the settings write the user's home directory.
func DefaultHomeHint() string { return "~" }
