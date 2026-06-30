package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const githubAPI = "https://api.github.com/repos/mitchell-wallace/plaqq/releases/latest"

var (
	updateYes              bool
	fetchLatestVersionFunc = fetchLatestVersion
	installLatestVersionFn = installLatestVersion
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for a newer version and update",
	Long: `Check the GitHub releases page for a newer version of plaqq.

Prints the current and latest versions. If a newer version is available,
runs the install script immediately.`,
	Run: func(cmd *cobra.Command, args []string) {
		latest, err := fetchLatestVersionFunc()
		if err != nil {
			exit(2, "update: %v", err)
		}

		// A non-release version (empty, "dev", a bare git build hash, or a
		// "-dev" local build) can't be compared against the latest release.
		// Rather than failing, treat it as eligible for installing the latest
		// release so `plaqq update` is useful from a locally built binary.
		dev := !isReleaseVersion(version)
		current := version
		if current == "" {
			current = "dev"
		}

		if !dev {
			cmp, err := compareVersions(version, latest)
			if err != nil {
				exit(2, "update: %v", err)
			}
			if cmp >= 0 {
				if jsonOutput {
					printJSON(map[string]any{
						"currentVersion": current,
						"latestVersion":  latest,
						"upToDate":       true,
						"updated":        false,
					})
				} else {
					fmt.Printf("Current version: %s\nLatest version:  %s\n", current, latest)
					fmt.Println("You are up to date.")
				}
				return
			}
		}

		if !jsonOutput {
			fmt.Printf("Current version: %s\nLatest version:  %s\n", current, latest)
			if dev {
				fmt.Println("You are running a development build.")
			}
		}

		if err := installLatestVersionFn(); err != nil {
			exit(2, "update: install failed: %v", err)
		}
		if jsonOutput {
			printJSON(map[string]any{
				"currentVersion": current,
				"latestVersion":  latest,
				"devBuild":       dev,
				"upToDate":       false,
				"updated":        true,
			})
		}
	},
}

func init() {
	updateCmd.Flags().BoolVarP(&updateYes, "yes", "y", false, "accepted for compatibility; update no longer prompts")
	rootCmd.AddCommand(updateCmd)
}

func fetchLatestVersion() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", githubAPI, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else if token := os.Getenv("GH_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		if (resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests) && resp.Header.Get("X-RateLimit-Remaining") == "0" {
			return "", fmt.Errorf("GitHub API rate limit exceeded. Please try again later or set GITHUB_TOKEN environment variable")
		}
		return "", fmt.Errorf("github API returned %s", resp.Status)
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("parse release: %w", err)
	}

	return strings.TrimPrefix(payload.TagName, "v"), nil
}

func installLatestVersion() error {
	var installCmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		installCmd = exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", "irm https://raw.githubusercontent.com/mitchell-wallace/plaqq/main/install.ps1 | iex")
	default:
		installCmd = exec.Command("sh", "-c", "curl -fsSL https://raw.githubusercontent.com/mitchell-wallace/plaqq/main/install.sh | bash")
	}
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	return installCmd.Run()
}

// isReleaseVersion reports whether v is a clean release version of the form
// X.Y.Z (an optional leading "v" is allowed). Development builds — the empty
// string, "dev", a bare git build hash, or anything carrying a pre-release or
// build suffix (e.g. "0.3.0-dev+1a06be9", "0.3.0-5-g1a06be9-dirty") — return
// false, since they cannot be meaningfully compared against a release tag.
func isReleaseVersion(v string) bool {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		if _, err := strconv.Atoi(p); err != nil {
			return false
		}
	}
	return true
}

func compareVersions(a, b string) (int, error) {
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "v")
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}
	for i := 0; i < maxLen; i++ {
		var av, bv int
		if i < len(aParts) {
			v, err := strconv.Atoi(aParts[i])
			if err != nil {
				return 0, fmt.Errorf("invalid version %q", a)
			}
			av = v
		}
		if i < len(bParts) {
			v, err := strconv.Atoi(bParts[i])
			if err != nil {
				return 0, fmt.Errorf("invalid version %q", b)
			}
			bv = v
		}
		if av < bv {
			return -1, nil
		}
		if av > bv {
			return 1, nil
		}
	}
	return 0, nil
}
