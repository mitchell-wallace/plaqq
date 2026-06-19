# Colors

## ADDED Requirements

### Requirement: Semantic adaptive palette

The system SHALL provide exactly five named colours — `alert`, `warn`, `info`,
`ok`, and `focus` — and no other named presets. Each SHALL be an adaptive colour
with distinct light- and dark-background variants chosen for high contrast on
both. `info` SHALL be the default notice colour.

#### Scenario: Default colour

- **WHEN** a notice is rendered with no colour specified
- **THEN** it uses the `info` adaptive colour

#### Scenario: Each preset is adaptive

- **WHEN** any of the five presets is resolved
- **THEN** it yields a light variant and a dark variant (not a single fixed value)

#### Scenario: Listing colours

- **WHEN** the available named colours are listed
- **THEN** exactly `alert`, `warn`, `info`, `ok`, and `focus` are offered

### Requirement: Raw colour escape hatch

The system SHALL accept a raw colour in addition to the named presets: a hex code
(`#rrggbb`) or an ANSI 256 index (`0`–`255`).

#### Scenario: Hex colour

- **WHEN** `plaqq --color "#ff5f87" "hi"` is run
- **THEN** the notice renders in that colour without error

#### Scenario: ANSI index

- **WHEN** `plaqq --color 213 "hi"` is run
- **THEN** the notice renders in ANSI colour 213 without error

### Requirement: Unknown colour name is a hard error

The system SHALL reject a colour that is neither one of the five presets, a valid
hex code, nor a valid ANSI index — failing with a non-zero exit and a message
listing the valid preset names, whether the value comes from the `--color` flag
or a config-file value.

#### Scenario: Removed preset via flag

- **WHEN** `plaqq --color coral "hi"` is run
- **THEN** the command exits non-zero and the error names `alert`, `warn`, `info`, `ok`, `focus`

#### Scenario: Unknown colour via config file

- **WHEN** the config file sets a removed preset name and a notice is rendered
- **THEN** the command exits non-zero with an error listing the valid preset names
