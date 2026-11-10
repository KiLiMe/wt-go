package sys

import "os"

// Executable is the exe this process runs from: a report says which build
// answered, and the terminal plugin looks for libghostty-vt next to it.
func Executable() (string, error) { return os.Executable() }

// Getenv is one environment variable, for the report.
func Getenv(key string) string { return os.Getenv(key) }

// OwnPID is this process's id: the parent of every program a terminal runs,
// which is how the window learns which process a new terminal started.
func OwnPID() int { return os.Getpid() }
