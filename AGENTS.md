# AGENTS.md (Developer & AI Agent Reference for Plaqq)

This document describes the mechanics, design guidelines, and automation rules for the `plaqq` CLI tool.

---

## 🏗️ Codebase Overview

- **Core Technology**: Go 1.26+ and Bubble Tea (interactive terminal UI framework) + Lipgloss (styling).
- **Primary Layout Files**:
  - `cmd/plaqq/main.go`: Entry point setting the version variable and invoking Cobra.
  - `internal/cmd/root.go`: Main interactive app using Bubble Tea. Runs in alternate screen buffer and handles key/window resize events.
  - `internal/cmd/styles.go`: Named color presets and lookup.
  - `internal/cmd/config.go`: `config` command, including the interactive `huh`-based picker for bare `plaqq config`.
  - `internal/font/`: Font subsystem. `registry.go` defines the `Font` interface plus `Get`/`Names`/`Wrap` and the `Block` glyph renderer; `block.go` holds the default 5-row block glyphs; `heavy.go` holds the 7-row solid block glyphs; `compact.go` is a 3-row half-block font (mostly transcribed from TOIlet `pagga`, with re-authored Q/M/W); `wide.go` is a roomier 5-row medium-weight face.
  - `internal/cmd/update.go`: Implements the `update` command checking GitHub releases and triggering update scripts.
  - `cmd/fontgallery/`: Dev-only visual harness. `just visual` renders every font to `artifacts/visual/<font>.png` (faithful cell-painted glyph geometry, font-independent), plus `<font>.ansi` and `manifest.json`. Inspect the PNGs when authoring or reviewing glyphs; use the `.ansi` files for exact cell-level debugging.

---

## 🔄 VERSION Auto-Tagging & CI/CD Flow

To release a new version of `plaqq`:
1. Increment the version in the `VERSION` file (e.g. `0.2.0`). Follow Semantic Versioning.
2. Push your changes directly to the `main` branch.
3. The `.github/workflows/auto-tag.yml` workflow triggers automatically. It checks if the `VERSION` file changed.
4. If changed, the workflow reads the version, configures git user, tags the commit (e.g. `v0.2.0`), and pushes the tag to the remote repository.
5. Pushing the tag automatically triggers `.github/workflows/release.yml`, which executes `goreleaser` to build binaries for macOS, Linux, and Windows and publishes them under a new GitHub release.

---

## 🎨 Layout & Font Architecture

- **Font Registry**: Fonts implement the `font.Font` interface (`Render(line) []string`, `Width(s) int`) and register themselves under a name. `font.Get(name)` resolves a font (falling back to `compact`); `font.Names()` lists them with the default first. New fonts are added as a `map[rune][]string` wrapped in `NewBlock`.
- **Zero Runtime Dependencies**: The `go-figure` dependency and FIGlet ASCII faces are removed. Fonts are defined entirely by static, committed Unicode block glyph maps (`block`, `heavy`, `compact`, `wide`) packaged inside the binary. Block fonts upper-case input; each block glyph's rows must share one width (enforced by `TestBlockRowWidths`).
- **Word Wrapping**: `font.Wrap(f, text, maxWidth)` greedily packs whole words using the font's own `Width`. Max columns = terminal width − 8.
- **Centering**: Each wrapped line is rendered as a block and centered with one uniform pad so that column alignments stay correct (see `model.View`).
- **Color**: Named presets in `internal/cmd/styles.go` are five semantic, adaptive colors (`alert`, `warn`, `info`, `ok`, `focus`) configured as Light/Dark `lipgloss.AdaptiveColor` pairs to ensure legibility on any background. `parseColor` accepts a preset name, a hex code, or an ANSI index. Default is `info` (adaptive teal).
- **Persistent Config**: `color`, `font`, `bold`, `hint`, `no_hint` in TOML. `config.Save` writes only set fields; the interactive picker (`plaqq config`) seeds from the current file.
- **Style Resolution & Precedence**: `resolveStyle` in `internal/cmd/root.go` layers styles from multiple sources with precedence: built-in defaults < config file < session env vars (`PLAQQ_COLOR`, `PLAQQ_FONT`, `PLAQQ_BOLD`, `PLAQQ_HINT`, `PLAQQ_NO_HINT`) < session state < CLI flags. Flags/config invalid values error immediately, while env/session invalids log warning to stderr and fall through.
- **Session-State Store**: Managed by `internal/session`, keyed by parent shell PID (`os.Getppid()`). Written by the customise flow and `plaqq config --session`, cleared via `plaqq config --session --clear`. Stores temporary state as TOML under `$XDG_RUNTIME_DIR/plaqq/` (or fallback temp dir) with file mode `0600`.
- **Key Bindings**: Pressing the `Spacebar`, `Enter`, `Esc`, `q`, or `Ctrl+C` immediately quits the application, returning you back to the main terminal screen.
