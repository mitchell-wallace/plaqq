# Distribution & self-update — full reference

Complete, copy-paste building blocks for shipping a single-binary Go CLI via
GitHub Releases and letting it update itself. Replace `mytool` with your binary
name and `you/mytool` with the repo. The reference implementation is `plaqq`
(this repo): `internal/cmd/update.go`, `install.sh`, `install.ps1`, `justfile`,
`.goreleaser.yaml`, and `.github/workflows/{auto-tag,release}.yml`.

## How the pieces fit

```
VERSION file ── push to main ──▶ auto-tag.yml ── creates tag vX.Y.Z ──▶ release.yml
                                                                            │
                                                            goreleaser builds matrix
                                                                            │
                                          GitHub Release: mytool_X.Y.Z_os_arch.tar.gz
                                                          + checksums.txt
                                                                            │
        install.sh / install.ps1 ◀── download asset ──── mytool update ◀────┘
```

Three contracts must stay in sync:
1. **Asset name** — `name_template` in `.goreleaser.yaml` == the name the install
   scripts build == what `mytool update` expects.
2. **Version string** — `VERSION` file (`0.3.1`) vs git tag (`v0.3.1`): strip the
   leading `v` wherever they meet.
3. **Install location** — the scripts' `INSTALL_DIR` is where `update` re-installs.

## Version embedding

`main.version` defaults to `"dev"` and is overridden at link time.

```yaml
# .goreleaser.yaml — release builds get the tag version
builds:
  - id: mytool
    main: ./cmd/mytool
    binary: mytool
    ldflags: [-s -w -X main.version={{.Version}}]
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
archives:
  - format: tar.gz
    format_overrides: [{ goos: windows, format: zip }]
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
checksum:
  name_template: "checksums.txt"
```

```just
# justfile — local builds get a meaningful dev version, NOT a bare git hash.
# Base = VERSION file; marked -dev with the short commit; .dirty if uncommitted.
version := shell('base="$(tr -d "[:space:]" < VERSION 2>/dev/null || echo 0.0.0)"; hash="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"; dirty=""; [ -n "$(git status --porcelain 2>/dev/null)" ] && dirty=".dirty"; printf "%s-dev+%s%s" "$base" "$hash" "$dirty"')

build:
	go build -ldflags "-X main.version={{version}}" -o bin/mytool ./cmd/mytool

# Install locally via atomic rename so it works even while an old copy runs.
install: build
	mkdir -p ~/.local/bin
	cp bin/mytool ~/.local/bin/.mytool.new
	chmod +x ~/.local/bin/.mytool.new
	mv -f ~/.local/bin/.mytool.new ~/.local/bin/mytool
```

> Why not `git describe --tags --always --dirty`? With no tags fetched it
> returns a bare hash like `1a06be9`, which is not semver and crashes a naive
> version compare. Basing the dev version on the `VERSION` file avoids that and
> keeps local builds clearly marked as `-dev`.

## Release automation

```yaml
# .github/workflows/auto-tag.yml — tag when VERSION changes on main
on: { push: { branches: [main] } }
permissions: { contents: write, actions: write }
jobs:
  auto-tag:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 2 }
      - run: |
          if git diff HEAD~1 --quiet -- VERSION; then echo "unchanged"; exit 0; fi
          RAW="$(tr -d '[:space:]' < VERSION)"
          printf '%s' "$RAW" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$' || { echo "VERSION must be semver"; exit 1; }
          TAG="v$RAW"
          if git ls-remote --exit-code --tags origin "refs/tags/$TAG" >/dev/null; then
            echo "tag exists"
          else
            git config user.name  "github-actions[bot]"
            git config user.email "github-actions[bot]@users.noreply.github.com"
            git tag "$TAG" && git push origin "$TAG"
          fi
        env: { GH_TOKEN: "${{ github.token }}" }
```

`release.yml` triggers on `tags: ["v*"]`, runs `goreleaser/goreleaser-action`,
and should no-op if the release already exists (idempotent re-runs). To cut a
release: edit `VERSION`, commit, push to `main`. That's the only manual step.

## install.sh (Linux & macOS)

Key points: resolve version (arg > `MYTOOL_VERSION` > latest), map `uname` to
goreleaser's os/arch, download the asset, and **install via a same-filesystem
temp + atomic rename** so an in-use binary doesn't cause `text file busy`.

```sh
#!/usr/bin/env bash
set -euo pipefail
REPO="you/mytool"
INSTALL_DIR="$HOME/.local/bin"
VERSION="${1:-${MYTOOL_VERSION:-}}"; VERSION="${VERSION#v}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in linux) OS=linux;; darwin) OS=darwin;; *) echo "unsupported OS $OS"; exit 1;; esac
ARCH=$(uname -m)
case "$ARCH" in x86_64) ARCH=amd64;; aarch64|arm64) ARCH=arm64;; *) echo "unsupported arch $ARCH"; exit 1;; esac

if [ -z "$VERSION" ]; then
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name":' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/'); VERSION="${VERSION#v}"
  [ -n "$VERSION" ] || { echo "failed to resolve latest release"; exit 1; }
fi

ASSET="mytool_${VERSION}_${OS}_${ARCH}.tar.gz"
curl -fsSL "https://github.com/${REPO}/releases/download/v${VERSION}/${ASSET}" -o "/tmp/${ASSET}"

mkdir -p "$INSTALL_DIR"
# Atomic install: extract to a temp file IN THE INSTALL DIR (same filesystem),
# then rename over the target. A cross-filesystem mv would copy-and-truncate and
# re-trigger ETXTBSY on a running binary, so the temp must live here.
TMP="$(mktemp "${INSTALL_DIR}/.mytool.XXXXXX")"
trap 'rm -f "$TMP" "/tmp/${ASSET}"' EXIT
tar -xzOf "/tmp/${ASSET}" mytool > "$TMP"
chmod +x "$TMP"
mv -f "$TMP" "$INSTALL_DIR/mytool"
trap - EXIT
rm -f "/tmp/${ASSET}"
echo "Installed mytool ${VERSION} to ${INSTALL_DIR}/mytool"
# ...then append "$INSTALL_DIR" to PATH in ~/.bashrc / ~/.zshrc / fish config if missing.
```

## install.ps1 (Windows)

Windows won't overwrite or delete a running `.exe`, but it *will* let you rename
one. Extract to a temp dir, move a locked target aside, then move the new one in.

```powershell
$ErrorActionPreference = "Stop"
$Repo = "you/mytool"
$InstallDir = Join-Path $env:LOCALAPPDATA "Programs\mytool"
$Arch = switch ($env:PROCESSOR_ARCHITECTURE) { "AMD64" {"amd64"} "ARM64" {"arm64"} default { throw "unsupported arch" } }

$Tag = (Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest").tag_name
$Version = $Tag -replace '^v',''
$Asset = "mytool_${Version}_windows_${Arch}.zip"
$Tmp = Join-Path $env:TEMP $Asset
Invoke-WebRequest "https://github.com/$Repo/releases/download/$Tag/$Asset" -OutFile $Tmp -UseBasicParsing

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$Extract = Join-Path $env:TEMP ("mytool_" + [guid]::NewGuid().ToString("N"))
Expand-Archive -Path $Tmp -DestinationPath $Extract -Force

$Target = Join-Path $InstallDir "mytool.exe"
if (Test-Path $Target) {
  $Old = Join-Path $InstallDir "mytool.old.exe"
  Remove-Item $Old -Force -ErrorAction SilentlyContinue
  try { Move-Item $Target $Old -Force }      # rename a running exe aside
  catch { Write-Error "Could not replace $Target (is mytool running?). Close it and retry."; exit 1 }
}
Get-ChildItem $Extract | ForEach-Object { Move-Item $_.FullName (Join-Path $InstallDir $_.Name) -Force }
Remove-Item $Extract -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item $Tmp -Force
# ...then add $InstallDir to the user PATH if missing.
```

## The `update` command

Fetch the latest tag, decide whether to install, then re-run the platform
installer. The whole correctness story is in the `dev := !isReleaseVersion(...)`
branch: a development build must never reach `compareVersions`.

```go
const githubAPI = "https://api.github.com/repos/you/mytool/releases/latest"

var (
	updateYes              bool
	fetchLatestVersionFunc = fetchLatestVersion   // swapped in tests
	installLatestVersionFn = installLatestVersion
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for a newer version and optionally update",
	Run: func(cmd *cobra.Command, args []string) {
		latest, err := fetchLatestVersionFunc()
		if err != nil { exit(2, "update: %v", err) }

		// A non-release version (empty, "dev", a bare hash, or a -dev/-dirty
		// build) can't be compared to a release tag. Treat it as eligible to
		// install the latest release instead of failing.
		dev := !isReleaseVersion(version)
		current := version
		if current == "" { current = "dev" }

		if !dev {
			cmp, err := compareVersions(version, latest)
			if err != nil { exit(2, "update: %v", err) }
			if cmp >= 0 {
				report(current, latest, map[string]any{"upToDate": true, "updated": false},
					"Current version: %s\nLatest version:  %s\nYou are up to date.\n")
				return
			}
		}

		if !updateYes {
			if jsonOutput {
				printJSON(map[string]any{"currentVersion": current, "latestVersion": latest,
					"devBuild": dev, "upToDate": false, "updated": false})
				return
			}
			fmt.Printf("Current version: %s\nLatest version:  %s\n", current, latest)
			prompt := "Update to latest version? [Y/n] "
			if dev { fmt.Println("You are running a development build."); prompt = "Install the latest release? [Y/n] " }
			fmt.Print(prompt)
			r, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			if r = strings.TrimSpace(strings.ToLower(r)); r != "" && r != "y" && r != "yes" {
				fmt.Println("Update cancelled."); return
			}
		}

		if err := installLatestVersionFn(); err != nil { exit(2, "update: install failed: %v", err) }
		if jsonOutput {
			printJSON(map[string]any{"currentVersion": current, "latestVersion": latest,
				"devBuild": dev, "upToDate": false, "updated": true})
		}
	},
}

// installLatestVersion re-runs the one-line installer for the platform.
func installLatestVersion() error {
	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass",
			"-Command", "irm https://raw.githubusercontent.com/you/mytool/main/install.ps1 | iex")
	} else {
		c = exec.Command("sh", "-c",
			"curl -fsSL https://raw.githubusercontent.com/you/mytool/main/install.sh | bash")
	}
	c.Stdout, c.Stderr = os.Stdout, os.Stderr
	return c.Run()
}
```

```go
// isReleaseVersion: strict X.Y.Z (optional leading v). Everything else — "",
// "dev", "1a06be9", "0.3.1-dev+1a06be9.dirty", "0.3.1-5-g1a06be9" — is a dev
// build and returns false.
func isReleaseVersion(v string) bool {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	if len(parts) != 3 { return false }
	for _, p := range parts {
		if p == "" { return false }
		if _, err := strconv.Atoi(p); err != nil { return false }
	}
	return true
}

// compareVersions: -1/0/1 for a<b/a==b/a>b over numeric X.Y.Z. Only ever called
// on values that passed isReleaseVersion.
func compareVersions(a, b string) (int, error) {
	a, b = strings.TrimPrefix(a, "v"), strings.TrimPrefix(b, "v")
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	n := len(as); if len(bs) > n { n = len(bs) }
	for i := 0; i < n; i++ {
		av, bv := 0, 0
		if i < len(as) { v, err := strconv.Atoi(as[i]); if err != nil { return 0, fmt.Errorf("invalid version %q", a) }; av = v }
		if i < len(bs) { v, err := strconv.Atoi(bs[i]); if err != nil { return 0, fmt.Errorf("invalid version %q", b) }; bv = v }
		if av != bv { if av < bv { return -1, nil }; return 1, nil }
	}
	return 0, nil
}
```

## Background "update available" notice

A foreground UI can opportunistically check for updates without ever blocking:

- Run it in a goroutine with a **short timeout (~2s)** and a buffered channel.
- **Swallow every error** — a failed check must be invisible.
- **Gate on `isReleaseVersion(version)`** so dev builds neither nag nor make the
  network call.
- Print the notice only *after* any alt-screen TUI is torn down.

```go
if isReleaseVersion(version) {
	go func() {
		// 2s client, fetch latest tag, compareVersions(version, latest);
		// on cmp < 0 send a one-line "vX is available, run `mytool update`".
	}()
}
```

## Auth & rate limits

Unauthenticated GitHub API calls are rate-limited (~60/h per IP). Forward a
token when present so power users and CI don't hit the wall:

```go
if t := os.Getenv("GITHUB_TOKEN"); t != "" { req.Header.Set("Authorization", "Bearer "+t) } else
if t := os.Getenv("GH_TOKEN"); t != "" { req.Header.Set("Authorization", "Bearer "+t) }
```

Detect the rate-limit response specifically (`403`/`429` with
`X-RateLimit-Remaining: 0`) and return a message that tells the user to set
`GITHUB_TOKEN`, rather than a generic HTTP error.

## Testing the update path

Swap the injected function vars; never hit the network or install for real:

```go
func TestUpdateDevBuildInstalls(t *testing.T) {
	o1, o2, o3, o4 := fetchLatestVersionFunc, installLatestVersionFn, updateYes, version
	t.Cleanup(func() { fetchLatestVersionFunc, installLatestVersionFn, updateYes, version = o1, o2, o3, o4 })
	fetchLatestVersionFunc = func() (string, error) { return "0.3.1", nil }
	installed := false
	installLatestVersionFn = func() error { installed = true; return nil }
	updateYes, version = true, "1a06be9" // bare hash that previously crashed compareVersions
	updateCmd.Run(updateCmd, nil)
	if !installed { t.Error("dev build update should install the latest release") }
}
```
