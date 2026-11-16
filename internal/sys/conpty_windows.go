//go:build windows

package sys

import (
	"os"
	"path/filepath"
)

// conptyDirName is the subdirectory of the executable's folder the bundled
// ConPTY is unpacked into (ADR-83): those files are Microsoft's, not the
// app's, and keeping them together is what makes "this is not ours" visible.
// conpty.dll runs the OpenConsole.exe beside it, so both live here.
const conptyDirName = "conpty"

// bundledConPTYDir is the ConPTY directory beside the executable, "" when
// there is none: a copy unzipped from the archive without it has nothing to
// load, and Windows' own console is used.
func bundledConPTYDir() string {
	dir, err := executableDir()
	if err != nil {
		return ""
	}
	dir = filepath.Join(dir, conptyDirName)
	if !hasConPTY(dir) {
		return ""
	}
	return dir
}

// UseBundledConPTY points the terminal plugin at the ConPTY unpacked beside
// the executable, so that the terminals are made in the console the report
// names. It must run before the first terminal: the plugin reads MYGO_CONPTY
// once, when it makes the first one.
//
// A person who set MYGO_CONPTY decides, including the literal "system": that
// variable is how a machine that wants Windows' own console says so, and
// setting it here would take the choice away.
func UseBundledConPTY() {
	if os.Getenv("MYGO_CONPTY") != "" {
		return
	}
	if dir := bundledConPTYDir(); dir != "" {
		os.Setenv("MYGO_CONPTY", dir)
		adopted = dir
	}
}

// adopted is the directory UseBundledConPTY put in MYGO_CONPTY: the variable
// then names this program's own choice rather than a person's, and a report of
// which ConPTY runs must not say someone pointed it there.
var adopted string
