package cmd

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// colorPresets is the semantic adaptive palette: exactly five colours named for
// *why* you'd drop the notice, each a Light/Dark AdaptiveColor pair tuned for
// legibility on both light and dark terminals. The old single-hex presets washed
// out on one background or the other; adaptive pairs stay readable on both.
//
//	info  (teal, default)    — neutral
//	alert (red)              — failure / danger
//	warn  (amber)            — caution
//	ok    (green)            — success / done
//	focus (magenta)          — "this is what you're on"
//
// The Light variant is shown on light backgrounds (kept deeper for contrast);
// the Dark variant is shown on dark backgrounds (kept brighter for contrast).
// Raw hex (#rrggbb) and ANSI index (0-255) remain supported via parseColor as a
// power-user escape hatch outside this map.
var colorPresets = map[string]lipgloss.TerminalColor{
	"alert": lipgloss.AdaptiveColor{Light: "#d70000", Dark: "#ff6b6b"},
	"warn":  lipgloss.AdaptiveColor{Light: "#b8860b", Dark: "#ffc14d"},
	"info":  lipgloss.AdaptiveColor{Light: "#00d7af", Dark: "#00f5d4"},
	"ok":    lipgloss.AdaptiveColor{Light: "#1a7f37", Dark: "#3fb950"},
	"focus": lipgloss.AdaptiveColor{Light: "#a21caf", Dark: "#e879f9"},
}

// presetOrder is the stable display order for presets in help text and the
// interactive picker. Call sites (root.go flag help + parseColor error message)
// interpolate this slice, so the symbol name is load-bearing — do not rename.
var presetOrder = []string{
	"alert", "warn", "info", "ok", "focus",
}

// presetColor looks up a color preset by name, case-insensitively.
func presetColor(name string) (lipgloss.TerminalColor, bool) {
	c, ok := colorPresets[strings.ToLower(strings.TrimSpace(name))]
	return c, ok
}
