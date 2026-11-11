//go:build darwin

package sys

import "golang.org/x/sys/unix"

// Processes is every process the machine runs, as the kernel lists them: macOS
// has no /proc, so the window asks sysctl for the table instead. The state it
// gives as a number is written the way Linux writes it, so that the rest of the
// window reads one kind of state.
func Processes() ([]Proc, error) {
	procs, err := unix.SysctlKinfoProcSlice("kern.proc.all")
	if err != nil {
		return nil, err
	}
	out := make([]Proc, 0, len(procs))
	for _, p := range procs {
		out = append(out, Proc{
			Pid:    int(p.Proc.P_pid),
			Parent: int(p.Eproc.Ppid),
			Name:   unix.ByteSliceToString(p.Proc.P_comm[:]),
			State:  stateLetter(p.Proc.P_stat),
		})
	}
	return out, nil
}

// stateLetter is what the system says a process does, written as Linux writes
// it: idle, running, sleeping, stopped, or a zombie that has ended.
func stateLetter(stat int8) byte {
	switch stat {
	case 1:
		return 'I'
	case 2:
		return 'R'
	case 3:
		return 'S'
	case 4:
		return 'T'
	case 5:
		return 'Z'
	}
	return 0
}
