package theme

import (
	"strings"

	"github.com/egoist/mygo/plugins/terminal"
)

// scheme is a scheme written as settings.json writes one.
type scheme struct {
	name, background, foreground, cursor, selection string
	palette                                         [16]string
}

// builtinSchemes are the schemes docs/settings.md lists that Ghostty has not:
// Windows Terminal's own two, and the Solarized, Tango and Vintage of its
// list.
var builtinSchemes = []scheme{
	{
		name:       "Campbell",
		background: "#0c0c0c",
		foreground: "#cccccc",
		cursor:     "#ffffff",
		selection:  "#264f78",
		palette: [16]string{
			"#0c0c0c", "#c50f1f", "#13a10e", "#c19c00", "#0037da", "#881798", "#3a96dd", "#cccccc",
			"#767676", "#e74856", "#16c60c", "#f9f1a5", "#3b78ff", "#b4009e", "#61d6d6", "#f2f2f2",
		},
	},
	{
		name:       "Campbell Powershell",
		background: "#012456",
		foreground: "#cccccc",
		cursor:     "#ffffff",
		selection:  "#264f78",
		palette: [16]string{
			"#0c0c0c", "#c50f1f", "#13a10e", "#c19c00", "#0037da", "#881798", "#3a96dd", "#cccccc",
			"#767676", "#e74856", "#16c60c", "#f9f1a5", "#3b78ff", "#b4009e", "#61d6d6", "#f2f2f2",
		},
	},
	{
		name:       "Solarized Dark",
		background: "#002b36",
		foreground: "#839496",
		cursor:     "#839496",
		selection:  "#073642",
		palette: [16]string{
			"#073642", "#dc322f", "#859900", "#b58900", "#268bd2", "#d33682", "#2aa198", "#eee8d5",
			"#002b36", "#cb4b16", "#586e75", "#657b83", "#839496", "#6c71c4", "#93a1a1", "#fdf6e3",
		},
	},
	{
		name:       "Solarized Light",
		background: "#fdf6e3",
		foreground: "#657b83",
		cursor:     "#657b83",
		selection:  "#eee8d5",
		palette: [16]string{
			"#073642", "#dc322f", "#859900", "#b58900", "#268bd2", "#d33682", "#2aa198", "#eee8d5",
			"#002b36", "#cb4b16", "#586e75", "#657b83", "#839496", "#6c71c4", "#93a1a1", "#fdf6e3",
		},
	},
	{
		name:       "Tango Dark",
		background: "#000000",
		foreground: "#d3d7cf",
		cursor:     "#ffffff",
		selection:  "#ffffff",
		palette: [16]string{
			"#000000", "#cc0000", "#4e9a06", "#c4a000", "#3465a4", "#75507b", "#06989a", "#d3d7cf",
			"#555753", "#ef2929", "#8ae234", "#fce94f", "#729fcf", "#ad7fa8", "#34e2e2", "#eeeeec",
		},
	},
	{
		name:       "Tango Light",
		background: "#ffffff",
		foreground: "#555753",
		cursor:     "#000000",
		selection:  "#000000",
		palette: [16]string{
			"#000000", "#cc0000", "#4e9a06", "#c4a000", "#3465a4", "#75507b", "#06989a", "#d3d7cf",
			"#555753", "#ef2929", "#8ae234", "#fce94f", "#729fcf", "#ad7fa8", "#34e2e2", "#eeeeec",
		},
	},
	{
		name:       "Vintage",
		background: "#000000",
		foreground: "#c0c0c0",
		cursor:     "#c0c0c0",
		selection:  "#000000",
		palette: [16]string{
			"#000000", "#800000", "#008000", "#808000", "#000080", "#800080", "#008080", "#c0c0c0",
			"#808080", "#ff0000", "#00ff00", "#ffff00", "#0000ff", "#ff00ff", "#00ffff", "#ffffff",
		},
	},
}

// builtin is the built in scheme called name, or nil.
func builtin(name string) *terminal.Theme {
	for _, sc := range builtinSchemes {
		if strings.EqualFold(sc.name, name) {
			return sc.theme()
		}
	}
	return nil
}

// BuiltinSchemes are the names of the schemes this program carries, the ones
// Ghostty has not.
func BuiltinSchemes() []string {
	out := make([]string, 0, len(builtinSchemes))
	for _, sc := range builtinSchemes {
		out = append(out, sc.name)
	}
	return out
}

// theme is the colors sc writes.
func (sc scheme) theme() *terminal.Theme {
	t := &terminal.Theme{}
	override(&t.Background, sc.background)
	override(&t.Foreground, sc.foreground)
	override(&t.Cursor, sc.cursor)
	override(&t.Selection, sc.selection)
	for i, s := range sc.palette {
		override(&t.Palette[i], s)
	}
	return t
}
