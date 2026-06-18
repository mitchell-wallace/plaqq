# plaqq

`plaqq` is a lightweight Go CLI tool for displaying stylized notices in terminal panes. 

It takes any custom notice text, renders it in a centered, chunky ASCII block font using adaptive terminal styling, and prompts the user to dismiss it with a single keystroke.

---

## Features

- **Large Chunky Block Font**: Renders notices using beautiful high-contrast unicode block characters (`█`, `▄`, `▀`).
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

*   **`--color`**: Notice text color as a hex code (`#00f5d4`) or ANSI index (`0`-`255`). Defaults to an adaptive teal that suits both light and dark terminals.
    ```bash
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
