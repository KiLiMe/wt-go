//go:build !windows

package store

// DefaultFontFamily is the family a terminal draws with until the settings
// name one: the monospace family fontconfig resolves.
func DefaultFontFamily() string { return "monospace" }

// DefaultShellHint is the command line the settings show before a profile
// names one.
func DefaultShellHint() string { return "/bin/bash" }

// DefaultHomeHint is how the settings write the user's home directory.
func DefaultHomeHint() string { return "~" }
