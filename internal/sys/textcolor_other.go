//go:build !linux

package sys

// FixTextColor has nothing to fix: the other systems rasterize glyphs
// through their own APIs, which hand MyGo the coverage it asked for.
func FixTextColor() bool { return true }
