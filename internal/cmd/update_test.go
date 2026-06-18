package cmd

import "testing"

func TestIsReleaseVersion(t *testing.T) {
	releases := []string{"0.3.0", "v0.3.0", "1.0.0", "12.34.56"}
	for _, v := range releases {
		if !isReleaseVersion(v) {
			t.Errorf("isReleaseVersion(%q) = false; want true", v)
		}
	}

	devs := []string{
		"",
		"dev",
		"1a06be9",           // bare git build hash (no tags)
		"0.3.0-dev+1a06be9", // local `just install` build
		"0.3.0-dev+1a06be9.dirty",
		"0.3.0-5-g1a06be9", // git describe with commits ahead of tag
		"0.3.0-5-g1a06be9-dirty",
		"0.3.0-next", // goreleaser snapshot
		"0.3",        // not three parts
		"0.3.0.1",    // too many parts
		"v0.3.x",
	}
	for _, v := range devs {
		if isReleaseVersion(v) {
			t.Errorf("isReleaseVersion(%q) = true; want false", v)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"0.3.0", "0.3.0", 0},
		{"v0.3.0", "0.3.0", 0},
		{"0.2.0", "0.3.0", -1},
		{"0.3.1", "0.3.0", 1},
		{"1.0.0", "0.9.9", 1},
		{"0.3", "0.3.0", 0}, // missing parts treated as zero
	}
	for _, c := range cases {
		got, err := compareVersions(c.a, c.b)
		if err != nil {
			t.Errorf("compareVersions(%q, %q) error: %v", c.a, c.b, err)
			continue
		}
		if got != c.want {
			t.Errorf("compareVersions(%q, %q) = %d; want %d", c.a, c.b, got, c.want)
		}
	}
}

// TestUpdateDevBuildInstalls verifies that running `update` from a development
// build (e.g. a bare git hash) no longer fails on version comparison, and
// instead offers to install the latest release. With --yes it installs without
// prompting.
func TestUpdateDevBuildInstalls(t *testing.T) {
	origFetch, origInstall, origYes, origVersion := fetchLatestVersionFunc, installLatestVersionFn, updateYes, version
	t.Cleanup(func() {
		fetchLatestVersionFunc, installLatestVersionFn, updateYes, version = origFetch, origInstall, origYes, origVersion
	})

	fetchLatestVersionFunc = func() (string, error) { return "0.3.0", nil }
	installed := false
	installLatestVersionFn = func() error { installed = true; return nil }
	updateYes = true
	version = "1a06be9" // bare build hash that previously crashed compareVersions

	updateCmd.Run(updateCmd, nil)

	if !installed {
		t.Error("expected dev build update to install the latest release")
	}
}
