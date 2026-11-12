//go:build windows

package sys

// defaultStartDir is where a tab starts when neither the tab nor the settings
// name a directory. Windows gives a started exe no directory worth inheriting:
// Explorer hands it the exe's own folder and the Start menu %SystemRoot%\System32,
// so "where I was started" describes the launcher, not the person. The home
// directory is what Windows Terminal's own default says, and -d is how a launch
// that does mean a directory says so.
func defaultStartDir() string {
	if home := Home(); home != "" {
		return home
	}
	// No home directory to give: where we started is the next best answer.
	return WorkDir()
}
