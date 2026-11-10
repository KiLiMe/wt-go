//go:build windows

package sys

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Processes is every process the machine runs.
func Processes() ([]Proc, error) {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snap)
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	out := make([]Proc, 0, 256)
	for err = windows.Process32First(snap, &entry); err == nil; err = windows.Process32Next(snap, &entry) {
		out = append(out, Proc{
			Pid:    int(entry.ProcessID),
			Parent: int(entry.ParentProcessID),
			Name:   windows.UTF16ToString(entry.ExeFile[:]),
		})
	}
	// Walking past the last entry is the loop's end, not a failure.
	if errors.Is(err, windows.ERROR_NO_MORE_FILES) {
		return out, nil
	}
	return out, err
}
