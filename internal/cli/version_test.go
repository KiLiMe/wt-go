package cli

import (
	"strings"
	"testing"
)

func TestTidyVersion(t *testing.T) {
	cases := map[string]string{
		"v0.1.1-0.20261110153946-e9f374cbc6b0":       "v0.1.1-dev",
		"v0.1.1-0.20261110153946-e9f374cbc6b0+dirty": "v0.1.1-dev",
		"v0.2.0":  "v0.2.0",
		"devel":   "devel",
		"unknown": "unknown",
	}
	for in, want := range cases {
		if got := tidyVersion(in); got != want {
			t.Errorf("tidyVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

// A build a script stamped names the release it was packaged as, beside the
// build info that says which commit answered.
func TestPackagedVersion(t *testing.T) {
	old := packagedVersion
	packagedVersion = "0.3.0"
	t.Cleanup(func() { packagedVersion = old })

	out, _, handled, code := run(t, "version")
	if !handled || code != 0 || !strings.Contains(out, "(packaged 0.3.0)") {
		t.Errorf("Run(version) = %q, %v, %d, want the packaged version in it", out, handled, code)
	}
}
