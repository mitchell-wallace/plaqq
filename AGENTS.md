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
  - `internal/font/`: Font subsystem. `registry.go` defines the `Font` interface plus `Get`/`Names`/`Wrap` and the `Block` glyph renderer; `block.go` holds the default 5-row block glyphs; `heavy.go` holds the 7-row solid block glyphs; `compact.go` is a 3-row half-block font (mostly transcribed from TOIlet `pagga`, with re-authored Q/M/W); `wide.go` is a roomier 5-row medium-weight face; `terminal.go` renders the message as plain text (preserves case, no glyph lookup).
  - `internal/border/`: Frame subsystem. `border.go` defines the `Kind` enum (the four frame kinds `none`, `single`, `double`, `block`) plus a pure `Wrap` renderer that glues box-drawing characters around already-rendered content rows.
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

- **Font Registry**: Fonts implement the `font.Font` interface (`Render(line) []string`, `Width(s) int`) and register themselves under a name. `font.Get(name)` resolves a font (falling back to `compact`); `font.Names()` lists them with the default first. New fonts are added as a `map[rune][]string` wrapped in `NewBlock`; the `terminal` font is the plain-text exception (it preserves case and bypasses glyph lookup).
- **Zero Runtime Dependencies**: The `go-figure` dependency and FIGlet ASCII faces are removed. Fonts are defined entirely by static, committed Unicode block glyph maps (`block`, `heavy`, `compact`, `wide`) packaged inside the binary. Block fonts upper-case input; each block glyph's rows must share one width (enforced by `TestBlockRowWidths`).
- **Word Wrapping**: `font.WrapSegments(f, text, maxWidth)` splits the message on explicit `\n` (after normalizing `\r\n`/`\r`) and word-wraps each segment independently with the font's own `Width`; the flat `font.Wrap(f, text, maxWidth)` helper concatenates the segments for callers that don't need the boundaries. Max columns = terminal width − 8. Block fonts upper-case input and must never receive an embedded `\n` (it renders as a `?` glyph); the segment split in `WrapSegments` is what guarantees that.
- **Paragraph Gaps**: In `model.getScreenLines`, each `\n` in the message renders as a paragraph break of `gapRowsFor(font)` blank rows (terminal: 1, block/compact/heavy/wide: 2), and consecutive newlines stack — so `A\nB` shows 2 blank rows (block) and `A\n\nB` shows 4. Word-wrap breaks within a segment use the same per-boundary cost, and one *global* uniform left-pad centers all rows so glyph columns stay aligned across paragraphs.
- **Multiline Input**: In the interactive form, `Alt+Enter`/`Ctrl+J`/`Ctrl+Enter` insert a line break in the message field, and `Ctrl+G` (also huh's default `Ctrl+E`) opens the message in `$EDITOR` (falling back to `nano`) via `tea.ExecProcess`, which suspends the TUI and reads the edited file back. Note that `Ctrl+Enter` is not distinguishable from plain `Enter` on all terminals, so `Alt+Enter`/`Ctrl+J` remain as universal fallbacks. The editor round-trip preserves embedded newlines.
- **Centering**: Each wrapped line is rendered as a block and centered with one uniform pad so that column alignments stay correct (see `model.View`).
- **Frames**: `internal/border` defines the four frame kinds (`none`, `single`, `double`, `block`) and a pure `Wrap` renderer. `border.Kind` is an orthogonal setting that can wrap any font. Per-font horizontal margins (`terminal`: 1, `compact`/`block`/`wide`: 2, `heavy`: 3) keep the frame from crowding chunky block faces, and a 2-cell buffer is reserved on each side of the frame when content overflows the viewport so the scrollbar has room. The frame is dropped on terminals too narrow to fit at least 2 cells of content.
- **Color**: Named presets in `internal/cmd/styles.go` are five semantic, adaptive colors (`alert`, `warn`, `info`, `ok`, `focus`) configured as Light/Dark `lipgloss.AdaptiveColor` pairs to ensure legibility on any background. `parseColor` accepts a preset name, a hex code, or an ANSI index. Default is `info` (adaptive teal). `--frame` follows the same parse + nearest-match suggestion pattern.
- **Persistent Config**: `color`, `font`, `frame`, `bold`, `hint`, `no_hint` in TOML. `config.Save` writes only set fields; the interactive picker (`plaqq config`) seeds from the current file.
- **Style Resolution & Precedence**: `resolveStyle` in `internal/cmd/root.go` layers styles from multiple sources with precedence: built-in defaults < config file < session env vars (`PLAQQ_COLOR`, `PLAQQ_FONT`, `PLAQQ_FRAME`, `PLAQQ_BOLD`, `PLAQQ_HINT`, `PLAQQ_NO_HINT`) < session state < CLI flags. Flags/config invalid values error immediately, while env/session invalids log warning to stderr and fall through.
- **Session-State Store**: Managed by `internal/session`, keyed by parent shell PID (`os.Getppid()`). Written by the customise flow and `plaqq config --session`, cleared via `plaqq config --session --clear`. The session record may also carry a `frame` value alongside the existing `color`/`font` fields. Stores temporary state as TOML under `$XDG_RUNTIME_DIR/plaqq/` (or fallback temp dir) with file mode `0600`.
- **Key Bindings**: Pressing the `Spacebar`, `Esc`, `q`, or `Ctrl+C` dismisses the application, returning you back to the main terminal screen and printing a `plaqq -c` replay hint. Pressing `Enter` returns to the prompt so the message, font, or color can be edited.

---

## 🧪 Verifying Layout Changes (TUI without a TTY)

plaqq renders in the alternate screen buffer via Bubble Tea, so it needs a real TTY — piping stdin or using `script` will hang with a 0×0 size (no `tea.WindowSizeMsg`). When iterating on layout (wrapping, paragraph gaps, frames, fonts), use these approaches in order of fidelity:

1. **Unit tests first.** `model.getScreenLines()` is pure: it takes `width`/`height`/`font`/`frame` and returns the rendered rows. Tests like `TestGetScreenLinesHonorsExplicitNewlines`, `TestGetScreenLinesInterLineGapTwoForBlock`, and `font.WrapSegments` tests cover the exact gap/centering logic without any terminal. This is where the bulk of rendering logic is validated.
2. **`just visual` (offline glyph rendering).** `cmd/fontgallery` paints each font's glyphs to a PNG + `.ansi` under `artifacts/visual/`. It is for *glyph shape* review only — it calls `font.Render(line)` directly and does **not** exercise `font.Wrap`/`WrapSegments` or `getScreenLines`, so it cannot validate paragraph gaps, wrapping, or framing. Use it when authoring glyphs, not when changing layout.
3. **tmux for true end-to-end.** tmux allocates a sized PTY so the notice actually renders. Render a one-shot message and capture the pane:
   ```sh
   just build   # → bin/plaqq
   s=plaqq_e2e; tmux new-session -d -s "$s" -x 80 -y 24 \
     "bin/plaqq --font block --frame single --no-hint \$'LINE ONE\\nLINE TWO'; sleep 0.3"
   sleep 0.6; tmux capture-pane -p -t "$s" > /tmp/shot.txt; tmux kill-session -t "$s"
   ```
   To exercise the **interactive form** (Ctrl-Enter newline, Ctrl-G editor), drive it with `tmux send-keys`. Because tmux inherits its environment from the *server* (not the calling shell), set `EDITOR` inline in the session command to use a non-interactive stand-in editor that writes a known multiline payload to `$1` and exits — this proves the round-trip (typed text → editor temp file → edited payload → rendered notice) without launching a real editor. Send `M-Enter` for Alt+Enter (newline) and `C-g` for Ctrl-G. **Never run a bare `tmux kill-server`** to clean up — it kills the user's own sessions/editors; kill only the session you created by name.

