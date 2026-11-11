//go:build linux

package sys

import (
	"unsafe"

	"github.com/ebitengine/purego"
)

// textColorRules takes the subpixel order away from every font fontconfig
// matches, as patterns and as fonts.
//
// MyGo rasterizes a glyph in grayscale first and takes any glyph whose
// channels differ for a color glyph, which it draws in its own color. cairo
// before 1.17.8 rasterizes subpixel wherever the font it matched carries a
// subpixel order, so every glyph comes out white with a green alpha: pale
// green text, unreadable on light fills and tooltips. A font without an
// order stays gray; MyGo's own subpixel glyphs still ask for their order.
const textColorRules = `<?xml version="1.0"?>
<!DOCTYPE fontconfig SYSTEM "urn:fontconfig:fonts.dtd">
<fontconfig>
  <match target="pattern"><edit name="rgba" mode="assign"><const>none</const></edit></match>
  <match target="font"><edit name="rgba" mode="assign"><const>none</const></edit></match>
</fontconfig>
`

// FixTextColor keeps text in its own color, and reports whether it could:
// without fontconfig, which comes with the GTK MyGo loads, there is nothing
// to change. Call it on the main thread once MyGo is ready, before the first
// window lays out text.
func FixTextColor() bool {
	current, err := purego.Dlsym(purego.RTLD_DEFAULT, "FcConfigGetCurrent")
	if err != nil {
		return false
	}
	load, err := purego.Dlsym(purego.RTLD_DEFAULT, "FcConfigParseAndLoadFromMemory")
	if err != nil {
		return false
	}
	config, _, _ := purego.SyscallN(current)
	if config == 0 {
		return false
	}
	// The rules go after the system's, so they are the ones that count.
	rules := append([]byte(textColorRules), 0)
	ok, _, _ := purego.SyscallN(load, config, uintptr(unsafe.Pointer(&rules[0])), 1)
	return int32(ok) != 0
}
