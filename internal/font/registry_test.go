package font

import "testing"

func TestNamesIncludeExpected(t *testing.T) {
	got := Names()
	want := []string{"compact", "terminal", "block", "heavy", "wide"}
	if len(got) != len(want) {
		t.Fatalf("Names() = %v; want exactly %v", got, want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("Names() = %v; want exactly %v", got, want)
		}
		if !Has(w) {
			t.Errorf("expected font %q to be registered; Names()=%v", w, got)
		}
	}
}

func TestGetFallsBackToDefault(t *testing.T) {
	if Get("") != Get(DefaultName) {
		t.Error("Get(\"\") should return the default font")
	}
	if Get("definitely-not-a-font") != Get(DefaultName) {
		t.Error("Get(unknown) should fall back to the default font")
	}
}

// TestEveryFontRenders guards against a registered font that panics or yields
// nothing.
func TestEveryFontRenders(t *testing.T) {
	for _, name := range Names() {
		f := Get(name)
		rows := f.Render("Plaqq 123")
		if len(rows) == 0 {
			t.Errorf("font %q rendered no rows", name)
		}
		if f.Width("Plaqq 123") <= 0 {
			t.Errorf("font %q reported non-positive width", name)
		}
	}
}

// TestBlockRowWidths enforces the per-glyph equal-width invariant the renderer
// relies on for the hand-authored block fonts.
func TestBlockRowWidths(t *testing.T) {
	for _, name := range []string{"block", "heavy", "compact", "wide"} {
		t.Run(name, func(t *testing.T) {
			f := Get(name)
			for _, r := range Charset {
				rows := f.Render(string(r))
				if len(rows) == 0 {
					continue
				}
				w := -1
				for i, row := range rows {
					n := len([]rune(row))
					if w == -1 {
						w = n
					} else if n != w {
						t.Errorf("font %q glyph %q row %d width %d; want %d", name, string(r), i, n, w)
					}
				}
			}
		})
	}
}

func TestWrap(t *testing.T) {
	f := Get("block")

	// Test case 1: normal sentence that fits
	got := Wrap(f, "HELLO WORLD", 200)
	want := []string{"HELLO WORLD"}
	if !equalSlices(got, want) {
		t.Errorf("Wrap(200) = %v; want %v", got, want)
	}

	// Test case 2: wrapping to new line without hyphens
	got = Wrap(f, "HELLO WORLD", 40)
	want = []string{"HELLO", "WORLD"}
	if !equalSlices(got, want) {
		t.Errorf("Wrap(40) = %v; want %v", got, want)
	}

	// Test case 3: word is too wide, must be split with hyphen.
	// Width of "H" = 6, "E" = 6, "L" = 6, "O" = 6, "W" = 8, "-" = 5, spacing = 1
	// Width of "H-" = 6+1+5 = 12
	// Width of "HE-" = 6+1+6+1+5 = 19
	// Width of "HEL-" = 6+1+6+1+6+1+5 = 26
	// Let's set maxWidth to 20.
	// "HE-" (width 19) fits. "HEL-" (width 26) does not.
	// "LL-" (width 19) fits.
	// "OW-" (width 6+1+8+1+5 = 21) does not fit. "O-" (width 12) fits.
	// "WO-" (width 8+1+6+1+5 = 21) does not fit. "W-" (width 14) fits.
	// "OR-" (width 19) fits.
	// "LD" (width 13) fits.
	got = Wrap(f, "HELLOWORLD", 20)
	want = []string{"HE-", "LL-", "O-", "W-", "OR-", "LD"}
	if !equalSlices(got, want) {
		t.Errorf("Wrap(20) = %v; want %v", got, want)
	}

	// Test case 4: maxWidth is too small to even fit a single character with hyphen.
	// It should still make progress (split character by character).
	got = Wrap(f, "ABC", 2)
	want = []string{"A-", "B-", "C"}
	if !equalSlices(got, want) {
		t.Errorf("Wrap(2) = %v; want %v", got, want)
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
