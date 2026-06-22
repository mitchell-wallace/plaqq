# plaqq

`plaqq` is a lightweight Go CLI tool for displaying stylized notices in terminal panes. 

It takes any custom notice text, renders it in a centered, chunky ASCII block font using adaptive terminal styling, and prompts the user to dismiss it with a single keystroke.

---

## Features

- **Multiple Fonts**: Ships Unicode block faces (`block`, `heavy`, `compact`).
- **Color Presets**: Pick a named color (`teal`, `coral`, `amber`, `lime`, `azure`, `violet`, `magenta`, `rose`, `crimson`, `slate`) or supply your own hex / ANSI value.
- **Interactive Config Picker**: Run `plaqq config` (no subcommand) to edit your styling defaults in a friendly form.
- **Dynamic Centering & Word Wrapping**: Automatically wraps text to fit within your terminal pane margin and keeps the notice perfectly centered vertically and horizontally.
- **Adaptive Theme Styling**: Looks great on both light and dark terminals using adaptive theme styling.
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

*   **`--font`**: Font used to render the notice. Block faces: `block` (default), `heavy`, `compact`.
    ```bash
    plaqq --font heavy "shipped"
    plaqq --font compact "heads up"
    ```
*   **`--color`**: Notice text color as a preset name (`teal`, `coral`, `amber`, `lime`, `azure`, `violet`, `magenta`, `rose`, `crimson`, `slate`), a hex code (`#00f5d4`), or an ANSI index (`0`-`255`). Defaults to an adaptive teal that suits both light and dark terminals.
    ```bash
    plaqq --color coral "build failed"
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
color = "coral"
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
