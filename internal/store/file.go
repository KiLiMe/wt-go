package store

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Boundary: settings.json as bytes -- the parse that keeps what it does not
// know, the keys a file may name, and the write that puts the unknown ones
// back. The shape it parses into is settings.go.

// Parse reads settings.json: a key the file leaves out keeps its default,
// so a shorter file is still a whole set of settings. A file that is JSON but
// names a setting outside the range the window can run with is rejected whole:
// the caller keeps the settings it already has rather than running a
// half-applied file. Parse answers with no settings when it rejects a file, so
// a caller cannot mistake the defaults for what the file said.
func Parse(b []byte) (Settings, error) {
	s := Default()
	if err := json.Unmarshal(b, &s); err != nil {
		return Settings{}, fmt.Errorf("settings: %w", err)
	}
	s.fillEmptyLists()
	// A key no field names is kept as it was written, so that saving from the
	// settings page does not drop what a person or a later version put there.
	// Reading b twice rather than teaching Settings to read itself keeps the
	// struct a plain shape to write, and a file that is not an object at all is
	// already rejected above.
	var all map[string]json.RawMessage
	if err := json.Unmarshal(b, &all); err != nil {
		return Settings{}, fmt.Errorf("settings: %w", err)
	}
	for key := range knownKeys {
		delete(all, key)
	}
	if len(all) > 0 {
		s.Extra = all
	} else {
		s.Extra = nil
	}
	if tag, ok := languageNames[s.Language]; ok {
		s.Language = tag
	}
	// A word no setting takes is a hand-written mistake, not a range to
	// refuse a file over: it goes back to the default, as docs/配置参考.md
	// says of the names that are not one of the words. The ranges below are
	// different — a font no program can draw is a file the window cannot run.
	s.fallbackUnknown()
	if err := s.Validate(); err != nil {
		return Settings{}, err
	}
	return s, nil
}

// knownKeys are the keys Settings has a field for, and so the ones Extra does
// not keep. A key added to Settings without being added here would be kept in
// Extra too, which Marshal writes second: the field would win, and the test of
// the two lists would say so.
var knownKeys = map[string]struct{}{
	"language": {}, "theme": {}, "defaultProfile": {}, "startOn": {},
	"startDirectory": {},

	"font": {}, "colorScheme": {}, "padding": {}, "backgroundOpacity": {},
	"cursorShape": {}, "cursorBlink": {}, "pointerText": {}, "scrollbackMB": {},
	"paddingColor": {}, "minimumContrast": {},

	"copyOnSelect": {}, "rightClick": {}, "rightClickCopyPastePriority": {},
	"pasteMultilineConfirm": {},
	"confirmClose":          {}, "closeOnLastTab": {}, "confirmCloseWindow": {},
	"scrollMultiplier": {}, "clipboardRead": {}, "clipboardWrite": {},
	"links": {},

	"newTabPosition": {}, "tabWidth": {}, "alwaysOnTop": {},
	"titleFromActiveTerminal": {},

	"maxPanes": {},

	"schemes": {}, "profiles": {}, "keybindings": {},
}

// Marshal writes settings.json, with the keys Parse kept that no field names:
// a file that a person extended by hand, or one a later version wrote, keeps
// what this window does not understand. A key a field names is written as the
// field now holds it, except the families the window ships: those are left out,
// since a file that leaves the family out reads as the default families, and a
// machine drawing with the default list follows the list the version it runs
// ships rather than the one it was first opened with. Every other setting is
// written as it is, so that a file says what a window does.
func Marshal(s Settings) ([]byte, error) {
	if s.Font.Family == DefaultFontFamily() {
		s.Font.Family = ""
	}
	b, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("settings: %w", err)
	}
	if len(s.Extra) > 0 {
		fields := map[string]json.RawMessage{}
		if err := json.Unmarshal(b, &fields); err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}
		for k, v := range s.Extra {
			if _, known := fields[k]; !known {
				fields[k] = v
			}
		}
		if b, err = json.Marshal(fields); err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}
	}
	var indented bytes.Buffer
	if err := json.Indent(&indented, b, "", "  "); err != nil {
		return nil, fmt.Errorf("settings: %w", err)
	}
	indented.WriteByte('\n')
	return indented.Bytes(), nil
}
