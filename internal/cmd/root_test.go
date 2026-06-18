package cmd

import "testing"

func TestParseColor(t *testing.T) {
	valid := []string{"#fff", "#00f5d4", "#ABCDEF", "0", "255", "213", " #fff "}
	for _, s := range valid {
		if _, err := parseColor(s); err != nil {
			t.Errorf("parseColor(%q) returned error: %v", s, err)
		}
	}

	invalid := []string{"", "#", "#zz", "#12", "#1234567", "256", "-1", "nope", "ff00ff"}
	for _, s := range invalid {
		if _, err := parseColor(s); err == nil {
			t.Errorf("parseColor(%q) expected error, got nil", s)
		}
	}
}
