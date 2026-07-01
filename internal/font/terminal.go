package font

import "unicode/utf8"

type Terminal struct{}

func (Terminal) Render(line string) []string {
	if line == "" {
		return []string{""}
	}
	return []string{line}
}

func (Terminal) Width(s string) int {
	return utf8.RuneCountInString(s)
}

func init() { register("terminal", Terminal{}) }
