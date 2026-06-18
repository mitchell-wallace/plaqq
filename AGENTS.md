# AGENTS.md (Developer & AI Agent Reference for Plaqq)

This document describes the mechanics, design guidelines, and automation rules for the `plaqq` CLI tool.

---

## 🏗️ Codebase Overview

- **Core Technology**: Go 1.26+ and Bubble Tea (interactive terminal UI framework) + Lipgloss (styling).
- **Primary Layout Files**:
  - `cmd/plaqq/main.go`: Entry point setting the version variable and invoking Cobra.
  - `internal/cmd/root.go`: Main interactive app using Bubble Tea. Runs in alternate screen buffer and handles key/window resize events.
  - `internal/font/font.go`: Defines the 5-row custom chunky block font glyphs and word-wrapping routines.
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

- **Chunky Block Glyph Set**: Standard block unicode elements (`█`, `▄`, `▀`) are mapped in a `map[rune][5]string` array.
- **Word Wrapping**: Handled in `font.WrapText`. Calculates maximum columns allowed (terminal width - 8) and dynamically wraps lines to avoid clipping.
- **Color Theme**: Adaptive teal (`#00f5d4` on dark terminals, `#00d7af` on light terminals).
- **Key Bindings**: Pressing the `Spacebar`, `Enter`, `Esc`, `q`, or `Ctrl+C` immediately quits the application, returning you back to the main terminal screen.
