# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Breaking Changes
- **FIGlet Fonts Removed**: Removed all 9 `go-figure` FIGlet ASCII fonts (`standard`, `slant`, `banner`, `big`, `small`, `doom`, `larry3d`, `mini`, `cyberlarge`). `plaqq` now ships with Unicode block fonts only (`block`, `heavy`, `compact`, `wide`).
- **Color Presets Overhauled**: Removed the 10 old fixed color presets (`teal`, `coral`, `amber`, `lime`, `azure`, `violet`, `magenta`, `rose`, `crimson`, `slate`).
- **Semantic Palette introduced**: Added 5 semantic theme-adaptive presets (`alert`, `warn`, `info` [default], `ok`, `focus`) configured as Light/Dark adaptive color pairs that adjust automatically to terminal background light/dark modes.
- **Strict Validation**: Invalid font or color preset names in CLI flags or config files now trigger a hard error (non-zero exit) with nearest-match suggestions, rather than silently falling back to defaults. (The `plaqq config` command remains tolerant to permit recovery).

### Added
- **Curated Fonts**: Added committed glyph maps for `block` (clean 5-row medium block), `heavy` (7-row solid block), `compact` (3-row half-block, mostly transcribed from TOIlet `pagga`), and `wide` (a roomier 5-row medium-weight face).
- **Edit Distance Suggestions**: Displays spelling corrections for unrecognized presets (e.g., `did you mean "heavy"?`).
- **Zero Runtime Dependencies**: Removed `go-figure` dependency, eliminating all external runtime font rendering libraries.
- **Visual Font Harness**: `just visual` (via `cmd/fontgallery`) renders every font to PNG by painting each terminal cell's exact sub-cell geometry, alongside raw `.ansi` dumps and a `manifest.json`, so glyphs can be reviewed visually without depending on a system font.

### Changed
- **Font Glyph Cleanup**: Restored the original `block` face as the default and moved the previous re-authored block design to `wide`, then reworked it (clean K/X/Y/V/B/R/P/D/S/O). Cleaned up `heavy` (symmetric O/U, connected X centre, tidier R/B/K/Q/P/S) and improved `compact` (Q drops its backslash tail for a solid bottom-right block; M/W re-authored so they no longer collide with H/U).
