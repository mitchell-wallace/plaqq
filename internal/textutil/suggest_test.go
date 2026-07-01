package textutil

import "testing"

func TestSuggestName(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		options []string
		want    string
		wantOK  bool
	}{
		{"empty input", "", []string{"a", "b"}, "", false},
		{"exact match", "hello", []string{"hello", "world"}, "hello", true},
		{"case-insensitive", "HELLO", []string{"hello", "world"}, "hello", true},
		{"trimmed", "  hello  ", []string{"hello"}, "hello", true},
		{"close miss", "singl", []string{"none", "single", "double", "block"}, "single", true},
		{"distant miss", "xyzqqq", []string{"none", "single", "double", "block"}, "", false},
		{"ties return false", "ab", []string{"aa", "bb", "cc"}, "", false},
		{"no options", "anything", nil, "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := SuggestName(c.input, c.options)
			if ok != c.wantOK {
				t.Fatalf("SuggestName(%q, %v) ok = %v; want %v", c.input, c.options, ok, c.wantOK)
			}
			if got != c.want {
				t.Errorf("SuggestName(%q, %v) = %q; want %q", c.input, c.options, got, c.want)
			}
		})
	}
}

func TestEditDistance(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"kitten", "sitting", 3},
		{"flaw", "lawn", 2},
		{"héllo", "hello", 1},
	}
	for _, c := range cases {
		t.Run(c.a+"_"+c.b, func(t *testing.T) {
			if got := EditDistance(c.a, c.b); got != c.want {
				t.Errorf("EditDistance(%q, %q) = %d; want %d", c.a, c.b, got, c.want)
			}
		})
	}
}
