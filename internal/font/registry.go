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

// DefaultName is the registry key for the default font.
const DefaultName = "compact"

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
	preferred := []string{DefaultName, "terminal", "block", "heavy", "wide"}
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
// maxWidth. Explicit line breaks ('\n') are honored as segment boundaries:
// each '\n' starts a new segment that is word-wrapped independently, and the
// segments are concatenated. Carriage returns ('\r\n' and bare '\r') are
// normalized to '\n'. An empty input yields no lines.
func Wrap(f Font, text string, maxWidth int) []string {
	var out []string
	for _, seg := range WrapSegments(f, text, maxWidth) {
		out = append(out, seg...)
	}
	return out
}

// WrapSegments splits text into segments on explicit line breaks ('\n') and
// word-wraps each segment independently. The outer slice has one entry per
// '\n'-delimited segment (so consecutive newlines produce empty inner slices),
// letting callers apply per-segment gap rules. Carriage returns are normalized
// to '\n' first. An entirely empty/blank input returns nil.
func WrapSegments(f Font, text string, maxWidth int) [][]string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	segments := strings.Split(text, "\n")
	out := make([][]string, 0, len(segments))
	for _, seg := range segments {
		out = append(out, wrapSegment(f, seg, maxWidth))
	}
	return out
}

// wrapSegment greedily packs whole words of a single (already '\n'-stripped)
// segment into lines whose rendered width does not exceed maxWidth, splitting
// long words with hyphens if they exceed maxWidth on their own. It uses
// strings.Fields so intra-segment whitespace collapses, matching the legacy
// single-line wrapping behavior.
func wrapSegment(f Font, text string, maxWidth int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	var current []string

	for len(words) > 0 {
		word := words[0]
		words = words[1:]

		joined := word
		if len(current) > 0 {
			joined = strings.Join(current, " ") + " " + word
		}

		if f.Width(joined) <= maxWidth {
			current = append(current, word)
		} else {
			if len(current) > 0 {
				lines = append(lines, strings.Join(current, " "))
				current = nil
				// Re-process this word on a new, empty line
				words = append([]string{word}, words...)
			} else {
				// The word is too wide for an empty line. Split it.
				runes := []rune(word)
				if len(runes) <= 1 {
					// Cannot split. Just put it on the line.
					lines = append(lines, word)
				} else {
					k := 1
					for i := 2; i < len(runes); i++ {
						if f.Width(string(runes[:i])+"-") <= maxWidth {
							k = i
						} else {
							break
						}
					}
					// Add prefix with hyphen to lines
					lines = append(lines, string(runes[:k])+"-")
					// Prepend the suffix to words to be processed next
					suffix := string(runes[k:])
					words = append([]string{suffix}, words...)
				}
			}
		}
	}
	if len(current) > 0 {
		lines = append(lines, strings.Join(current, " "))
	}
	return lines
}
