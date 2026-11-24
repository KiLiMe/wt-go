package sys

// Boundary: window-state.json -- the window's place, size and state, beside the
// settings, since the geometry a window was left in is a preference too.

// windowStateName is where a window was left: beside settings.json, which is
// the other thing a person edits and the window reads.
const windowStateName = "window-state.json"

// WindowStatePath is where the window's place, size and state live.
func WindowStatePath() (string, error) { return configFile(windowStateName) }

// ReadWindowState is window-state.json, and whether there was one.
func ReadWindowState() ([]byte, bool, error) { return readConfigFile(windowStateName) }

// WriteWindowState makes the settings directory and writes window-state.json.
func WriteWindowState(b []byte) error { return writeConfigFile(windowStateName, b) }
