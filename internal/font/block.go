package font

import "strings"

// Block is a fixed-height glyph font: every rune maps to the same number of
// rows, and a glyph's rows all share one width. Unknown runes fall back to '?'.
// Text is upper-cased before lookup, matching the block aesthetic.
type Block struct {
	glyphs  map[rune][]string
	height  int
	spacing int
	missing []string
}

// NewBlock builds a Block from a glyph table. The height is taken from the
// glyphs, and a single space column separates adjacent letters. The table must
// contain a '?' glyph to render unknown runes.
func NewBlock(glyphs map[rune][]string) *Block {
	height := 0
	for _, g := range glyphs {
		height = len(g)
		break
	}
	b := &Block{glyphs: glyphs, height: height, spacing: 1}
	if g, ok := glyphs['?']; ok {
		b.missing = g
	} else {
		b.missing = make([]string, height)
	}
	return b
}

func (b *Block) glyph(r rune) []string {
	if g, ok := b.glyphs[r]; ok {
		return g
	}
	return b.missing
}

// Render lays the glyphs of line side by side, returning b.height rows.
func (b *Block) Render(line string) []string {
	runes := []rune(strings.ToUpper(line))
	rows := make([]string, b.height)
	gap := strings.Repeat(" ", b.spacing)
	for i, r := range runes {
		g := b.glyph(r)
		for row := 0; row < b.height; row++ {
			if i > 0 {
				rows[row] += gap
			}
			if row < len(g) {
				rows[row] += g[row]
			}
		}
	}
	return rows
}

// Width returns the rendered column width of s.
func (b *Block) Width(s string) int {
	runes := []rune(strings.ToUpper(s))
	w := 0
	for i, r := range runes {
		if i > 0 {
			w += b.spacing
		}
		g := b.glyph(r)
		if len(g) > 0 {
			w += len([]rune(g[0]))
		}
	}
	return w
}

// blockGlyphs is the default medium-weight block font ("block"): a clean,
// legible, 5-row face drawn from full and half Unicode block elements. It is
// the workhorse face and the authoring pattern the other fonts follow.
//
// Provenance / licence: hand-authored for plaqq. The design referenced the
// letterform topology of the `cfonts` `block`/`simpleBlock` faces, but those
// are distributed under GPL-3.0 (verified against the upstream LICENSE at
// https://github.com/dominikwilkowski/cfonts), which is incompatible with
// plaqq's MIT licence. No cfonts glyph rasterization was transcribed; every
// glyph below is an original drawing. Letter shapes themselves are not
// copyrightable; the specific renderings are.
//
// Authoring rules (enforced by TestBlockRowWidths): uppercase-only; every row
// of a given glyph shares one display width; '?' is the not-found marker.
var blockGlyphs = map[rune][]string{
	'A': {
		"  ▄▄▄▄  ",
		" █    █ ",
		"█▀▀▀▀▀▀█",
		"█      █",
		"█      █",
	},
	'B': {
		"████▀▀  ",
		"█    █  ",
		"████▄▄  ",
		"█    █  ",
		"████▀▀  ",
	},
	'C': {
		" ▄████▄ ",
		"█▀    ▀ ",
		"█       ",
		"█▄    ▄ ",
		" ▀████▀ ",
	},
	'D': {
		"████▀▀  ",
		"█    █  ",
		"█    █  ",
		"█    █  ",
		"████▀▀  ",
	},
	'E': {
		"███████ ",
		"█       ",
		"█████▀  ",
		"█       ",
		"███████ ",
	},
	'F': {
		"███████ ",
		"█       ",
		"█████▀  ",
		"█       ",
		"█       ",
	},
	'G': {
		" ▄████▄ ",
		"█▀    ▀ ",
		"█  ████ ",
		"█     █ ",
		" ▀████▀ ",
	},
	'H': {
		"█     █ ",
		"█     █ ",
		"███████ ",
		"█     █ ",
		"█     █ ",
	},
	'I': {
		"█████",
		"  █  ",
		"  █  ",
		"  █  ",
		"█████",
	},
	'J': {
		"  ████▄ ",
		"      █ ",
		"      █ ",
		"█     █ ",
		" ▀████▀ ",
	},
	'K': {
		"█    █▄ ",
		"█  █▄ █ ",
		"█████▀  ",
		"█  ▀█ █ ",
		"█    ▀█ ",
	},
	'L': {
		"█       ",
		"█       ",
		"█       ",
		"█       ",
		"███████ ",
	},
	'M': {
		"█     █ ",
		"██   ██ ",
		"█ █ █ █ ",
		"█  █  █ ",
		"█     █ ",
	},
	'N': {
		"█     █ ",
		"██    █ ",
		"█ █   █ ",
		"█  █  █ ",
		"█    ██ ",
	},
	'O': {
		" ▄███▄  ",
		"█    █  ",
		"█    █  ",
		"█    █  ",
		" ▀███▀  ",
	},
	'P': {
		"████▀▀  ",
		"█    █  ",
		"████▀▀  ",
		"█       ",
		"█       ",
	},
	'Q': {
		" ▄███▄  ",
		"█    █  ",
		"█    █  ",
		"█  █ █  ",
		" ▀███▄▄ ",
	},
	'R': {
		"████▀▀  ",
		"█    █  ",
		"████▄▄  ",
		"█  ▀█   ",
		"█   █▄  ",
	},
	'S': {
		" ▄████▄ ",
		"█▀    ▀ ",
		" ▄████▀ ",
		"▄    ▄█ ",
		" ▀████▀ ",
	},
	'T': {
		"███████",
		"   █   ",
		"   █   ",
		"   █   ",
		"   █   ",
	},
	'U': {
		"█     █ ",
		"█     █ ",
		"█     █ ",
		"█     █ ",
		" ▀███▀  ",
	},
	'V': {
		"█     █ ",
		"█     █ ",
		"█     █ ",
		" ▄   ▄  ",
		"  ▀▀▀   ",
	},
	'W': {
		"█      █ ",
		"█  █  █  ",
		"█  █  █  ",
		"█ ▄▀▀▄ █ ",
		" ▀    ▀  ",
	},
	'X': {
		"█    █▄ ",
		" ▄  █▀  ",
		"  ▀▀▀   ",
		" ▄  █▄  ",
		"█    ▀█ ",
	},
	'Y': {
		"█    █▄ ",
		" ▄  █▀  ",
		"  ▀▀▀   ",
		"   █    ",
		"   █    ",
	},
	'Z': {
		"███████ ",
		"     ▄▀ ",
		"   ▄▀   ",
		" ▄▀     ",
		"███████ ",
	},
	'0': {
		" ▄███▄  ",
		"█    █  ",
		"█   ▀█  ",
		"█    █  ",
		" ▀███▀  ",
	},
	'1': {
		"  ▄█   ",
		" █ █   ",
		"   █   ",
		"   █   ",
		"███████",
	},
	'2': {
		" ▄███▄  ",
		"█    █  ",
		"   ▄▀   ",
		" ▄▀     ",
		"███████ ",
	},
	'3': {
		"█████▄  ",
		"     █  ",
		" ▄███▀  ",
		"     █  ",
		"█████▀  ",
	},
	'4': {
		"█    █  ",
		"█    █  ",
		"███████ ",
		"     █  ",
		"     █  ",
	},
	'5': {
		"███████ ",
		"█       ",
		"█████▄  ",
		"     █  ",
		"█████▀  ",
	},
	'6': {
		" ▄███▄  ",
		"█       ",
		"█████▄  ",
		"█    █  ",
		" ▀███▀  ",
	},
	'7': {
		"███████ ",
		"    ▄▀  ",
		"   ▄▀   ",
		"  ▄▀    ",
		" ▄▀     ",
	},
	'8': {
		" ▄███▄  ",
		"█    █  ",
		" ▀███▀  ",
		"█    █  ",
		" ▀███▀  ",
	},
	'9': {
		" ▄███▄  ",
		"█    █  ",
		" ▀████▄ ",
		"     █  ",
		" ▀███▀  ",
	},
	' ': {
		"        ",
		"        ",
		"        ",
		"        ",
		"        ",
	},
	'!': {
		"█",
		"█",
		"█",
		" ",
		"█",
	},
	'.': {
		" ",
		" ",
		" ",
		" ",
		"█",
	},
	',': {
		" ",
		" ",
		" ",
		"▄",
		"█",
	},
	'?': {
		" ▄███▄  ",
		"█    █  ",
		"  ▄▀▀   ",
		"        ",
		"  ▄▀    ",
	},
	':': {
		" ",
		"▄",
		" ",
		"▄",
		" ",
	},
	';': {
		" ",
		"▄",
		" ",
		"▄",
		"█",
	},
	'\'': {
		"█",
		"█",
		" ",
		" ",
		" ",
	},
	'-': {
		"       ",
		"       ",
		"███████",
		"       ",
		"       ",
	},
	'/': {
		"      █",
		"    ▄▀ ",
		"  ▄▀   ",
		"▄▀     ",
		"█      ",
	},
	'&': {
		" ▄▀▄    ",
		"█  █    ",
		" ▄▀▀▄▄  ",
		"█  ▄▀ █ ",
		" ▀▀▀  ▀ ",
	},
	'%': {
		"██  ▄▀",
		"██ ▄▀ ",
		"  ▄▀  ",
		" ▄▀ ██",
		"▄▀  ██",
	},
	'(': {
		"  ▄█",
		" █  ",
		" █  ",
		" █  ",
		"  ▀█",
	},
	')': {
		"█▄  ",
		"  █ ",
		"  █ ",
		"  █ ",
		"█▀  ",
	},
}

func init() { register(DefaultName, NewBlock(blockGlyphs)) }
