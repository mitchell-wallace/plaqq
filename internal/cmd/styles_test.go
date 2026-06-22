package cmd

import (
	"reflect"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestPresetOrderIsSemanticPalette asserts the palette exposes exactly the five
// semantic names in the stable display order.
func TestPresetOrderIsSemanticPalette(t *testing.T) {
	want := []string{"alert", "warn", "info", "ok", "focus"}
	if !reflect.DeepEqual(presetOrder, want) {
		t.Errorf("presetOrder = %v; want %v", presetOrder, want)
	}
}

// TestPresetColorAdaptive asserts every preset resolves to an AdaptiveColor with
// distinct, non-empty Light and Dark variants (the whole point of the redesign:
// the old single-hex presets washed out on one background).
func TestPresetColorAdaptive(t *testing.T) {
	for _, name := range presetOrder {
		c, ok := presetColor(name)
		if !ok {
			t.Errorf("presetColor(%q) not found", name)
			continue
		}
		ac, ok := c.(lipgloss.AdaptiveColor)
		if !ok {
			t.Errorf("presetColor(%q) = %T; want lipgloss.AdaptiveColor", name, c)
			continue
		}
		if ac.Light == "" || ac.Dark == "" {
			t.Errorf("presetColor(%q) has empty variant: %+v", name, ac)
		}
		if ac.Light == ac.Dark {
			t.Errorf("presetColor(%q) Light == Dark (%q); variants must differ", name, ac.Light)
		}
	}
}

// TestPresetColorNoOldPresets asserts the removed preset names no longer resolve.
func TestPresetColorNoOldPresets(t *testing.T) {
	for _, old := range []string{"teal", "coral", "amber", "lime", "azure", "violet", "magenta", "rose", "crimson", "slate"} {
		if _, ok := presetColor(old); ok {
			t.Errorf("presetColor(%q) unexpectedly resolved; old presets must be gone", old)
		}
	}
}

// TestPresetColorCaseInsensitive asserts lookups ignore case and surrounding
// whitespace (the documented contract that parseColor and the config picker rely
// on).
func TestPresetColorCaseInsensitive(t *testing.T) {
	for _, in := range []string{"INFO", "Info", " info ", "\tOk\n"} {
		if _, ok := presetColor(in); !ok {
			t.Errorf("presetColor(%q) unexpectedly not found", in)
		}
	}
}

// TestInfoIsCurrentAdaptiveTeal asserts the default info preset stays the
// pre-existing adaptive teal so the default render is unchanged.
func TestInfoIsCurrentAdaptiveTeal(t *testing.T) {
	c, ok := presetColor("info")
	if !ok {
		t.Fatal("info preset missing")
	}
	ac, ok := c.(lipgloss.AdaptiveColor)
	if !ok {
		t.Fatalf("info preset = %T; want AdaptiveColor", c)
	}
	want := lipgloss.AdaptiveColor{Light: "#00d7af", Dark: "#00f5d4"}
	if ac != want {
		t.Errorf("info preset = %+v; want %+v (current adaptive teal)", ac, want)
	}
}
