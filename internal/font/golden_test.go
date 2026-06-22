package font

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update golden snapshot files")

func TestFontGolden(t *testing.T) {
	for _, name := range Names() {
		t.Run(name, func(t *testing.T) {
			f := Get(name)
			rows := f.Render(Charset)
			actual := strings.Join(rows, "\n") + "\n"

			goldenPath := filepath.Join("testdata", name+".golden")
			if *update {
				err := os.MkdirAll(filepath.Dir(goldenPath), 0755)
				if err != nil {
					t.Fatalf("failed to create testdata dir: %v", err)
				}
				err = os.WriteFile(goldenPath, []byte(actual), 0644)
				if err != nil {
					t.Fatalf("failed to write golden file: %v", err)
				}
			}

			expectedBytes, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("failed to read golden file: %v. Run with -update to generate.", err)
			}
			expected := string(expectedBytes)

			if actual != expected {
				t.Errorf("font %q render does not match golden snapshot.\nExpected:\n%s\nActual:\n%s", name, expected, actual)
			}
		})
	}
}

func TestCharsetCoverage(t *testing.T) {
	charsetRunes := []rune(Charset)

	// We check the three target fonts: block, heavy, compact
	targetFonts := []string{"block", "heavy", "compact"}

	for _, name := range targetFonts {
		t.Run(name, func(t *testing.T) {
			f := Get(name)
			
			// 1. If it's the block (default/legacyBlock) font, check letters map
			if name == "block" {
				for _, r := range charsetRunes {
					if _, ok := letters[r]; !ok {
						t.Errorf("legacy letters map missing rune %q", string(r))
					}
				}
				return
			}

			// 2. Otherwise, check if it's a *Block font
			bf, ok := f.(*Block)
			if !ok {
				t.Fatalf("font %q is not a *Block font (currently %T)", name, f)
			}

			for _, r := range charsetRunes {
				if _, ok := bf.glyphs[r]; !ok {
					t.Errorf("font %q glyph map missing rune %q", name, string(r))
				}
			}
		})
	}
}
