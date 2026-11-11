package sys

import (
	"os"
	"runtime"
	"sort"
	"strings"
)

// Redactions are the absolute paths a report may quote, longest first, each
// with the placeholder that says the same thing without naming the account: a
// report pasted into an issue stays useful and stops at "whose machine".
func Redactions() [][2]string {
	home, _ := os.UserHomeDir()
	placeholder := "~"
	if runtime.GOOS == "windows" {
		placeholder = "%USERPROFILE%"
	}
	candidates := [][2]string{
		{os.Getenv("LOCALAPPDATA"), "%LOCALAPPDATA%"},
		{os.Getenv("APPDATA"), "%APPDATA%"},
		{os.Getenv("USERPROFILE"), "%USERPROFILE%"},
		{home, placeholder},
		{os.Getenv("TEMP"), "%TEMP%"},
		{os.Getenv("XDG_CONFIG_HOME"), "$XDG_CONFIG_HOME"},
	}
	seen := make(map[string]bool, len(candidates))
	out := make([][2]string, 0, len(candidates))
	for _, c := range candidates {
		from := strings.TrimRight(strings.TrimSpace(c[0]), `\/`)
		// A root or a bare drive letter would rewrite every path it prefixes.
		if len(from) < 4 || seen[from] {
			continue
		}
		seen[from] = true
		out = append(out, [2]string{from, c[1]})
	}
	sort.SliceStable(out, func(i, j int) bool { return len(out[i][0]) > len(out[j][0]) })
	return out
}
