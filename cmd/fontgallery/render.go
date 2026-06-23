package main

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// The gallery renders font output to PNG by treating every terminal cell as a
// small bitmap and painting the exact sub-cell geometry of each Unicode block
// element. This is deliberately font-independent: it shows the *true* shape a
// glyph row describes (which half/quadrant of each cell is filled) rather than
// however a particular monospace font happens to draw ▀▄█. That makes it a
// faithful surface for judging legibility and spotting defects like a missing
// centre connection or an asymmetric bowl.

// fill is a rectangle inside a single cell, in fractional coordinates where
// (0,0) is the top-left corner and (1,1) the bottom-right.
type fill struct{ x0, y0, x1, y1 float64 }

var full = fill{0, 0, 1, 1}

// cellFills maps each block-element rune used by the fonts to the foreground
// rectangles that compose it. Halves, quadrants and the two- quadrant combining
// blocks are covered. Runes absent here (other than space and the diagonals,
// handled separately) fall back to a full block so unexpected data stays
// visible during authoring.
var cellFills = map[rune][]fill{
	'█': {full},
	'▀': {{0, 0, 1, 0.5}},
	'▄': {{0, 0.5, 1, 1}},
	'▌': {{0, 0, 0.5, 1}},
	'▐': {{0.5, 0, 1, 1}},
	'▖': {{0, 0.5, 0.5, 1}},
	'▗': {{0.5, 0.5, 1, 1}},
	'▘': {{0, 0, 0.5, 0.5}},
	'▝': {{0.5, 0, 1, 0.5}},
	'▙': {{0, 0, 0.5, 1}, {0.5, 0.5, 1, 1}},   // left half + bottom-right
	'▟': {{0.5, 0, 1, 1}, {0, 0.5, 0.5, 1}},   // right half + bottom-left
	'▛': {{0, 0, 1, 0.5}, {0, 0.5, 0.5, 1}},   // top half + bottom-left
	'▜': {{0, 0, 1, 0.5}, {0.5, 0.5, 1, 1}},   // top half + bottom-right
	'▚': {{0, 0, 0.5, 0.5}, {0.5, 0.5, 1, 1}}, // top-left + bottom-right
	'▞': {{0.5, 0, 1, 0.5}, {0, 0.5, 0.5, 1}}, // top-right + bottom-left
	'┌': {{0.4, 0.4, 1, 0.6}, {0.4, 0.4, 0.6, 1}},
	'│': {{0.4, 0, 0.6, 1}},
	'└': {{0.4, 0.4, 1, 0.6}, {0.4, 0, 0.6, 0.6}},
	'╋': {{0, 0.4, 1, 0.6}, {0.4, 0, 0.6, 1}},
	' ': nil,
}

// renderRowsPNG paints the given display rows into an RGBA image, one cell per
// rune. Rows may have differing lengths; the image is widened to the longest
// and shorter rows are left-aligned (trailing background).
func renderRowsPNG(rows []string, cellW, cellH int, fg, bg color.Color) *image.RGBA {
	maxCols := 1
	for _, row := range rows {
		if n := len([]rune(row)); n > maxCols {
			maxCols = n
		}
	}
	h := len(rows) * cellH
	if h == 0 {
		h = cellH
	}
	img := image.NewRGBA(image.Rect(0, 0, maxCols*cellW, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	for ry, row := range rows {
		oy := ry * cellH
		for cx, r := range []rune(row) {
			drawCell(img, r, cx*cellW, oy, cellW, cellH, fg)
		}
	}
	return img
}

func drawCell(img *image.RGBA, r rune, ox, oy, cw, ch int, fg color.Color) {
	switch r {
	case ' ', 0:
		return
	case '/', '\\':
		drawDiagonal(img, r, ox, oy, cw, ch, fg)
		return
	}
	fills, ok := cellFills[r]
	if !ok {
		fills = []fill{full} // unknown rune: paint solid so it is obvious
	}
	for _, f := range fills {
		rect := image.Rect(
			ox+int(f.x0*float64(cw)),
			oy+int(f.y0*float64(ch)),
			ox+int(f.x1*float64(cw)),
			oy+int(f.y1*float64(ch)),
		)
		draw.Draw(img, rect, &image.Uniform{fg}, image.Point{}, draw.Src)
	}
}

// drawDiagonal approximates the ASCII '/' and '\' strokes (used by the compact
// font) as a thick diagonal band so they read as strokes rather than full
// blocks.
func drawDiagonal(img *image.RGBA, r rune, ox, oy, cw, ch int, fg color.Color) {
	const halfThick = 0.13
	for py := 0; py < ch; py++ {
		fy := (float64(py) + 0.5) / float64(ch)
		for px := 0; px < cw; px++ {
			fx := (float64(px) + 0.5) / float64(cw)
			var d float64
			if r == '\\' {
				d = math.Abs(fy - fx)
			} else {
				d = math.Abs(fy - (1 - fx))
			}
			if d < halfThick {
				img.Set(ox+px, oy+py, fg)
			}
		}
	}
}
