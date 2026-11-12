package app

import (
	"encoding/json"

	"github.com/egoist/mygo"

	"wtgo/internal/sys"
)

// windowState keeps the window's place, size and state in window-state.json
// beside the settings, where mygo would keep them in its own folder instead:
// one folder for what the window was told and where it was left.
type windowState struct{}

// LoadWindowStates is what the last save left. A damaged file reads as no
// states, and the next save replaces it, as mygo's own store does.
func (windowState) LoadWindowStates() (map[string]mygo.WindowState, error) {
	b, ok, err := sys.ReadWindowState()
	if err != nil || !ok {
		return nil, err
	}
	var states map[string]mygo.WindowState
	if err := json.Unmarshal(b, &states); err != nil {
		return nil, err
	}
	return states, nil
}

func (windowState) SaveWindowStates(states map[string]mygo.WindowState) error {
	b, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		return err
	}
	return sys.WriteWindowState(b)
}
