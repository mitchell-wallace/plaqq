# Fonts

## ADDED Requirements

### Requirement: Curated three-font set

The system SHALL provide exactly three named block fonts — `block`, `heavy`, and
`compact` — and no others. `block` SHALL be the default. Each font SHALL render
text as a bold, block-character figure legible at a glance across a terminal pane.

#### Scenario: Default font

- **WHEN** a notice is rendered with no font specified
- **THEN** it is rendered with the `block` font

#### Scenario: Listing fonts

- **WHEN** the available fonts are listed
- **THEN** exactly `block`, `heavy`, and `compact` are offered, with `block` first

### Requirement: Static glyph maps, zero runtime font dependency

Fonts SHALL be implemented as static, committed glyph maps. The system SHALL NOT
rasterize fonts at runtime and SHALL NOT depend on any external font library
(including `go-figure`) at runtime.

#### Scenario: No font library in the build

- **WHEN** the module's dependencies are inspected
- **THEN** `go-figure` is absent and no runtime font-rasterization dependency is present

#### Scenario: Rendering is self-contained

- **WHEN** any of the three fonts renders a line
- **THEN** the glyphs are produced from in-binary static data with no file or library access

### Requirement: Uppercase character coverage with equal-width rows

Each font SHALL provide glyphs for `A`–`Z`, `0`–`9`, the punctuation set
`! ? . , : ; ' - / & % ( )`, and space. Lower-case input SHALL be upper-cased
before rendering. Within a single glyph, every row SHALL have the same display
width.

#### Scenario: Lower-case input is upper-cased

- **WHEN** the text `ship it` is rendered
- **THEN** it renders identically to `SHIP IT`

#### Scenario: Equal-width glyph rows

- **WHEN** any covered glyph in any of the three fonts is rendered
- **THEN** all of its rows share one display width

#### Scenario: Covered punctuation renders

- **WHEN** a notice containing `! ? . , : ; ' - / & % ( )` is rendered
- **THEN** every one of those characters produces a glyph (none is dropped or substituted with the missing-glyph marker)

### Requirement: Unknown font name is a hard error

Resolving a font name that is not one of the three SHALL fail with a non-zero
exit and a message that lists the valid font names, whether the name comes from
the `--font` flag or a config-file value. The system SHALL NOT silently fall back
to the default for a named-but-unknown font at this resolution point.

#### Scenario: Unknown font via flag

- **WHEN** `plaqq --font slant "hi"` is run
- **THEN** the command exits non-zero and the error names `block`, `heavy`, `compact`

#### Scenario: Unknown font via config file

- **WHEN** the config file sets a removed font name and a notice is rendered
- **THEN** the command exits non-zero with an error listing the valid font names

### Requirement: Deterministic, inspectable rendering

The rendered output of each font over its full character set SHALL be
reproducible and verifiable without a TTY, so glyphs can be reviewed and
regressions detected automatically.

#### Scenario: Headless gallery

- **WHEN** the font-gallery tool is run without a TTY
- **THEN** it prints, to stdout, a sample render of every font and a full-charset dump per font

#### Scenario: Golden snapshot

- **WHEN** a font's glyph map changes
- **THEN** its golden snapshot test reflects the change as a diff (and fails until the snapshot is updated)
