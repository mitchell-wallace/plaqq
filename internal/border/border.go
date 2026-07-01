// Package border wraps already-rendered content rows in a box-drawing frame.
//
// The package is pure: it does not depend on lipgloss, the filesystem, or any
// caller state. Styling (color, bold, padding) is the caller's responsibility;
// Wrap only knows how to glue box-drawing characters around a slice of strings
// while keeping the result a clean rectangle.
package border

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/mitchell-wallace/plaqq/internal/textutil"
)

type Kind string

const (
	None   Kind = "none"
	Single Kind = "single"
	Double Kind = "double"
	Block  Kind = "block"
)

const DefaultName = "none"

var displayOrder = []string{string(None), string(Single), string(Double), string(Block)}

// BlockFrameMode controls how the Block kind renders. It is ignored for every
// other kind. Different faces benefit from different visual weights: the
// single-row terminal face and the 3/5-row block faces get a subtle half-block
// border, while the 7-row heavy face needs an extra leading row and inverted
// halves to keep the frame from feeling cramped.
type BlockFrameMode int

const (
	// BlockFrameSplit is the default Block rendering: the top edge is upper-half
	// blocks (▀) with full-block corners, the bottom edge is lower-half blocks
	// (▄) with full-block corners, and the sides are full blocks. Suits the
	// 1-row terminal face and the 3/5-row block faces.
	BlockFrameSplit BlockFrameMode = iota
	// BlockFrameInverted renders the Block frame flipped vertically with a
	// blank leading row: the top edge is all lower-half blocks (▄), the
	// bottom edge is all upper-half blocks (▀), and the corners are
	// half-blocks (matching the row's half). Sides stay as full blocks.
	// Suits the 7-row heavy face.
	BlockFrameInverted
)

func (m BlockFrameMode) String() string {
	switch m {
	case BlockFrameSplit:
		return "split"
	case BlockFrameInverted:
		return "inverted"
	}
	return ""
}

type glyphs struct {
	topLeft, top, topRight          string
	side                            string
	bottomLeft, bottom, bottomRight string
}

var glyphsByKind = map[Kind]glyphs{
	Single: {
		topLeft:     "┌",
		top:         "─",
		topRight:    "┐",
		side:        "│",
		bottomLeft:  "└",
		bottom:      "─",
		bottomRight: "┘",
	},
	Double: {
		topLeft:     "╔",
		top:         "═",
		topRight:    "╗",
		side:        "║",
		bottomLeft:  "╚",
		bottom:      "═",
		bottomRight: "╝",
	},
}

func glyphsFor(k Kind, mode BlockFrameMode) glyphs {
	if k != Block {
		return glyphsByKind[k]
	}
	switch mode {
	case BlockFrameInverted:
		return glyphs{
			topLeft:     "▄",
			top:         "▄",
			topRight:    "▄",
			side:        "█",
			bottomLeft:  "▀",
			bottom:      "▀",
			bottomRight: "▀",
		}
	default:
		return glyphs{
			topLeft:     "█",
			top:         "▀",
			topRight:    "█",
			side:        "█",
			bottomLeft:  "█",
			bottom:      "▄",
			bottomRight: "█",
		}
	}
}

func (k Kind) String() string {
	return string(k)
}

func Has(name string) bool {
	_, err := Parse(name)
	return err == nil
}

func Names() []string {
	out := make([]string, len(displayOrder))
	copy(out, displayOrder)
	return out
}

func Parse(name string) (Kind, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return None, nil
	}
	lower := strings.ToLower(trimmed)
	if k, ok := matchKind(lower); ok {
		return k, nil
	}
	return None, fmt.Errorf("unknown frame %q: valid frames are %s%s", lower, strings.Join(Names(), ", "), suggest(trimmed))
}

func matchKind(lower string) (Kind, bool) {
	switch lower {
	case string(None):
		return None, true
	case string(Single):
		return Single, true
	case string(Double):
		return Double, true
	case string(Block):
		return Block, true
	}
	return "", false
}

func suggest(input string) string {
	s, ok := textutil.SuggestName(input, Names())
	if !ok {
		return ""
	}
	return fmt.Sprintf("; did you mean %q?", s)
}

func (k Kind) Wrap(rows []string, hMargin int, mode BlockFrameMode) []string {
	if k == None || len(rows) == 0 {
		return rows
	}

	inner := computeInnerWidth(rows, hMargin)
	if inner < 0 {
		inner = 0
	}

	g, ok := glyphsByKind[k]
	if k == Block || ok {
		g = glyphsFor(k, mode)
	}
	if g == (glyphs{}) {
		return rows
	}

	padded := make([]string, len(rows))
	for i, r := range rows {
		padded[i] = padRow(r, inner, hMargin)
	}

	out := make([]string, 0, len(padded)+4)
	out = append(out, buildEdge(g.topLeft, g.top, g.topRight, inner))
	if k == Block && mode == BlockFrameInverted {
		out = append(out, g.side+strings.Repeat(" ", inner)+g.side)
	}
	for _, row := range padded {
		out = append(out, g.side+row+g.side)
	}
	if k == Block && mode == BlockFrameInverted {
		out = append(out, g.side+strings.Repeat(" ", inner)+g.side)
	}
	out = append(out, buildEdge(g.bottomLeft, g.bottom, g.bottomRight, inner))
	return out
}

func computeInnerWidth(rows []string, hMargin int) int {
	maxContent := 0
	anyNonBlank := false
	for _, r := range rows {
		if strings.TrimSpace(r) == "" {
			continue
		}
		anyNonBlank = true
		if w := utf8.RuneCountInString(r); w > maxContent {
			maxContent = w
		}
	}
	if !anyNonBlank {
		return 2 * hMargin
	}
	return maxContent + 2*hMargin
}

func padRow(r string, inner, hMargin int) string {
	if strings.TrimSpace(r) == "" {
		return strings.Repeat(" ", inner)
	}
	textWidth := inner - 2*hMargin
	if textWidth < 0 {
		textWidth = 0
	}
	current := utf8.RuneCountInString(r)
	if current < textWidth {
		r += strings.Repeat(" ", textWidth-current)
	}
	return strings.Repeat(" ", hMargin) + r + strings.Repeat(" ", hMargin)
}

func buildEdge(left, edge, right string, inner int) string {
	return left + strings.Repeat(edge, inner) + right
}
