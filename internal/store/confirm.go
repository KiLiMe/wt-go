package store

import (
	"encoding/json"
)

// Boundary: the confirm setting alone -- a word a file may say, and the one
// question it answers. It reads true and false too, which files before 0.3
// wrote.

// Confirm is when the window asks before it closes a tab that still runs
// something: "running" only when a program runs besides the shell, "always"
// whenever the tab lives, "never" without asking. A file written when the
// setting was true or false reads as "running" and "never".
type Confirm string

// The answers Confirm takes.
const (
	ConfirmRunning Confirm = "running"
	ConfirmAlways  Confirm = "always"
	ConfirmNever   Confirm = "never"
)

// Ask reports whether closing a tab asks first, given whether a program still
// runs in it.
func (c Confirm) Ask(running bool) bool {
	switch c {
	case ConfirmAlways:
		return true
	case ConfirmNever:
		return false
	}
	return running
}

// Valid reports whether c is one of the answers a file may write.
func (c Confirm) Valid() bool {
	switch c {
	case ConfirmRunning, ConfirmAlways, ConfirmNever:
		return true
	}
	return false
}

// UnmarshalJSON reads the answer, and the true or false older files wrote.
func (c *Confirm) UnmarshalJSON(data []byte) error {
	var on bool
	if err := json.Unmarshal(data, &on); err == nil {
		*c = ConfirmNever
		if on {
			*c = ConfirmRunning
		}
		return nil
	}
	var word string
	if err := json.Unmarshal(data, &word); err != nil {
		return err
	}
	*c = Confirm(word)
	return nil
}

// MarshalJSON writes the answer as the word the setting documents.
func (c Confirm) MarshalJSON() ([]byte, error) { return json.Marshal(string(c)) }
