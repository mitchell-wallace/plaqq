# Tasks: refine-fonts-and-palette

Ordered so the preview/test tooling lands before glyph authoring.

## 1. Tooling first (so glyphs can be seen & regression-guarded)

- [ ] 1.1 Add `cmd/fontgallery/main.go` — headless, stdout-only: render a fixed
  sample (`SHIP IT! BUILD FAILED 0123`) in every font, then a full-charset dump
  per font. No TTY, no Bubble Tea.
- [ ] 1.2 Add golden snapshot tests: freeze each font's full-charset render to
  `internal/font/testdata/<name>.golden`; compare in `TestFontGolden`, with an
  `-update` flag to regenerate.
- [ ] 1.3 Extend `TestBlockRowWidths` to assert the equal-width invariant across
  all three fonts' full charsets.

## 2. Fonts: cut to three, uppercase-only, static maps

- [ ] 2.1 Define the uppercase charset constant (`A–Z 0–9 ! ? . , : ; ' - / & % ( )`
  + space) and a coverage test asserting every font supplies every glyph.
- [ ] 2.2 Re-author `block` (clean medium) glyph map; verify in the gallery.
- [ ] 2.3 Re-author `heavy` (solid filled) glyph map by hand (no longer derived
  from banner3); verify in the gallery.
- [ ] 2.4 Re-author `compact` (3-row half-block) glyph map, transcribing from
  TOIlet `pagga` where it fits; verify in the gallery.
- [ ] 2.5 Record provenance + licence for any transcribed glyphs (header comment
  / `NOTICE`); confirm each source's licence before copying.
- [ ] 2.6 Delete `internal/font/figlet.go` and any `mappedFiglet`/`figletFont`
  machinery; update the registry to register only the three fonts.
- [ ] 2.7 Remove `go-figure` from `go.mod`/`go.sum` (`go mod tidy`); confirm it's
  gone from the dependency graph.

## 3. Colours: 5 semantic adaptive presets

- [ ] 3.1 Rewrite `internal/cmd/styles.go`: `alert`/`warn`/`info`/`ok`/`focus` as
  `lipgloss.AdaptiveColor` Light/Dark pairs; `info` is the default; remove the
  old 10 presets and `presetOrder`.
- [ ] 3.2 Tune each Light/Dark pair for contrast on both backgrounds (eyeball via
  a temporary render; keep the current adaptive teal as `info`).
- [ ] 3.3 Keep `parseColor` accepting hex + ANSI index; update the preset branch
  and its valid-name list.

## 4. Hard-error resolution

- [ ] 4.1 In `internal/cmd/root.go`, reject an unknown named font/colour at style
  resolution (before any `font.Get` fallback) with a non-zero exit and a message
  listing valid names — for both flags and config-file values.
- [ ] 4.2 Tests: unknown font and unknown colour each error and name the valid
  options; valid presets/hex/ANSI still succeed.

## 5. Config picker tolerance

- [ ] 5.1 In `internal/cmd/config.go`, make the picker seed an invalid stored
  font/colour as the default selection instead of erroring; update `fontOptions`
  / `colorOptions` to the new sets.
- [ ] 5.2 Test: a config with a removed font/colour name still opens the picker
  with defaults selected; saving writes only valid values.

## 6. Docs & changelog

- [ ] 6.1 Update `README.md` (fonts, semantic palette, escape hatch, hard-error
  behaviour) and `AGENTS.md` (font architecture, removed `go-figure`/FIGlet).
- [ ] 6.2 Note the breaking removals (FIGlet fonts + old presets) in the release
  notes / changelog. Do **not** bump `VERSION` here — that triggers auto-tag on
  push; bump it as a deliberate release step.

## 7. Validate

- [ ] 7.1 `gofmt -w`, `go vet ./...`, `go build ./...`, `go test ./...` all green.
- [ ] 7.2 Run `cmd/fontgallery` and eyeball all three fonts end-to-end.
- [ ] 7.3 `openspec validate refine-fonts-and-palette --strict`.
