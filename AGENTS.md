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
  - `internal/font/`: Font subsystem. `registry.go` defines the `Font` interface plus `Get`/`Names`/`Wrap`; `font.go` holds the default 5-row block glyphs; `block.go` is a generic glyph-table font; `compact.go` is a hand-authored 3-row block font; `figlet.go` wraps go-figure FIGlet fonts (and derives the `heavy` block font from banner3).
  - `internal/cmd/update.go`: Implements the `update` command checking GitHub releases and triggering update scripts.

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

- **Font Registry**: Fonts implement the `font.Font` interface (`Render(line) []string`, `Width(s) int`) and register themselves under a name. `font.Get(name)` resolves a font (falling back to `block`); `font.Names()` lists them with the default first. New fonts: a `map[rune][]string` wrapped in `NewBlock`, or a `figletFont`/`mappedFiglet` entry.
- **Two Font Families**: Unicode block faces (`block`, `heavy`, `compact`) and FIGlet ASCII faces from go-figure. `heavy` is banner3 transliterated `#`→`█`. Block fonts upper-case input; each block glyph's rows must share one width (enforced by `TestBlockRowWidths`).
- **Word Wrapping**: `font.Wrap(f, text, maxWidth)` greedily packs whole words using the font's own `Width`. Max columns = terminal width − 8.
- **Centering**: Each wrapped line is rendered as a block and centered with one uniform pad so ragged-width FIGlet rows stay column-aligned (see `model.View`).
- **Color**: Named presets in `internal/cmd/styles.go`; `parseColor` accepts a preset, a hex code, or an ANSI index. Default is adaptive teal (`#00f5d4` dark / `#00d7af` light).
- **Persistent Config**: `color`, `font`, `bold`, `hint`, `no_hint` in TOML. `config.Save` writes only set fields; the interactive picker (`plaqq config`) seeds from the current file.
- **Key Bindings**: Pressing the `Spacebar`, `Enter`, `Esc`, `q`, or `Ctrl+C` immediately quits the application, returning you back to the main terminal screen.
