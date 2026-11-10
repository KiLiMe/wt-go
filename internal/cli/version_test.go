package cli

import "testing"

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
