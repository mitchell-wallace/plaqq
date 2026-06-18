package font

import (
	"testing"
)

func TestRuneWidth(t *testing.T) {
	tests := []struct {
		r        rune
		expected int
	}{
		{'A', 6},
		{'B', 6},
		{' ', 4},
		{'!', 1},
	}

	for _, tt := range tests {
		got := RuneWidth(tt.r)
		if got != tt.expected {
			t.Errorf("RuneWidth(%c) = %d; expected %d", tt.r, got, tt.expected)
		}
	}
}

func TestStringWidth(t *testing.T) {
	tests := []struct {
		s        string
		expected int
	}{
		{"A", 6},
		{"AB", 13},  // 6 ('A') + 1 (spacing) + 6 ('B')
		{"A B", 18}, // 6 ('A') + 1 + 4 (' ') + 1 + 6 ('B')
	}

	for _, tt := range tests {
		got := StringWidth(tt.s)
		if got != tt.expected {
			t.Errorf("StringWidth(%q) = %d; expected %d", tt.s, got, tt.expected)
		}
	}
}

func TestWrapText(t *testing.T) {
	text := "hello world wraps correctly"
	// Word widths:
	// "hello": h(6)+1+e(6)+1+l(5)+1+l(5)+1+o(6) = 31
	// "world": w(8)+1+o(6)+1+r(6)+1+l(5)+1+d(6) = 39
	// Joined "hello world": 31 + 1 + 4 + 1 + 39 = 76

	// Let's set maxWidth = 50. "hello" (31) fits. "hello world" (76) does not. So "world" wraps.
	// "world" (39) fits. "world wraps": 39 + 1 + 4 + 1 + (w(8)+1+r(6)+1+a(6)+1+p(6)+1+s(6) = 34) = 79. Does not fit. So "wraps" wraps.
	// "wraps" (34) + 1 + 4 + 1 + "correctly" (c(6)+1+o(6)+1+r(6)+1+r(6)+1+e(6)+1+c(6)+1+t(7)+1+l(5)+1+y(5) = 55) = 100. Does not fit. So "correctly" wraps.
	// Lines should be:
	// 1: "hello"
	// 2: "world"
	// 3: "wraps"
	// 4: "correctly"

	lines := WrapText(text, 50)
	expectedLines := []string{"hello", "world", "wraps", "correctly"}

	if len(lines) != len(expectedLines) {
		t.Fatalf("expected %d lines, got %d", len(expectedLines), len(lines))
	}

	for i, expected := range expectedLines {
		if lines[i] != expected {
			t.Errorf("line %d = %q; expected %q", i, lines[i], expected)
		}
	}
}
