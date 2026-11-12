package sys

// defaultStartDir is where a tab starts when neither the tab nor the settings
// name a directory: the user's home. Every platform can hand a terminal the
// directory it was launched in, but which one arrives is the launcher's habit
// rather than the person's wish — Explorer passes the exe's own folder, the
// Start menu %SystemRoot%\System32, a desktop session whatever it happened to
// be started in — so the answer is the same everywhere instead of a guess
// per platform. Home is also what Windows Terminal's own default says. A
// launch that does mean a directory says so with -d, or by writing "current".
func defaultStartDir() string {
	if home := Home(); home != "" {
		return home
	}
	// No home directory to give: where we started is the next best answer.
	return WorkDir()
}
