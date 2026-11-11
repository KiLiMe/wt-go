//go:build windows

package store

// DefaultFontFamily is the family a terminal draws with until the settings
// name one: the one Windows Terminal installs, with the generic family after
// it, so that a machine without it draws with a monospace font rather than a
// proportional one.
func DefaultFontFamily() string { return "Cascadia Mono,monospace" }

// DefaultShellHint is the command line the settings show before a profile
// names one.
func DefaultShellHint() string { return "pwsh.exe" }

// DefaultHomeHint is how the settings write the user's home directory.
func DefaultHomeHint() string { return "%USERPROFILE%" }
