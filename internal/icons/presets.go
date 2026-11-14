package icons

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// The tab presets are drawn here rather than kept in svg/, as the Remix icons
// are: three palettes of one shape differ only in color, and a file per
// palette would be nine copies of the same path data.
//
// Each shape is drawn for the 16 DIP a tab has, not for a window: strokes are
// an eighth of the viewBox, which is the one DIP line a 16 DIP mark can show,
// and nothing is smaller than that. The dots and the title bar buttons of the
// larger icons people draw for these programs are gone, and the C:\> of cmd is
// a > and a line: at this size they would be a smudge.

// pwsh7Shape is PowerShell 7: a rounded square, a > and a line under it.
const pwsh7Shape = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 128 128">
  <defs>
    <linearGradient id="face" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s"/></linearGradient>
    <linearGradient id="mark" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s"/></linearGradient>
  </defs>
  <rect x="4" y="4" width="120" height="120" rx="32" fill="url(#face)"/>
  <path d="M42 44 L72 64 L42 84" fill="none" stroke="url(#mark)" stroke-width="14" stroke-linecap="round" stroke-linejoin="round"/>
  <line x1="78" y1="84" x2="94" y2="84" stroke="%s" stroke-width="14" stroke-linecap="round"/>
</svg>`

// powershellShape is Windows PowerShell: the same > and line, in a square that
// is squarer than the one PowerShell 7 wears, which is how the two read apart
// at 16 DIP where their marks are the same.
const powershellShape = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 128 128">
  <defs>
    <linearGradient id="face" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s"/></linearGradient>
  </defs>
  <rect x="4" y="4" width="120" height="120" rx="14" fill="url(#face)"/>
  <path d="M42 44 L72 64 L42 84" fill="none" stroke="%s" stroke-width="14" stroke-linecap="round" stroke-linejoin="round"/>
  <line x1="78" y1="84" x2="94" y2="84" stroke="%s" stroke-width="14" stroke-linecap="round"/>
</svg>`

// cmdShape is the command prompt: a window with a title bar, and a > and a
// line in the room under it.
const cmdShape = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 128 128">
  <rect x="4" y="16" width="120" height="96" rx="16" fill="%s"/>
  <path d="M20 16 H108 A16 16 0 0 1 124 32 V40 H4 V32 A16 16 0 0 1 20 16 Z" fill="%s"/>
  <path d="M30 62 L54 76 L30 90" fill="none" stroke="%s" stroke-width="12" stroke-linecap="round" stroke-linejoin="round"/>
  <line x1="60" y1="90" x2="78" y2="90" stroke="%s" stroke-width="12" stroke-linecap="round"/>
</svg>`

// bashShape is bash: a rounded square and a $, for the shell prompt the name
// stands for.
const bashShape = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 128 128">
  <defs>
    <linearGradient id="face" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s"/></linearGradient>
  </defs>
  <rect x="4" y="4" width="120" height="120" rx="32" fill="url(#face)"/>
  <line x1="64" y1="28" x2="64" y2="100" stroke="%s" stroke-width="10" stroke-linecap="round"/>
  <path d="M88 44 C88 32 40 32 40 48 C40 62 88 62 88 78 C88 94 40 94 40 82" fill="none" stroke="%s" stroke-width="10" stroke-linecap="round"/>
</svg>`

// presets holds the marks a tab can wear by name, "shape:palette".
var presets = map[string]*Icon{}

// presetOrder is the presets in the order the settings page lists them: by
// shape, the palettes of a shape together.
var presetOrder []string

func init() {
	for _, p := range []struct {
		key    string
		shape  string
		colors []any
	}{
		{"pwsh7:official", pwsh7Shape, []any{"#2c2c32", "#141417", "#60a5fa", "#2563eb", "#f8fafc"}},
		{"pwsh7:violet", pwsh7Shape, []any{"#2e1065", "#0f0720", "#c084fc", "#7e22ce", "#f8fafc"}},
		{"pwsh7:light", pwsh7Shape, []any{"#ffffff", "#e4e4e7", "#3b82f6", "#1d4ed8", "#18181b"}},
		{"powershell:classic", powershellShape, []any{"#2b71c2", "#012456", "#ffffff", "#93c5fd"}},
		{"powershell:ocean", powershellShape, []any{"#0891b2", "#164e63", "#ffffff", "#a5f3fc"}},
		{"powershell:amber", powershellShape, []any{"#3f3f46", "#18181b", "#fde68a", "#f59e0b"}},
		{"cmd:black", cmdShape, []any{"#0c0c0c", "#e4e4e7", "#f4f4f5", "#a1a1aa"}},
		{"cmd:matrix", cmdShape, []any{"#050a06", "#14532d", "#22c55e", "#4ade80"}},
		{"cmd:amber", cmdShape, []any{"#140d04", "#78350f", "#f59e0b", "#fcd34d"}},
		{"bash:gnu", bashShape, []any{"#3e474f", "#22272b", "#ffffff", "#4eaa25"}},
		{"bash:ubuntu", bashShape, []any{"#5e2750", "#2c001e", "#ffffff", "#e95420"}},
		{"bash:arch", bashShape, []any{"#1e293b", "#0f172a", "#f8fafc", "#1793d1"}},
	} {
		presets[p.key] = &Icon{Name: p.key, SVG: ui.MustParseSVG([]byte(fmt.Sprintf(p.shape, p.colors...)))}
		presetOrder = append(presetOrder, p.key)
	}
}

// Preset is the mark a tab wears by name, nil for a name this package does not
// draw: a settings file may hold a mark a later build took away.
func Preset(key string) *Icon { return presets[key] }

// Presets are the names of the marks, in the order the settings page lists
// them.
func Presets() []string { return append([]string(nil), presetOrder...) }
