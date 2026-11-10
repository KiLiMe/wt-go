//go:build !windows

package sys

// HoldChildren is the Windows job object, which the other systems do the same
// way by ending the session the shells run in.
func HoldChildren() error { return nil }
