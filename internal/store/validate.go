package store

import (
	"fmt"
	"strings"
)

// Validate reports the settings that are outside the range the window can
// run with. The ranges are wider than the settings page allows: they keep a
// hand-written file from asking for a font no program can draw or a scrollback
// no machine can hold, not to enforce the page's own limits.
func (s Settings) Validate() error {
	var bad []string
	check := func(name string, ok bool, want string) {
		if !ok {
			bad = append(bad, fmt.Sprintf("%s must be %s", name, want))
		}
	}
	check("font.weight", between(float64(s.Font.Weight), 100, 900), "between 100 and 900")
	check("font.size", between(float64(s.Font.Size), 4, 200), "between 4 and 200")
	check("font.lineHeight", between(float64(s.Font.LineHeight), 0.5, 4), "between 0.5 and 4")
	check("padding.x", between(float64(s.Padding.X), 0, 256), "between 0 and 256")
	check("padding.y", between(float64(s.Padding.Y), 0, 256), "between 0 and 256")
	check("backgroundOpacity", between(s.BackgroundOpacity, 0.3, 1), "between 0.3 and 1")
	check("scrollbackMB", between(float64(s.ScrollbackMB), 0, 1024), "between 0 and 1024")
	check("scrollMultiplier", between(float64(s.ScrollMultiplier), 0.1, 10), "between 0.1 and 10")
	check("minimumContrast", between(float64(s.MinimumContrast), 0, 21), "between 0 and 21")
	check("paddingColor", oneOf(s.PaddingColor, enumPaddingColors...), `"background", "extend" or "always"`)
	check("clipboardRead", oneOf(s.ClipboardRead, enumClipboardReads...), `"deny", "allow" or "ask"`)
	check("clipboardWrite", oneOf(s.ClipboardWrite, enumClipboardWrites...), `"allow" or "deny"`)
	check("maxPanes", between(float64(s.MaxPanes), 0, 64), "between 0 and 64")
	check("confirmClose", s.ConfirmClose.Valid(), `"running", "always" or "never"`)
	if len(bad) == 0 {
		return nil
	}
	return fmt.Errorf("settings: %s", strings.Join(bad, "; "))
}

// ProfileIDs names the profiles the file cannot open: an id that is empty, or
// the same as one before it. internal/profile drops those when it builds the
// list of profiles, so the file's entry never appears anywhere; this is how a
// caller can say which line of the file was lost instead of leaving the person
// to wonder. It is not part of Validate because such a file still runs: the
// profiles it names that are fine are the ones the window shows.
func ProfileIDs(profiles []Profile) []string {
	var bad []string
	seen := make(map[string]bool, len(profiles))
	for i, p := range profiles {
		if p.ID == "" {
			bad = append(bad, fmt.Sprintf("profiles[%d].id is empty", i))
			continue
		}
		if seen[p.ID] {
			bad = append(bad, fmt.Sprintf("profiles[%d].id is %q, which a profile above it has", i, p.ID))
		}
		seen[p.ID] = true
	}
	return bad
}

// fillEmptyLists turns a list the file wrote as null into an empty one. Clone
// and Equal compare what the window shows with reflect.DeepEqual, where a nil
// list and an empty one are different answers: a file that said null would
// otherwise read as a change the person never made.
func (s *Settings) fillEmptyLists() {
	if s.Schemes == nil {
		s.Schemes = []Scheme{}
	}
	if s.Profiles == nil {
		s.Profiles = []Profile{}
	}
}

// fallbackUnknown puts back the default for each name the file wrote that no
// setting takes. A word the window does not know is a hand-written mistake, and
// docs/配置参考.md promises it is read as the default rather than refused: the
// ranges in Validate are the other case, a file the window cannot run with. The
// window never writes such a word, so this runs once over a file read from disk
// and leaves an already-clean set exactly as it was.
func (s *Settings) fallbackUnknown() {
	d := Default()
	s.Language = knownLanguage(s.Language)
	s.Theme = knownTheme(s)
	s.StartOn = known(s.StartOn, enumStartOn, d.StartOn)
	s.CursorShape = known(s.CursorShape, enumCursorShapes, d.CursorShape)
	s.PaddingColor = known(s.PaddingColor, enumPaddingColors, d.PaddingColor)
	s.RightClick = known(s.RightClick, enumRightClicks, d.RightClick)
	s.ClipboardRead = known(s.ClipboardRead, enumClipboardReads, d.ClipboardRead)
	s.ClipboardWrite = known(s.ClipboardWrite, enumClipboardWrites, d.ClipboardWrite)
	s.NewTabPosition = known(s.NewTabPosition, enumTabPositions, d.NewTabPosition)
	s.TabWidth = known(s.TabWidth, enumTabWidths, d.TabWidth)
	// A family the file leaves empty is the default families, which is also what
	// the window draws with: an empty setting would say nothing at all.
	s.FillEmptyFont()
	// A weight the file leaves out is the regular one, which is what terminal
	// text is drawn at: a zero says nothing, and the page has to show the face
	// the terminal draws with rather than a weight no family has.
	if s.Font.Weight == 0 {
		s.Font.Weight = d.Font.Weight
	}
	if !s.ConfirmClose.Valid() {
		s.ConfirmClose = d.ConfirmClose
	}
}
