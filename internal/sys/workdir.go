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

// Home is the user's home directory, or "" when the system names none.
func Home() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// StartDir is the directory a tab starts in, for what the settings call a
// directory: "home" is the user's, "current" is where this process was
// started, and anything else is a path, handed back as it is for ExpandPath to
// read. Nothing at all is the platform's own answer, which is not the same
// question everywhere (see defaultStartDir). The words are answered here
// because they are facts about this OS, while the settings stay a file of text.
func StartDir(pref string) string {
	switch pref {
	case "":
		return defaultStartDir()
	case "current":
		return WorkDir()
	case "home":
		return Home()
	}
	return pref
}
