package font

import (
	"strings"

	figure "github.com/common-nighthawk/go-figure"
)

// figletFonts is the curated subset of go-figure's bundled FIGlet fonts that
// plaqq exposes. These are ASCII line-art fonts (a different aesthetic from the
// Unicode block fonts); the full library ships ~150, but a tasteful handful is
// far easier to pick from than an exhaustive list.
var figletFonts = []string{
	"standard",
	"slant",
	"banner",
	"big",
	"small",
	"doom",
	"larry3d",
	"mini",
	"cyberlarge",
}

// figletFont adapts a single go-figure font to the Font interface.
type figletFont struct{ name string }

// render is the panic-safe core: go-figure can panic on glyphs a font lacks, so
// any failure degrades to empty output rather than taking the program down.
func (f figletFont) render(line string) (rows []string) {
	defer func() { _ = recover() }()
	out := figure.NewFigure(line, f.name, true).String()
	rows = strings.Split(strings.TrimRight(out, "\n"), "\n")
	return trimBlankEdges(rows)
}

func (f figletFont) Render(line string) []string { return f.render(line) }

func (f figletFont) Width(s string) int {
	w := 0
	for _, row := range f.render(s) {
		if n := len([]rune(row)); n > w {
			w = n
		}
	}
	return w
}

// trimBlankEdges drops fully blank leading and trailing rows so that fonts with
// padding rows still center cleanly.
func trimBlankEdges(rows []string) []string {
	start, end := 0, len(rows)
	for start < end && strings.TrimSpace(rows[start]) == "" {
		start++
	}
	for end > start && strings.TrimSpace(rows[end-1]) == "" {
		end--
	}
	return rows[start:end]
}

// mappedFiglet renders an underlying FIGlet font and rewrites individual glyph
// characters one-for-one. It powers "block" fonts derived from a clean ASCII
// face: banner3 has crisp, legible letterforms drawn entirely from '#', so
// swapping '#' for a Unicode full block yields a solid block font for free.
type mappedFiglet struct {
	base figletFont
	repl *strings.Replacer
}

func (m mappedFiglet) Render(line string) []string {
	rows := m.base.render(line)
	for i, row := range rows {
		rows[i] = m.repl.Replace(row)
	}
	return rows
}

// Width is unchanged by the replacements, which map a single rune to a single
// rune, so it defers to the base font.
func (m mappedFiglet) Width(s string) int { return m.base.Width(s) }

func init() {
	for _, name := range figletFonts {
		register(name, figletFont{name: name})
	}
	// "heavy" is banner3's solid letterforms rendered in Unicode blocks — a
	// bold companion to the lighter built-in block font.
	register("heavy", mappedFiglet{
		base: figletFont{name: "banner3"},
		repl: strings.NewReplacer("#", "█"),
	})
}
