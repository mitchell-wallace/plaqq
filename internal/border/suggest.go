package border

import "strings"

func suggestName(input string, options []string) (string, bool) {
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return "", false
	}

	best := ""
	bestDistance := 0
	tied := false
	for _, option := range options {
		d := editDistance(input, strings.ToLower(option))
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

func editDistance(a, b string) int {
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
