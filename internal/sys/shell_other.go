//go:build !windows

package sys

import "os"

// Shell is a shell this machine has.
type Shell struct {
	ID      string
	Name    string
	Command string
	Args    []string
}

// Discover is the shells this machine has: here, the user's own shell.
func Discover() []Shell {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return nil
	}
	return []Shell{{ID: "shell", Name: "Shell", Command: shell}}
}
