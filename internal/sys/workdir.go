package sys

import "os"

// WorkDir is the directory this process was started in. A desktop or a file
// manager hands a terminal its directory by starting it there — chdir then
// exec is the only channel an exec-style launcher has — so that cwd is the
// caller saying "open a terminal here". mygo's terminal replaces an empty
// Options.Dir with the home directory, which would drop that, so the app has
// to read it before handing the options over.
func WorkDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}
