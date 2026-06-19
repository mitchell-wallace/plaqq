# Draft: refine fonts & palette into a focused, shouty alert toolkit

> Status: directional draft only. No proposal / specs / tasks yet — capturing intent
> and approach options so we can choose a direction before formalizing.

## Why

The recent styling work added breadth: 10 colour presets and 12 fonts (3 block +
9 go-figure FIGlet ASCII faces). It works, but it's unfocused for what plaqq is
actually *for*.

**plaqq's real job:** big, shouty alerts dropped into a pane of a multi-pane
terminal window — to grab human attention and remind you what this tab / session /
set of panes is working on. The win is *legibility at a glance* and *low decision
cost*, not coverage.

Most of the FIGlet ASCII faces (slant, standard, larry3d, mini, cyberlarge…) are
thin, ornamental, or hard to read across a pane — the opposite of "shouty." A few
great bold options beat a long menu of mediocre ones.

## Direction

- **Fonts:** cut to **3 refined, bespoke block-style fonts** — all big, bold, and
  unmistakably legible across a pane. Distinct from each other (e.g. a clean
  medium block, a heavy/solid block, and a 3D/shadow or outline block), but every
  one a genuine "shout."
- **Colours:** a **small, curated set of adaptive colours** (Light/Dark pairs)
  tuned to stay legible on both light and dark terminals. Ideally named for the
  alert use case rather than the hue.
- **Bias to zero/low runtime deps:** bake glyphs in as static data we fully
  control and can refine by hand. Treat external font libraries as *sources to
  transcribe from*, not runtime dependencies.

## Fonts — candidate sources & techniques

### A. Transcribe glyphs from other-language terminal font libraries
The cleanest path: lift well-designed glyph shapes and re-encode them into our
`map[rune][]string` block format (see `internal/font/block.go`).

- **npm `cfonts`** (Dominik Wilkowski) — *strongest lead.* Purpose-built for big
  bold console headings. Fonts like `block`, `simpleBlock`, `huge`, `3d`, `shade`,
  `grid` are exactly this aesthetic. Glyphs ship as JSON character maps
  (`fonts/*.json`: per-letter arrays of strings) → near-direct transcription.
  MIT-licensed.
- **TOIlet `.tlf` fonts**, especially **`pagga`** — already our exact aesthetic
  (Unicode `▀ ▄ █` half-blocks). Transcribe directly. Also `future`, `mono*`.
- **Python `pyfiglet` / `art`** — large bundled font collections to mine shapes
  from (mostly `.flf`; useful for the medium/clean block face).
- Licensing: check each font/glyph set's licence before copying (cfonts MIT;
  FIGlet/TOIlet fonts vary). Record provenance for whatever we transcribe.

### B. Rebuild go-figure / FIGlet shapes in block & box-drawing characters
Take a FIGlet letterform and deliberately *thicken / redraw* its strokes with
block and box-drawing runes (`█ ▉ ▛ ▜ ▙ ▟ ━ ┃ ┏ ┓ ┗ ┛`) to produce a bold
variant. Our current `heavy` (banner3 with `#`→`█`) is a crude, automatic version
of this — do it intentionally, per glyph, with corner/edge cleanup.

### C. Rasterize a real font → block characters (highest fidelity, later)
Render a bold bitmap/TTF glyph to a small pixel grid, then map cells to the
best-fit block symbol — half-block (1×2), quadrant (2×2), sextant, or braille
(2×4). This is the technique image-to-terminal tools use (`chafa`, `viu`,
`catimg`). Pros: arbitrary size, genuinely bold, scalable; could bake a glyph
atlas from one good bold bitmap font (e.g. a BDF). Cons: real machinery — likely a
v2 once the static set proves the shape of the feature.

## Colours — focused adaptive palette

- Replace the 10 mostly-*fixed* presets with **~4–6 adaptive colours**
  (`lipgloss.AdaptiveColor` Light/Dark pairs) picked for high contrast on both
  backgrounds — the current set is mostly single hex values that can wash out on a
  light terminal.
- Lean **semantic** for alerts, e.g. `alert` (red), `warn` (amber),
  `info`/default (teal), `ok` (green), `focus` (magenta). Names that map to *why
  you'd drop the alert*, not just the colour.
- Keep a raw hex / ANSI escape hatch for power users.

## Likely scope decisions (to resolve when we formalize)

- **Drop the FIGlet ASCII family + the `go-figure` runtime dependency?** Leaning
  yes — use `go-figure`/cfonts/etc. only as offline transcription sources, ship
  zero-dep static glyph maps.
- **Which 3 fonts exactly**, and should one be a size-adaptive "huge" that scales
  with pane width?
- **Static glyph maps (v1)** vs the rasterization engine (defer).
- **Fate of the just-added `compact` / `heavy` fonts and 10 presets** —
  deprecate, replace, or fold into the new curated set (and how loudly to handle
  configs that reference a removed name).

## Non-goals (for now)

- Full FIGlet support / arbitrary user-supplied font loading.
- A per-glyph rasterization engine (revisit as v2 if static maps fall short).
- Broad configurability — deliberately *few, curated* options.

## Open questions

- The exact 3 font styles, and whether one auto-scales to pane width.
- Should font choice auto-fallback when a pane is too narrow (e.g. degrade to a
  shorter face) rather than wrap?
- Licensing/provenance for any transcribed glyph sets.
- Migration: do we keep old preset/font names as aliases for one release, or break
  cleanly given the tool is still pre-1.0?
