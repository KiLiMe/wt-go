package pane

import "testing"

// newSet is a set of n tabs that runs nothing: a tab is a pane, and these
// tests only care where the tabs are.
func newSet(n int) *Set {
	s := &Set{}
	for i := 0; i < n; i++ {
		s.Add(&Tab{ProfileID: string(rune('a' + i))})
	}
	return s
}

// TestAdd covers a new tab: it goes at the end and comes in front.
func TestAdd(t *testing.T) {
	s := &Set{}
	s.Add(&Tab{ProfileID: "a"})
	s.Add(&Tab{ProfileID: "b"})
	if len(s.Tabs) != 2 || s.Active != 1 {
		t.Fatalf("tabs = %d, active = %d", len(s.Tabs), s.Active)
	}
	if tab, ok := s.Front(); !ok || tab.ProfileID != "b" {
		t.Errorf("front = %v, %v", tab, ok)
	}
}

// TestRemove covers dropping a tab: the neighbor comes in front, and the
// last tab reports the window has none left.
func TestRemove(t *testing.T) {
	s := newSet(3)
	s.Active = 2
	if !s.Remove(0) {
		t.Fatal("window reported empty")
	}
	if len(s.Tabs) != 2 || s.Active != 1 || s.Tabs[1].ProfileID != "c" {
		t.Errorf("tabs = %v, active = %d", ids(s), s.Active)
	}
	s.Active = 0
	if !s.Remove(1) {
		t.Fatal("window reported empty")
	}
	if tab, _ := s.Front(); tab.ProfileID != "b" {
		t.Errorf("front = %q", tab.ProfileID)
	}
	if s.Remove(0) {
		t.Error("window kept a tab")
	}
	if len(s.Tabs) != 0 || s.Active != 0 {
		t.Errorf("tabs = %v, active = %d", ids(s), s.Active)
	}
}

// TestRemoveActive covers dropping the tab in front: the tab before it comes
// in front, or the first one.
func TestRemoveActive(t *testing.T) {
	s := newSet(2)
	if !s.Remove(1) || s.Active != 0 {
		t.Errorf("active = %d", s.Active)
	}
	s = newSet(3)
	s.Active = 0
	s.Remove(0)
	if s.Active != 0 || s.Tabs[0].ProfileID != "b" {
		t.Errorf("tabs = %v, active = %d", ids(s), s.Active)
	}
}

// TestKeep covers closing the other tabs: the tab in front stays, and the
// dropped tabs come back so their terminals can be closed.
func TestKeep(t *testing.T) {
	s := newSet(3)
	s.Active = 1
	dropped := s.Keep(1)
	if len(s.Tabs) != 1 || s.Tabs[0].ProfileID != "b" || s.Active != 0 {
		t.Errorf("tabs = %v, active = %d", ids(s), s.Active)
	}
	if len(dropped) != 2 || dropped[0].ProfileID != "a" || dropped[1].ProfileID != "c" {
		t.Errorf("dropped = %v", ids(&Set{Tabs: dropped}))
	}
}

// TestMove covers dragging a tab to another place: it lands there and comes
// in front.
func TestMove(t *testing.T) {
	s := newSet(3)
	s.Active = 2
	s.Move(2, 0)
	if got := ids(s); got != "cab" {
		t.Errorf("order = %q", got)
	}
	if s.Active != 0 {
		t.Errorf("active = %d", s.Active)
	}
	s.Move(0, 2)
	if got := ids(s); got != "abc" {
		t.Errorf("order = %q", got)
	}
	s.Move(0, 3)
	if got := ids(s); got != "abc" {
		t.Errorf("out of range moved: %q", got)
	}
}

// TestFind covers finding the tab that runs a profile.
func TestFind(t *testing.T) {
	s := newSet(2)
	if tab, ok := s.Find("b"); !ok || tab != s.Tabs[1] {
		t.Errorf("found = %v, %v", tab, ok)
	}
	if _, ok := s.Find("zzz"); ok {
		t.Error("found a missing profile")
	}
}

func ids(s *Set) string {
	var b []rune
	for _, t := range s.Tabs {
		b = append(b, []rune(t.ProfileID)...)
	}
	return string(b)
}
