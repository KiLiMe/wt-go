//go:build !windows

package sys

// defaultStartDir is where a tab starts when neither the tab nor the settings
// name a directory: where this process was started. Inheriting it is how the
// desktops here say "open here" (a file manager chdirs before exec, Linux's
// TerminalDirCommand hands over a directory), so there is nothing to replace.
func defaultStartDir() string {
	return WorkDir()
}
