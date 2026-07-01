// Package textutil holds small, package-agnostic text helpers shared by the
// resolver (internal/cmd) and the registry packages (internal/border and
// friends). It exists to keep a single source of truth for the edit-distance
// suggestion used by "did you mean ..." style error messages.
package textutil

import "strings"

// SuggestName returns the best near-match for input among options, or
// ("", false) if no candidate is close enough. The match is case-insensitive,
// trims surrounding whitespace, and uses an edit-distance threshold of 2 for
// short names (≤ 6 runes) and 3 for longer names. Ties (two options with the
// same distance) return ("", false) so the suggestion stays unambiguous.
func SuggestName(input string, options []string) (string, bool) {
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return "", false
	}

	best := ""
	bestDistance := 0
	tied := false
	for _, option := range options {
		d := EditDistance(input, strings.ToLower(option))
		if best == "" || d < bestDistance {
			best = option
			bestDistance = d
			tied = false
		} else if d == bestDistance {
			tied = true
		}
	}
	if best == "" || tied {
		return "", false
	}

	limit := 2
	if len([]rune(input)) > 6 || len([]rune(best)) > 6 {
		limit = 3
	}
	if bestDistance > limit {
		return "", false
	}
	return best, true
}

// EditDistance returns the Levenshtein distance between a and b, counted in
// runes (not bytes) so multi-byte characters contribute one unit each.
func EditDistance(a, b string) int {
	ar := []rune(a)
	br := []rune(b)
	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}

	prev := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}

	for i, ca := range ar {
		curr := make([]int, len(br)+1)
		curr[0] = i + 1
		for j, cb := range br {
			cost := 0
			if ca != cb {
				cost = 1
			}
			curr[j+1] = minInt(
				curr[j]+1,
				prev[j+1]+1,
				prev[j]+cost,
			)
		}
		prev = curr
	}
	return prev[len(br)]
}

func minInt(a, b, c int) int {
	if b < a {
		a = b
	}
	if c < a {
		a = c
	}
	return a
}
