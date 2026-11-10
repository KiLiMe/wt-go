package pane

// Set is the tabs a window has and which one is in front. It owns nothing:
// dropping a tab from it drops what the tab holds, and the caller closes it.
type Set struct {
	Tabs []*Tab
	// Active is the tab in front.
	Active int
}

// Add appends a tab and brings it in front.
func (s *Set) Add(t *Tab) {
	s.Tabs = append(s.Tabs, t)
	s.Active = len(s.Tabs) - 1
}

// Remove drops the tab at i and brings its neighbor in front; the caller
// closes what the tab holds. It reports whether the window has tabs left.
func (s *Set) Remove(i int) bool {
	if i < 0 || i >= len(s.Tabs) {
		return len(s.Tabs) > 0
	}
	s.Tabs = append(s.Tabs[:i:i], s.Tabs[i+1:]...)
	if len(s.Tabs) == 0 {
		s.Active = 0
		return false
	}
	if s.Active > i {
		s.Active--
	}
	if s.Active >= len(s.Tabs) {
		s.Active = len(s.Tabs) - 1
	}
	return true
}

// Keep drops the tabs at keep, and closes nothing: it hands back the tabs it
// dropped so the caller can close what they hold.
func (s *Set) Keep(keep int) []*Tab {
	if keep < 0 || keep >= len(s.Tabs) {
		return nil
	}
	dropped := make([]*Tab, 0, len(s.Tabs)-1)
	for i, t := range s.Tabs {
		if i != keep {
			dropped = append(dropped, t)
		}
	}
	s.Tabs = []*Tab{s.Tabs[keep]}
	s.Active = 0
	return dropped
}

// Move moves the tab at from to to, and brings it in front.
func (s *Set) Move(from, to int) {
	if from < 0 || from >= len(s.Tabs) || to < 0 || to >= len(s.Tabs) || from == to {
		return
	}
	t := s.Tabs[from]
	s.Tabs = append(s.Tabs[:from], s.Tabs[from+1:]...)
	s.Tabs = append(s.Tabs[:to], append([]*Tab{t}, s.Tabs[to:]...)...)
	s.Active = to
}

// Front is the tab in front, and whether there is one.
func (s *Set) Front() (*Tab, bool) {
	if s.Active < 0 || s.Active >= len(s.Tabs) {
		return nil, false
	}
	return s.Tabs[s.Active], true
}

// Find is the tab that runs a profile.
func (s *Set) Find(id string) (*Tab, bool) {
	for _, t := range s.Tabs {
		if t.ProfileID == id {
			return t, true
		}
	}
	return nil, false
}

// Index is where a tab is, -1 while it is not in the set.
func (s *Set) Index(tab *Tab) int {
	for i, t := range s.Tabs {
		if t == tab {
			return i
		}
	}
	return -1
}
