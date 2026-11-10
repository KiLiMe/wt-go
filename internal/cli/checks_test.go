package cli

import (
	"strings"
	"testing"

	"wtgo/internal/fault"
)

// codes is the codes checks() answers for a report, keyed by code.
func codes(r Report) map[fault.Code]Check {
	out := map[fault.Code]Check{}
	for _, c := range checks(r) {
		out[c.Code] = c
	}
	return out
}

func TestChecksPass(t *testing.T) {
	r := Report{
		Settings: SettingsReport{Path: `C:\x\settings.json`, Keybindings: 36},
		Native:   NativeReport{Path: "ghostty-vt.dll", Found: true},
		Profiles: []ProfileReport{{ID: "pwsh7", Command: "pwsh.exe", Available: true}},
	}
	for code, c := range codes(r) {
		if !c.OK {
			t.Errorf("%s failed on a machine that is fine: %s", code, c.Detail)
		}
	}
}

func TestChecksFail(t *testing.T) {
	cases := []struct {
		name string
		r    Report
		want fault.Code
	}{
		{
			name: "a file that is not JSON",
			r:    Report{Settings: SettingsReport{Path: "p", Error: "settings: invalid character '}' looking for beginning of object key string"}},
			want: fault.BadSettings,
		},
		{
			name: "a field out of range",
			r:    Report{Settings: SettingsReport{Path: "p", Error: "settings: font.size must be between 4 and 200"}},
			want: fault.BadField,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.r.Native = NativeReport{Found: true}
			tc.r.Profiles = []ProfileReport{{ID: "pwsh7", Available: true}}
			tc.r.Checks = checks(tc.r)
			failed := tc.r.Failed()
			if len(failed) != 1 || failed[0].Code != tc.want {
				t.Fatalf("Failed() = %+v, want just %s", failed, tc.want)
			}
			if !strings.Contains(tc.r.Explain(), string(tc.want)) {
				t.Errorf("Explain() does not name %s", tc.want)
			}
		})
	}
}

func TestChecksNoNative(t *testing.T) {
	r := Report{
		Settings: SettingsReport{Path: "p"},
		Profiles: []ProfileReport{{ID: "pwsh7", Command: "pwsh.exe", Available: true}},
	}
	c, ok := codes(r)[fault.NativeLib]
	if !ok || c.OK {
		t.Fatalf("E102 = %+v, want it to fail without a library", c)
	}
	if !strings.Contains(c.Detail, "libghostty-vt") {
		t.Errorf("detail = %q, want where it was looked for", c.Detail)
	}
}

func TestChecksAMissingProfile(t *testing.T) {
	r := Report{
		Settings: SettingsReport{Path: "p", Error: ""},
		Native:   NativeReport{Found: true},
		Profiles: []ProfileReport{
			{ID: "pwsh7", Command: "pwsh.exe", Available: true},
			{ID: "wsl", Command: "wsl.exe", Available: false},
		},
	}
	c := codes(r)[fault.ChildFailed]
	if c.OK || !strings.Contains(c.Detail, "wsl") {
		t.Errorf("E301 = %+v, want the profile that cannot start", c)
	}
}
