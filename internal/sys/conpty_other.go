//go:build !windows

package sys

// bundledConPTYDir is "" on a platform whose terminals are a pty of the
// system's own: there is no console to unpack.
func bundledConPTYDir() string { return "" }

// UseBundledConPTY does nothing where there is no ConPTY to point at.
func UseBundledConPTY() {}

// adopted is the directory UseBundledConPTY would have put in MYGO_CONPTY,
// which no such platform has: the variable, when a person sets it, is theirs.
var adopted string
