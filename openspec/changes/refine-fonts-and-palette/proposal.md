# Refine fonts & palette into a focused, shouty alert toolkit

> Precursor: see [`draft.md`](./draft.md) for the directional exploration this
> proposal formalizes. Design decisions live in [`design.md`](./design.md).

## Why

plaqq's job is **big, shouty alerts dropped into a pane of a multi-pane terminal**
— to grab attention and remind you what a tab/session is working on. The win is
*legibility at a glance* and *low decision cost*, not coverage.

The recent styling work optimized for breadth instead: 10 colour presets and 12
fonts (3 block + 9 `go-figure` FIGlet ASCII faces). Most of the FIGlet faces
(`slant`, `mini`, `cyberlarge`, `larry3d`…) are thin or ornamental — the opposite
of "shouty" — and several presets are fixed hexes that wash out on light
terminals. A few great bold options beat a long menu of mediocre ones, and the
`go-figure` runtime dependency exists only to serve faces we don't want.

## What Changes

- **Fonts: cut to exactly 3 bespoke block fonts**, a size/weight gradient — all
  genuine "shouts", visually distinct, each refined by hand:
  - `block` — clean medium block (default)
  - `heavy` — solid filled block (loudest)
  - `compact` — 3-row half-block (dense; for short/narrow panes)
- **Remove the `go-figure` runtime dependency** and all 9 FIGlet ASCII faces.
  Fonts become static, committed glyph maps with zero runtime deps. *(BREAKING)*
- **Charset: uppercase-only** — `A–Z`, `0–9`, and the punctuation set
  `! ? . , : ; ' - / & % ( )` plus space. Equal-width rows per glyph.
- **Palette: replace the 10 presets with 5 semantic adaptive colours** —
  `alert` (red) / `warn` (amber) / `info` (teal, default) / `ok` (green) /
  `focus` (magenta), each a Light/Dark `AdaptiveColor`. Raw hex + ANSI index are
  retained as an escape hatch. *(BREAKING: old preset names removed.)*
- **Unknown font/colour names become a hard error** (non-zero exit) with a
  message listing the valid options — for both `--font`/`--color` flags and
  config-file values. The `plaqq config` picker is exempt: it tolerates an
  invalid stored value so it stays usable as the repair path.
- **Dev tooling: `cmd/fontgallery`** — a headless, no-TTY stdout previewer
  (sample render of every font + full-charset dump per font) plus golden snapshot
  tests, so glyphs can be iterated and regressions caught in CI.

## Impact

- **Affected specs:** new `fonts`, `colors`, `config` capability requirements.
- **Affected code:** `internal/font/` (drop `figlet.go`; rework `block.go`,
  `compact.go`, `font.go`, registry to the 3-font set + uppercase charset),
  `internal/cmd/styles.go` (semantic adaptive palette), `internal/cmd/root.go`
  (hard-error resolution), `internal/cmd/config.go` (picker tolerance, updated
  options), `go.mod` (remove `go-figure`), new `cmd/fontgallery/`, README +
  AGENTS docs.
- **Breaking for users:** configs/scripts referencing a removed font
  (`standard`, `slant`, `banner`, `big`, `small`, `doom`, `larry3d`, `mini`,
  `cyberlarge`) or a removed preset (`coral`, `lime`, `azure`, `violet`,
  `magenta`*, `rose`, `crimson`, `slate`; `amber`/`teal` survive as semantics)
  will now error and must be updated. Acceptable pre-1.0; documented in the
  changelog. *(`magenta` is reborn as the `focus` semantic.)*
