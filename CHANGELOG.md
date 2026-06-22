# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Breaking Changes
- **FIGlet Fonts Removed**: Removed all 9 `go-figure` FIGlet ASCII fonts (`standard`, `slant`, `banner`, `big`, `small`, `doom`, `larry3d`, `mini`, `cyberlarge`). `plaqq` now ships with exactly 3 Unicode block fonts (`block`, `heavy`, `compact`).
- **Color Presets Overhauled**: Removed the 10 old fixed color presets (`teal`, `coral`, `amber`, `lime`, `azure`, `violet`, `magenta`, `rose`, `crimson`, `slate`).
- **Semantic Palette introduced**: Added 5 semantic theme-adaptive presets (`alert`, `warn`, `info` [default], `ok`, `focus`) configured as Light/Dark adaptive color pairs that adjust automatically to terminal background light/dark modes.
- **Strict Validation**: Invalid font or color preset names in CLI flags or config files now trigger a hard error (non-zero exit) with nearest-match suggestions, rather than silently falling back to defaults. (The `plaqq config` command remains tolerant to permit recovery).

### Added
- **Curated Fonts**: Added committed glyph maps for `block` (clean 5-row medium block), `heavy` (7-row solid block), and `compact` (3-row half-block transcribed from TOIlet `pagga`).
- **Edit Distance Suggestions**: Displays spelling corrections for unrecognized presets (e.g., `did you mean "heavy"?`).
- **Zero Runtime Dependencies**: Removed `go-figure` dependency, eliminating all external runtime font rendering libraries.
