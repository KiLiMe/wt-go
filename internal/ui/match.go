package ui

import (
	"sort"
	"strings"
	"unicode"
)

// Boundary: matching pure functions -- a filter against a command, and against
// any string at all. No drawing and no state: the palette calls Match, and so
// may anything else that ranks text by what a person typed.

// Match is the commands that answer a filter: the ones whose title holds what
// is typed, in order, before the ones a word of the group or the action names,
// and every command for an empty filter.
func Match(filter string, cmds []Command) []Command {
	filter = strings.TrimSpace(filter)
	if filter == "" {
		return append([]Command{}, cmds...)
	}
	type scored struct {
		cmd   Command
		score int
	}
	hits := make([]scored, 0, len(cmds))
	for _, cmd := range cmds {
		score := match(filter, cmd)
		if score < 0 {
			continue
		}
		hits = append(hits, scored{cmd, score})
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	out := make([]Command, len(hits))
	for i, h := range hits {
		out[i] = h.cmd
	}
	return out
}

// match scores how well a filter finds a command, -1 for one it does not: the
// title as letters in order, which counts for more, or the group and the words
// of the action, each the start of a word.
func match(filter string, cmd Command) int {
	score := fuzzy(filter, cmd.Title)
	if score >= 0 {
		// A title that holds what is typed to its letters in order is the
		// command the reader most likely means.
		score += 50
	}
	return max(score, words(filter, cmd.Group+" "+strings.ReplaceAll(cmd.Action, ".", " ")))
}

// fuzzy scores how well query matches s, as its letters in order; start of
// words and runs of letters score more, and -1 is no match.
func fuzzy(query, s string) int {
	if query == "" {
		return 0
	}
	q := []rune(strings.ToLower(query))
	r := []rune(strings.ToLower(s))
	score, qi, run := 0, 0, 0
	for i := 0; i < len(r) && qi < len(q); i++ {
		if q[qi] == ' ' {
			qi++
			continue
		}
		if r[i] != q[qi] {
			run = 0
			continue
		}
		bonus := 1
		if i == 0 || !unicode.IsLetter(r[i-1]) && !unicode.IsDigit(r[i-1]) {
			bonus += 6
		}
		run++
		score += bonus + run*2
		qi++
	}
	if qi < len(q) {
		return -1
	}
	return score*10 - len(r)
}

// words scores how well query finds words, a list of them: each term of the
// query the start of one.
func words(query, list string) int {
	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		return -1
	}
	score := 0
	for _, t := range terms {
		best := -1
		for _, w := range strings.Fields(strings.ToLower(list)) {
			if strings.HasPrefix(w, t) {
				best = max(best, fuzzy(t, w))
			}
		}
		if best < 0 {
			return -1
		}
		score += best
	}
	return score
}
