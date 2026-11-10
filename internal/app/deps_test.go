package app_test

import (
	"os/exec"
	"slices"
	"strings"
	"testing"
)

// rules are the internal packages each package may import, as
// docs/代码导读.md draws the graph.
var rules = map[string][]string{
	"wtgo/cmd/wtgo":         {"wtgo/internal/app"},
	"wtgo/internal/app":     {"wtgo/internal/action", "wtgo/internal/i18n", "wtgo/internal/layout", "wtgo/internal/pane", "wtgo/internal/profile", "wtgo/internal/store", "wtgo/internal/sys", "wtgo/internal/theme", "wtgo/internal/ui"},
	"wtgo/internal/ui":      {"wtgo/internal/action", "wtgo/internal/i18n", "wtgo/internal/layout", "wtgo/internal/pane", "wtgo/internal/profile", "wtgo/internal/store", "wtgo/internal/theme"},
	"wtgo/internal/theme":   {"wtgo/internal/store"},
	"wtgo/internal/pane":    {"wtgo/internal/layout"},
	"wtgo/internal/layout":  {},
	"wtgo/internal/profile": {"wtgo/internal/store"},
	"wtgo/internal/store":   {"wtgo/internal/action"},
	"wtgo/internal/action":  {"wtgo/internal/i18n"},
	"wtgo/internal/sys":     {},
	"wtgo/internal/i18n":    {},
}

// TestDeps covers the dependency rules: a package imports the internal
// packages the graph allows, and no other.
func TestDeps(t *testing.T) {
	for pkg, allowed := range rules {
		for _, imp := range imports(t, pkg) {
			if strings.HasPrefix(imp, "wtgo/") && !slices.Contains(allowed, imp) {
				t.Errorf("%s imports %s", pkg, imp)
			}
		}
	}
}

// TestPure covers the packages that parse and draw rather than load: they
// read no file and start no process.
func TestPure(t *testing.T) {
	for _, pkg := range []string{"wtgo/internal/action", "wtgo/internal/i18n", "wtgo/internal/layout", "wtgo/internal/pane", "wtgo/internal/profile", "wtgo/internal/store", "wtgo/internal/theme", "wtgo/internal/ui"} {
		for _, imp := range imports(t, pkg) {
			switch imp {
			case "io/fs", "os", "os/exec", "path/filepath", "golang.org/x/sys/windows/registry":
				t.Errorf("%s imports %s", pkg, imp)
			}
		}
	}
}

// imports are the packages pkg imports.
func imports(t *testing.T, pkg string) []string {
	t.Helper()
	out, err := exec.Command("go", "list", "-f", "{{join .Imports \"\\n\"}}", pkg).Output()
	if err != nil {
		t.Fatalf("go list %s: %v", pkg, err)
	}
	var imports []string
	for _, line := range strings.Split(string(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			imports = append(imports, line)
		}
	}
	return imports
}
