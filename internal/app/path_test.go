package app

import (
	"runtime"
	"testing"
)

// TestIsProgramPath covers what counts as the path of a program. The drive
// and UNC cases depend on the platform, so each case says what both expect.
func TestIsProgramPath(t *testing.T) {
	for _, tc := range []struct {
		in                     string
		wantWindows, wantPosix bool
	}{
		{`D:\Applications\pwsh\pwsh.exe`, true, true},
		{`D:\Applications\pwsh\pwsh`, true, false},
		{`C:\Windows\System32\cmd.exe`, true, true},
		{`\\server\share\shell`, true, true},
		{`/usr/bin/zsh`, true, true},
		{`pwsh.exe`, true, true},
		{`todo.txt`, false, false},
		{`~/dir`, false, false},
		{`vim main.go`, false, false},
	} {
		want := tc.wantPosix
		if runtime.GOOS == "windows" {
			want = tc.wantWindows
		}
		if got := isProgramPath(tc.in); got != want {
			t.Errorf("isProgramPath(%q) = %v, want %v", tc.in, got, want)
		}
	}
}
