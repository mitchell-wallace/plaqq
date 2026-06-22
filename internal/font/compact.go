package font

// compactGlyphs is the dense companion face ("compact"): a 3-row half-block
// (▀▄█) alphabet that fits short or narrow panes where the taller block/heavy
// faces would overflow. Each glyph's rows share one width (the rendering
// pipeline relies on it; enforced by TestBlockRowWidths).
//
// Provenance / licence: transcribed from TOIlet's `pagga.tlf` font by Sam
// Hocevar (2010), distributed under the Do What The Fuck You Want To Public
// License, Version 2 (WTFPL). The WTFPL is an unconditional public-licence that
// is compatible with plaqq's MIT licence — verified against the font's own
// header and the TOIlet COPYING file at
// https://github.com/cacalabs/toilet (font: fonts/pagga.tlf). Transcription
// rule: pagga's leading ░ pad column is dropped (inter-letter spacing is added
// by the renderer) and its remaining ░ shade cells are mapped to spaces; the
// ▀▄█ block cells are kept verbatim. pagga's `/` and `\` diagonals on `0` and
// `Q` are preserved so 0 stays distinct from O.
var compactGlyphs = map[rune][]string{
	'A':  {"█▀█", "█▀█", "▀ ▀"},
	'B':  {"█▀▄", "█▀▄", "▀▀ "},
	'C':  {"█▀▀", "█  ", "▀▀▀"},
	'D':  {"█▀▄", "█ █", "▀▀ "},
	'E':  {"█▀▀", "█▀▀", "▀▀▀"},
	'F':  {"█▀▀", "█▀▀", "▀  "},
	'G':  {"█▀▀", "█ █", "▀▀▀"},
	'H':  {"█ █", "█▀█", "▀ ▀"},
	'I':  {"▀█▀", " █ ", "▀▀▀"},
	'J':  {"▀▀█", "  █", "▀▀ "},
	'K':  {"█ █", "█▀▄", "▀ ▀"},
	'L':  {"█  ", "█  ", "▀▀▀"},
	'M':  {"█▄█", "█ █", "▀ ▀"},
	'N':  {"█▀█", "█ █", "▀ ▀"},
	'O':  {"█▀█", "█ █", "▀▀▀"},
	'P':  {"█▀█", "█▀▀", "▀  "},
	'Q':  {"▄▀▄", "█\\█", " ▀\\"},
	'R':  {"█▀▄", "█▀▄", "▀ ▀"},
	'S':  {"█▀▀", "▀▀█", "▀▀▀"},
	'T':  {"▀█▀", " █ ", " ▀ "},
	'U':  {"█ █", "█ █", "▀▀▀"},
	'V':  {"█ █", "▀▄▀", " ▀ "},
	'W':  {"█ █", "█▄█", "▀ ▀"},
	'X':  {"█ █", "▄▀▄", "▀ ▀"},
	'Y':  {"█ █", " █ ", " ▀ "},
	'Z':  {"▀▀█", "▄▀ ", "▀▀▀"},
	'0':  {"▄▀▄", "█/█", " ▀ "},
	'1':  {"▀█ ", " █ ", "▀▀▀"},
	'2':  {"▀▀▄", "▄▀ ", "▀▀▀"},
	'3':  {"▀▀█", " ▀▄", "▀▀ "},
	'4':  {"█ █", " ▀█", "  ▀"},
	'5':  {"█▀▀", "▀▀▄", "▀▀ "},
	'6':  {"▄▀▀", "█▀▄", " ▀ "},
	'7':  {"▀▀█", "▄▀ ", "▀  "},
	'8':  {"▄▀▄", "▄▀▄", " ▀ "},
	'9':  {"▄▀▄", " ▀█", "▀▀ "},
	'!':  {"█", "▀", "▀"},
	'?':  {"▀▀█", " ▀ ", " ▀ "},
	'.':  {"  ", "  ", "▀ "},
	',':  {"   ", "   ", "▄▀ "},
	':':  {"   ", " ▀ ", " ▀ "},
	';':  {"   ", " ▀ ", "▄▀ "},
	'\'': {"▀", " ", " "},
	'-':  {"   ", "▄▄▄", "   "},
	'/':  {"  █", "▄▀ ", "▀  "},
	'&':  {"▄▀ ", "▄█▀", " ▀▀"},
	'%':  {"▀ █", "▄▀ ", "▀ ▀"},
	'(':  {"▄▀ ", "█  ", " ▀ "},
	')':  {"▀▄ ", " █ ", "▀  "},
	' ':  {" ", " ", " "},
}

func init() { register("compact", NewBlock(compactGlyphs)) }
