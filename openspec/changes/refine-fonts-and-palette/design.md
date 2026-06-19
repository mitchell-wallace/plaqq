# Design: refine-fonts-and-palette

Technical decisions behind the proposal. Captures the "why this way", so the
tasks and specs can stay terse.

## Rendering model: static glyph maps, zero runtime deps

**Decision:** ship hand/offline-authored block-glyph maps
(`map[rune][]string`); the binary does no font rasterization and depends on no
font library at runtime. Drop `go-figure`.

The real axis isn't "static vs rasterization" — it's two separable questions:

| | Authoring (how shapes are made) | Runtime (what the binary does) |
|---|---|---|
| hand-author | type `█▄▀` into source | read map, zero deps |
| offline-generate | a `go:generate`/dev tool transcribes a source → commits maps | read map, zero deps |
| runtime-rasterize | bundle a TTF/BDF | rasterize per run, needs `x/image/font`+freetype |

Authoring may use offline generation; **runtime stays static-map either way.**
The *only* thing that would force runtime rasterization is a font whose output
size depends on the live terminal width (a pane-adaptive "huge" face) — deferred
to a possible v2. At our sizes (3–6 rows) hand-tuned glyphs also read better than
downsampled real text, so deferring costs little.

## The 3 fonts (size/weight gradient)

Distinct in *function*, not just decoration, so the choice maps to pane reality:

- **`block`** — clean medium block, ~5 rows. The default workhorse.
- **`heavy`** — solid filled block, the loudest face for the biggest shout.
- **`compact`** — 3-row half-block (`▀▄█`), dense; fits short or narrow panes
  where the taller faces would wrap.

Names reuse the current `block`/`heavy`/`compact` keys (the glyph *shapes* are
re-authored, but the identifiers are stable).

## Charset: uppercase-only

`A–Z`, `0–9`, and `! ? . , : ; ' - / & % ( )` + space (~49 glyphs/font). Input is
already upper-cased today; uppercase matches the shouty aesthetic literally and
roughly halves the authoring/tuning work. Per-glyph rows must share one width
(existing `TestBlockRowWidths` invariant, extended to all three fonts).

## Authoring pipeline & provenance

Build the **tooling before the glyphs**:

- `cmd/fontgallery` — a standalone program (not imported by the CLI) that prints
  to ordinary **stdout** (no TTY, runs here and in CI): every font rendering a
  fixed sample (`SHIP IT! BUILD FAILED 0123`), then a full-charset dump per font
  so bad letters are obvious while editing.
- **Golden snapshot tests** — freeze each font's full-charset render to
  `testdata/`; glyph edits show up as readable diffs, regressions fail CI.

Then author each face, transcribing from a clean source where one fits and
hand-tuning the rest:

- `compact` (half-block) ← TOIlet **`pagga`** `.tlf` (already `▀▄█`).
- `block` (medium) ← npm **`cfonts`** `block`/`simpleBlock` JSON maps as a
  starting point.
- `heavy` (solid) ← hand-authored / thickened.

**Provenance:** record source + licence in a header comment and/or `NOTICE`.
**Verify each source's actual licence before copying any glyph** (cfonts is MIT;
TOIlet/FIGlet fonts vary) — do not transcribe on the basis of remembered
licensing.

## Colours: 5 semantic adaptive

`lipgloss.AdaptiveColor` Light/Dark pairs, named for *why you'd drop the alert*:

| name | hue | role |
|------|-----|------|
| `alert` | red | failure / danger |
| `warn` | amber | caution |
| `info` | teal | neutral (**default**, current adaptive teal) |
| `ok` | green | success / done |
| `focus` | magenta | "this is what you're on" |

Each pair tuned for contrast on both light and dark backgrounds (the old presets
were mostly single hexes that wash out). Raw **hex** (`#rrggbb`) and **ANSI
index** (`0–255`) stay supported as a power-user escape hatch.

## Migration: hard error on unknown names

Resolving an unknown **named** font or colour (from a `--font`/`--color` flag or
a config-file value) exits non-zero with a message listing the valid names. Note
`font.Get` currently *silently falls back* to `block`; the cmd layer must
validate (`font.Has` / preset lookup) and reject **before** that fallback is
reached. Hex/ANSI colours bypass the named-preset check as today.

**Exception — the `plaqq config` picker tolerates an invalid stored value.** It's
the repair path, so it must still open: an unknown stored font/colour is shown as
the default selection (not an error), letting the user pick a valid one and save.

Pre-1.0 with a tiny userbase, so no alias layer — the removed names are simply
gone and documented in the changelog.

## Out of scope (v2 candidates)

- Runtime rasterization / a pane-width auto-scaling "huge" font.
- Arbitrary user-supplied font loading; full FIGlet support.
- Lowercase glyphs.
