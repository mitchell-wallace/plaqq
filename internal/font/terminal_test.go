package font

import "testing"

func TestTerminalRender(t *testing.T) {
	f := Terminal{}

	got := f.Render("")
	want := []string{""}
	if !equalSlices(got, want) {
		t.Errorf("Render(\"\") = %v; want %v", got, want)
	}

	got = f.Render("Hello")
	want = []string{"Hello"}
	if !equalSlices(got, want) {
		t.Errorf("Render(\"Hello\") = %v; want %v", got, want)
	}

	got = f.Render("héllo")
	want = []string{"héllo"}
	if !equalSlices(got, want) {
		t.Errorf("Render(\"héllo\") = %v; want %v", got, want)
	}

	got = f.Render("ABCdef")
	want = []string{"ABCdef"}
	if !equalSlices(got, want) {
		t.Errorf("Render(\"ABCdef\") = %v; want %v", got, want)
	}
}

func TestTerminalWidth(t *testing.T) {
	f := Terminal{}

	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"hi", 2},
		{"héllo", 5},
		{"日本語", 3},
		{"ABCdef", 6},
	}
	for _, c := range cases {
		got := f.Width(c.in)
		if got != c.want {
			t.Errorf("Width(%q) = %d; want %d", c.in, got, c.want)
		}
	}
}

func TestTerminalRegistered(t *testing.T) {
	if !Has("terminal") {
		t.Error("expected font \"terminal\" to be registered")
	}
	found := false
	for _, name := range Names() {
		if name == "terminal" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Names() = %v; expected to contain \"terminal\"", Names())
	}
}

func TestTerminalInNamesOrder(t *testing.T) {
	if !Has("terminal") {
		t.Fatal("expected font \"terminal\" to be registered")
	}
	found := false
	for _, name := range Names() {
		if name == "terminal" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Names() = %v; expected to contain \"terminal\"", Names())
	}
}
