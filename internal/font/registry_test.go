package font

import "testing"

func TestNamesIncludeExpected(t *testing.T) {
	got := Names()
	if len(got) == 0 || got[0] != DefaultName {
		t.Fatalf("Names()[0] = %v; want %q first", got, DefaultName)
	}
	want := []string{"block", "compact", "heavy", "standard", "slant", "cyberlarge"}
	for _, w := range want {
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
// nothing, including the go-figure-backed ones.
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

const Charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!?.,:;'-/&%() "

// TestBlockRowWidths enforces the per-glyph equal-width invariant the renderer
// relies on for the hand-authored block fonts.
func TestBlockRowWidths(t *testing.T) {
	for _, name := range []string{"block", "heavy", "compact"} {
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

