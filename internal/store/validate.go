package store

import (
	"fmt"
	"strings"
)

// Validate reports the settings that are outside the range the window can
// run with. The ranges are wider than the settings page allows: they keep a
// hand-written file from asking for a font no program can draw or a scrollback
// no machine can hold, not to enforce the page's own limits.
func (s Settings) Validate() error {
	var bad []string
	check := func(name string, ok bool, want string) {
		if !ok {
			bad = append(bad, fmt.Sprintf("%s must be %s", name, want))
		}
	}
	check("font.size", between(float64(s.Font.Size), 4, 200), "between 4 and 200")
	check("font.lineHeight", between(float64(s.Font.LineHeight), 0.5, 4), "between 0.5 and 4")
	check("padding.x", between(float64(s.Padding.X), 0, 256), "between 0 and 256")
	check("padding.y", between(float64(s.Padding.Y), 0, 256), "between 0 and 256")
	check("backgroundOpacity", between(s.BackgroundOpacity, 0, 1), "between 0 and 1")
	check("scrollbackMB", between(float64(s.ScrollbackMB), 0, 1024), "between 0 and 1024")
	check("maxPanes", between(float64(s.MaxPanes), 0, 64), "between 0 and 64")
	check("startSize.columns", between(float64(s.StartSize.Columns), 0, 1000), "between 0 and 1000")
	check("startSize.rows", between(float64(s.StartSize.Rows), 0, 1000), "between 0 and 1000")
	if len(bad) == 0 {
		return nil
	}
	return fmt.Errorf("settings: %s", strings.Join(bad, "; "))
}

// between reports whether v is in the range lo through hi, both included.
func between(v, lo, hi float64) bool {
	return v >= lo && v <= hi
}
