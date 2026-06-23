package font

import (
	"sort"
	"strings"
)

// Font renders a single line of text into a stack of display rows. Different
// fonts have different row heights and per-row widths, so callers must center
// each rendered line as a block rather than assuming a fixed geometry.
type Font interface {
	// Render returns the display rows for one (already wrapped) line of text.
	Render(line string) []string
	// Width returns the rendered column width of s, used for wrapping.
	Width(s string) int
}

// DefaultName is the registry key for the built-in chunky block font.
const DefaultName = "block"

// Charset is the curated uppercase glyph coverage shared by the gallery and the
// font tests: A-Z, 0-9, ! ? . , : ; ' - / & % ( ) plus space.
const Charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!?.,:;'-/&%() @#$^*_=+<>`~[]{}"

var registry = map[string]Font{}

// register adds a font under a lower-cased name. It is called from init
// functions across the package.
func register(name string, f Font) {
	registry[strings.ToLower(name)] = f
}

// Get returns the font registered under name, falling back to the default block
// font when name is empty or unknown.
func Get(name string) Font {
	if name == "" {
		name = DefaultName
	}
	if f, ok := registry[strings.ToLower(name)]; ok {
		return f
	}
	return registry[DefaultName]
}

// Has reports whether a font is registered under name (case-insensitive).
func Has(name string) bool {
	_, ok := registry[strings.ToLower(name)]
	return ok
}

// Names returns font names in their curated display order, with any unexpected
// extras sorted after the built-ins.
func Names() []string {
	preferred := []string{DefaultName, "heavy", "compact", "wide"}
	names := make([]string, 0, len(registry))
	seen := make(map[string]struct{}, len(preferred))
	for _, name := range preferred {
		if _, ok := registry[name]; ok {
			names = append(names, name)
			seen[name] = struct{}{}
		}
	}

	var extras []string
	for name := range registry {
		if _, ok := seen[name]; ok {
			continue
		}
		extras = append(extras, name)
	}
	sort.Strings(extras)
	return append(names, extras...)
}

// Wrap splits text into lines whose rendered width in font f does not exceed
// maxWidth. It greedily packs whole words and never splits a single word that
// is itself wider than maxWidth.
func Wrap(f Font, text string, maxWidth int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	var current []string

	for _, word := range words {
		joined := word
		if len(current) > 0 {
			joined = strings.Join(current, " ") + " " + word
		}
		if f.Width(joined) <= maxWidth {
			current = append(current, word)
		} else {
			if len(current) > 0 {
				lines = append(lines, strings.Join(current, " "))
				current = []string{word}
			} else {
				lines = append(lines, word)
				current = nil
			}
		}
	}
	if len(current) > 0 {
		lines = append(lines, strings.Join(current, " "))
	}
	return lines
}
