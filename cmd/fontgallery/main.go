// Command fontgallery renders every registered font to inspectable artifacts:
// a PNG per font (faithful cell-painted geometry), the raw text/ANSI rows, and
// a manifest.json tying them together. It exists so humans and agents can
// actually *see* the glyphs while iterating on the fonts, rather than judging
// block characters in a transcript. Run via `just visual`.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchell-wallace/plaqq/internal/font"
)

// sampleLines are rendered (in order, separated by a blank row) into each
// font's artifact. They cover the full charset in readable chunks plus a
// pangram so letters can be judged both in isolation and in words.
var sampleLines = []string{
	"ABCDEFGHIJKLM",
	"NOPQRSTUVWXYZ",
	"0123456789",
	"!?.,:;'-/&%()",
	"THE QUICK BROWN",
	"FOX JUMPS OVER",
	"THE LAZY DOG",
}

// manifest documents an artifact run for downstream tooling and agents.
type manifest struct {
	CellWidth  int          `json:"cell_width"`
	CellHeight int          `json:"cell_height"`
	Foreground string       `json:"foreground"`
	Background string       `json:"background"`
	Samples    []string     `json:"samples"`
	Fonts      []fontRecord `json:"fonts"`
}

type fontRecord struct {
	Name    string `json:"name"`
	Default bool   `json:"default"`
	PNG     string `json:"png"`
	ANSI    string `json:"ansi"`
	Height  int    `json:"height"`
}

func main() {
	out := flag.String("out", "artifacts/visual", "output directory for PNG/ANSI/manifest artifacts")
	cellW := flag.Int("cell-w", 16, "pixel width of one terminal cell")
	cellH := flag.Int("cell-h", 28, "pixel height of one terminal cell")
	stdout := flag.Bool("stdout", false, "also print rendered samples to stdout")
	flag.Parse()

	fg := color.RGBA{0x00, 0xf5, 0xd4, 0xff} // adaptive "info" teal (dark variant)
	bg := color.RGBA{0x1b, 0x1b, 0x1b, 0xff}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "fontgallery:", err)
		os.Exit(1)
	}

	m := manifest{
		CellWidth:  *cellW,
		CellHeight: *cellH,
		Foreground: "#00f5d4",
		Background: "#1b1b1b",
		Samples:    sampleLines,
	}

	for _, name := range font.Names() {
		f := font.Get(name)

		// Stack every sample's rendered rows with one blank separator row
		// between groups so glyphs and the pangram share one image.
		var rows []string
		for i, line := range sampleLines {
			if i > 0 {
				rows = append(rows, "")
			}
			rows = append(rows, f.Render(line)...)
		}

		pngName := name + ".png"
		ansiName := name + ".ansi"

		img := renderRowsPNG(rows, *cellW, *cellH, fg, bg)
		if err := writePNG(filepath.Join(*out, pngName), img); err != nil {
			fmt.Fprintln(os.Stderr, "fontgallery:", err)
			os.Exit(1)
		}
		ansi := strings.Join(rows, "\n") + "\n"
		if err := os.WriteFile(filepath.Join(*out, ansiName), []byte(ansi), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "fontgallery:", err)
			os.Exit(1)
		}

		m.Fonts = append(m.Fonts, fontRecord{
			Name:    name,
			Default: name == font.DefaultName,
			PNG:     pngName,
			ANSI:    ansiName,
			Height:  len(f.Render("A")),
		})

		if *stdout {
			fmt.Printf("=== %s ===\n%s\n", name, ansi)
		}
	}

	mf, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "fontgallery:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(filepath.Join(*out, "manifest.json"), append(mf, '\n'), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "fontgallery:", err)
		os.Exit(1)
	}

	fmt.Printf("wrote %d fonts to %s\n", len(m.Fonts), *out)
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	return enc.Encode(f, img)
}
