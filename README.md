# plaqq

`plaqq` is a lightweight Go CLI tool for displaying stylized notices in terminal panes.

It takes any custom notice text, renders it in a centered, chunky Unicode block font using adaptive terminal styling, and lets the user dismiss or edit it with a single keystroke.

---

## Features

- **Four Fonts**: Ships four Unicode block faces: `compact` (dense 3-row half-block, default), `block` (clean medium), `heavy` (solid filled block), and `wide` (a roomier, more open medium-weight face).
- **Semantic Adaptive Palette**: Pick from five named semantic colors (`alert`, `warn`, `info` (default), `ok`, `focus`), each utilizing adaptive theme styling to look great on both light and dark terminal backgrounds.
- **Escape Hatch**: Supply custom hex codes (e.g., `#00f5d4`) or ANSI color indexes (`0`-`255`) for personalized coloring.
- **Strict Validation**: Invalid font or color preset names in CLI flags or config files trigger a non-zero exit status with a helpful suggestion message.
- **Interactive Config Picker**: Run `plaqq config` (no subcommand) to edit your styling defaults in a friendly, interactive form (which tolerates invalid stored configurations for easy repair).
- **Dynamic Centering & Word Wrapping**: Automatically wraps text to fit within your terminal pane margin and keeps the notice perfectly centered vertically and horizontally.
- **Fast Notice Actions**: Press `Space` to dismiss, or `Enter` to reopen the prompt and edit the message/font/color.
- **Per-Pane Continue**: `plaqq --continue` or `plaqq -c` reopens the last notice text from the current terminal pane.
- **Background Update Notification**: Automatically checks for newer versions and alerts you when updates are available.

---

## Installation

### Unix (Linux & macOS)

Run the following command to download and install to `$HOME/.local/bin`:

```bash
curl -fsSL https://raw.githubusercontent.com/mitchell-wallace/plaqq/main/install.sh | bash
```

### Windows (PowerShell)

Run the following command in PowerShell:

```powershell
irm https://raw.githubusercontent.com/mitchell-wallace/plaqq/main/install.ps1 | iex
```

---

## Command Reference

### Displaying Notices

To display a notice (if no message is specified, `plaqq` prompts you to enter one):
```bash
plaqq "remember to run e2e tests before pushing"
```

### Styling

The notice appearance can be customized with flags (run `plaqq -h` to see them all):

*   **`--font`**: Font used to render the notice. Supported block faces: `compact` (default), `block`, `heavy`, `wide`.
    ```bash
    plaqq --font heavy "shipped"
    plaqq --font compact "heads up"
    plaqq --font wide "all clear"
    ```
*   **`--color`**: Notice text color. Can be a semantic preset (`alert`, `warn`, `info`, `ok`, `focus`), a hex code (`#00f5d4`), or an ANSI index (`0`-`255`). Defaults to `info` (an adaptive teal).
    ```bash
    plaqq --color alert "build failed"
    plaqq --color "#ff5f87" "build failed"
    plaqq --color 213 "heads up"
    ```
*   **`--bold`**: Render the notice in bold (default `true`). Disable with `--bold=false`.
*   **`--hint`**: Override the dismiss-hint text shown beneath the notice.
    ```bash
    plaqq --hint "press space to continue" "meeting in 5"
    ```
*   **`--no-hint`**: Hide the dismiss hint entirely.
    ```bash
    plaqq --no-hint "stand clear"
    ```
*   **`-c`, `--continue`**: Reopen the last notice text from this terminal pane using the current resolved style.
    ```bash
    plaqq --continue
    plaqq -c
    ```

### Strict Validation

If you specify an unknown named font or color preset (either via CLI flags or in the TOML configuration file), `plaqq` exits with a non-zero status and prints a list of valid choices along with a nearest-match suggestion if one exists:

```
plaqq: unknown font "heavyy" from --font flag: valid fonts are compact, block, heavy, wide; did you mean "heavy"?
```

**Exception**: The interactive config picker (`plaqq config`) is designed to tolerate invalid config files so that it remains usable as a repair path. It seeds invalid values with defaults, allowing you to select and save valid choices.

### Persistent Configuration

Rather than passing flags every time, you can set styling defaults in a TOML config file.

The easiest way to manage it is the interactive picker — run `plaqq config` with no subcommand to navigate the font, color, bold, and hint options in a form and save your choices:

```bash
plaqq config        # interactive editor (navigate options, then save)
plaqq config path   # print the resolved config path
plaqq config init   # write a commented template to that path
```

The file lives at `$XDG_CONFIG_HOME/plaqq/config.toml` (typically `~/.config/plaqq/config.toml` on Linux, `~/Library/Application Support/plaqq/config.toml` on macOS), or wherever the `PLAQQ_CONFIG` environment variable points.

A config file looks like:

```toml
color = "alert"
font = "heavy"
bold = true
hint = "press space"
no_hint = false
```

### Style Precedence & Resolution

When determining the visual style (color, font, bold, hint, etc.) for a notice, `plaqq` resolves style properties by layering sources from lowest to highest priority:

| Priority | Source | Scope | How to Configure / Set |
| :--- | :--- | :--- | :--- |
| **1 (Lowest)** | **Built-in Defaults** | Hardcoded | Fallback settings (e.g., `info` color, `compact` font, bold enabled) |
| **2** | **Config File** | User-wide | TOML file (check path with `plaqq config path` or edit with `plaqq config`) |
| **3** | **Session Environment Variables** | Current shell process | Ambient `PLAQQ_*` environment variables (e.g., `export PLAQQ_COLOR=warn`) |
| **4** | **Session State** | Current terminal pane | Set interactively via the "Customise" flow or via `plaqq config --session` |
| **5 (Highest)** | **CLI Flags** | Single command execution | Command-line flags (e.g., `--color`, `--font`) |

A command-line flag always overrides all other sources for that specific run.

### Environment Variables (`PLAQQ_*`)

You can set ambient styles for your current shell pane using environment variables:

| Environment Variable | Style Property | Expected Format |
|---|---|---|
| `PLAQQ_COLOR` | Color | Preset name, hex code, or ANSI index |
| `PLAQQ_FONT` | Font | `block`, `heavy`, `compact`, or `wide` |
| `PLAQQ_BOLD` | Bold | `true` or `false` |
| `PLAQQ_HINT` | Hint text | String |
| `PLAQQ_NO_HINT` | Hide hint | `true` or `false` |
| `PLAQQ_TEXT` | `--continue` text override | String |

> [!NOTE]
> These variables are unrelated to `PLAQQ_CONFIG`, which is used solely to override the path to the TOML configuration file. `PLAQQ_TEXT` is only read by `plaqq --continue`; it is useful for shell-managed pane text and takes precedence over the saved last notice text.

#### Error Handling for Invalid Values

To avoid rendering failures in background/ambient scripts, environment variables are treated gently:
- If you supply an invalid style value via a **CLI flag** or **config file**, `plaqq` exits immediately with a **hard error**.
- If you supply an invalid style value via an **environment variable** or **session state**, `plaqq` prints a warning to `stderr` and safely falls back/degrades to the next priority level without halting.

### Per-Terminal Session State

Since children cannot alter their parent shell's environment variables, `plaqq` manages a temporary session-state store keyed by the parent shell's process ID (`os.Getppid()`). This allows styling choices to stick to a specific terminal pane/window.

- **Storage**: Written as TOML to `$XDG_RUNTIME_DIR/plaqq/` (or fallback temp folder) with restricted permissions (`0600`). It stores at most one last notice text plus optional session style and is automatically cleaned up when the user logs out.
- **Precedence Caveat**: Session state sits **above** environment variables. If you customize the session or set session state, a later `export PLAQQ_COLOR=...` in the same terminal pane will be overridden by the session state. Run `plaqq config --session --clear` to clear the session state and allow environment variables to take effect again.

#### Per-Pane Recipes

- **Interactively Customise**: Run bare `plaqq` with no arguments, fill in the message, and select **Customise**. Your choices will be saved to the session state for subsequent invocations in this pane.
- **Continue Last Notice**: Use `plaqq --continue` or `plaqq -c` to reopen the last notice text from this pane. `PLAQQ_TEXT` can override that saved text for shell-managed workflows.
- **Manually Save Session Style**: Use the `--session` flag with `--color` or `--font` to save styling to the current terminal pane:
  ```bash
  plaqq config --session --color alert --font heavy
  ```
- **Clear Session Style**: Reset the styling for the current shell pane back to config/env/built-in defaults:
  ```bash
  plaqq config --session --clear
  ```
- **Interactive Picker for Session**: Run `plaqq config --session` to select styling choices using an interactive picker:
  ```bash
  plaqq config --session
  ```
- **Environment Fallback**: You can also use standard shell exports for pane-wide fallback:
  ```bash
  export PLAQQ_COLOR=warn
  export PLAQQ_FONT=compact
  ```

### Interactive Confirm/Customise Flow

Running `plaqq` with no message arguments in an interactive terminal brings up an interactive form:
1. **Notice Message Input**: Enter your message.
2. **Action Select**:
   - **Confirm — show it now** (default): Renders the notice immediately using the resolved style.
   - **Customise — pick font & color**: Opens a second step to select the font and color. Submitting will save these styling options to the session state and render the notice using them.
   
You can navigate back and forth between the steps using `Tab` and `Shift+Tab`.

If `plaqq` is run without a message argument in a non-interactive terminal (e.g. CI, piped stdin) or with the `--json-output` flag, it exits with code `1` and prints `a message is required` to stderr (or JSON) rather than hanging.


### Message History

`plaqq` displays the notice on the alternate screen, which is torn down on dismissal. To keep a record in your scrollback, the notice text is echoed to stdout after final dismissal:

```
message: "remember to run e2e tests before pushing"
```

### Options & Subcommands

*   **`version`**: Prints the current version.
    ```bash
    plaqq version
    ```
*   **`update`**: Checks the GitHub releases page for a newer version and updates in-place.
    ```bash
    plaqq update
    ```
*   **`--json-output`**: Emits version or update status in structured JSON format.
    ```bash
    plaqq version --json-output
    ```

---

## Key Bindings (Notice Screen)

While the notice is active:
*   `Space`: Dismiss the notice.
*   `Enter`: Return to the prompt to edit the message/font/color.
*   `Esc`, `q` or `Ctrl+C`: Exit and return to the prompt.
