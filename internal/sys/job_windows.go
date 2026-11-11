//go:build windows

package sys

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// job is the job object this process joins. The handle stays open for the
// life of the process on purpose: Windows closes it when the process ends,
// however it ends, and the job then kills what it holds.
var job windows.Handle

// HoldChildren makes every process this program starts end with it. A shell
// started in a pane is a child of this process, so a window that crashes or is
// killed does not leave pwsh.exe (or wsl.exe, or what runs in it) behind.
//
// It joins the process itself rather than each child: children join a job on
// their own, and one call covers every pane, every split, and the processes
// those shells start later. A child that asks to break away, which only the
// window this one restarts does, leaves the job on purpose.
func HoldChildren() error {
	if job != 0 {
		return nil
	}
	h, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return fmt.Errorf("sys: create job object: %w", err)
	}
	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE | windows.JOB_OBJECT_LIMIT_BREAKAWAY_OK,
		},
	}
	if _, err := windows.SetInformationJobObject(
		h,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
	); err != nil {
		windows.CloseHandle(h)
		return fmt.Errorf("sys: set job limits: %w", err)
	}
	if err := windows.AssignProcessToJobObject(h, windows.CurrentProcess()); err != nil {
		windows.CloseHandle(h)
		return fmt.Errorf("sys: join job object: %w", err)
	}
	job = h
	return nil
}
