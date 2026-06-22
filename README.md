# plaqq

`plaqq` is a lightweight Go CLI tool for displaying stylized notices in terminal panes. 

It takes any custom notice text, renders it in a centered, chunky ASCII block font using adaptive terminal styling, and prompts the user to dismiss it with a single keystroke.

---

## Features

- **Three Fonts**: Ships three Unicode block faces: `block` (clean medium, default), `heavy` (solid filled block), and `compact` (dense 3-row half-block for tight spaces).
- **Semantic Adaptive Palette**: Pick from five named semantic colors (`alert`, `warn`, `info` (default), `ok`, `focus`), each utilizing adaptive theme styling to look great on both light and dark terminal backgrounds.
- **Escape Hatch**: Supply custom hex codes (e.g., `#00f5d4`) or ANSI color indexes (`0`-`255`) for personalized coloring.
- **Strict Validation**: Invalid font or color preset names in CLI flags or config files trigger a non-zero exit status with a helpful suggestion message.
- **Interactive Config Picker**: Run `plaqq config` (no subcommand) to edit your styling defaults in a friendly, interactive form (which tolerates invalid stored configurations for easy repair).
- **Dynamic Centering & Word Wrapping**: Automatically wraps text to fit within your terminal pane margin and keeps the notice perfectly centered vertically and horizontally.
- **Spacebar Dismissal**: Dismiss the overlay instantly by pressing `Space`.
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

*   **`--font`**: Font used to render the notice. Supported block faces: `block` (default), `heavy`, `compact`.
    ```bash
    plaqq --font heavy "shipped"
    plaqq --font compact "heads up"
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

### Strict Validation

If you specify an unknown named font or color preset (either via CLI flags or in the TOML configuration file), `plaqq` exits with a non-zero status and prints a list of valid choices along with a nearest-match suggestion if one exists:

```
plaqq: unknown font "heavyy" from --font flag: valid fonts are block, heavy, compact; did you mean "heavy"?
```

**Exception**: The interactive config picker (`plaqq config`) is designed to tolerate invalid config files so that it remains usable as a repair path. It seeds invalid values with defaults, allowing you to select and save valid choices.

### Persistent Configuration

Rather than passing flags every time, you can set styling defaults in a TOML config file. The resolution order is **built-in defaults → config file → CLI flags**, so a flag always wins over the config file.

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

### Message History

`plaqq` displays the notice on the alternate screen, which is torn down on dismissal. To keep a record in your scrollback, the notice text is echoed to stdout after closing:

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
*   `Space` or `Enter`: Dismiss the notice.
*   `Esc`, `q` or `Ctrl+C`: Exit and return to the prompt.
