package border

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestKindString(t *testing.T) {
	cases := []struct {
		kind Kind
		want string
	}{
		{None, "none"},
		{Single, "single"},
		{Double, "double"},
		{Block, "block"},
	}
	for _, c := range cases {
		if got := c.kind.String(); got != c.want {
			t.Errorf("%v.String() = %q; want %q", c.kind, got, c.want)
		}
	}
}

func TestHas(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"none", true},
		{"NONE", true},
		{"Single", true},
		{"SINGLE", true},
		{" block ", true},
		{"", true},
		{"\tDOUBLE\n", true},
		{"nope", false},
		{"singl", false},
		{"  ", true},
	}
	for _, c := range cases {
		if got := Has(c.in); got != c.want {
			t.Errorf("Has(%q) = %v; want %v", c.in, got, c.want)
		}
	}
}

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want Kind
	}{
		{"none", None},
		{"SINGLE", Single},
		{" block ", Block},
		{"DoUbLe", Double},
		{"", None},
		{"\tNone\n", None},
	}
	for _, c := range cases {
		got, err := Parse(c.in)
		if err != nil {
			t.Errorf("Parse(%q) returned error %v; want kind %v", c.in, err, c.want)
			continue
		}
		if got != c.want {
			t.Errorf("Parse(%q) = %v; want %v", c.in, got, c.want)
		}
	}
}

func TestParseErrorMentionsValidNames(t *testing.T) {
	_, err := Parse("nope")
	if err == nil {
		t.Fatal("Parse(\"nope\") returned nil error; want error")
	}
	msg := err.Error()
	for _, name := range []string{"none", "single", "double", "block"} {
		if !strings.Contains(msg, name) {
			t.Errorf("Parse(\"nope\") error %q does not mention valid name %q", msg, name)
		}
	}
	if !strings.Contains(msg, "unknown frame") {
		t.Errorf("Parse(\"nope\") error %q does not match expected shape", msg)
	}
}

func TestParseSuggestsNearMiss(t *testing.T) {
	cases := []struct {
		in      string
		suggest string
	}{
		{"singl", "single"},
		{"blok", "block"},
		{"nonee", "none"},
		{"doubl", "double"},
	}
	for _, c := range cases {
		_, err := Parse(c.in)
		if err == nil {
			t.Errorf("Parse(%q) returned nil error; want error with suggestion", c.in)
			continue
		}
		want := "did you mean \"" + c.suggest + "\"?"
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Parse(%q) error %q does not contain suggestion %q", c.in, err.Error(), want)
		}
	}
}

func TestParseNoSuggestionForDistantMiss(t *testing.T) {
	for _, in := range []string{"xyzqqq", "qqqqqq", "zxyw"} {
		_, err := Parse(in)
		if err == nil {
			t.Errorf("Parse(%q) returned nil error; want error", in)
			continue
		}
		if strings.Contains(err.Error(), "did you mean") {
			t.Errorf("Parse(%q) error %q unexpectedly contains a suggestion", in, err.Error())
		}
	}
}

func TestNames(t *testing.T) {
	got := Names()
	want := []string{"none", "single", "double", "block"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Names() = %v; want %v", got, want)
	}
}

func TestNoneWrapNoOp(t *testing.T) {
	in := []string{"a", "b", "c"}
	got := None.Wrap(in, 1)
	if !reflect.DeepEqual(got, in) {
		t.Errorf("None.Wrap(...) = %v; want %v (unchanged)", got, in)
	}
}

func TestWrapEmptyRowsNoOp(t *testing.T) {
	for _, k := range []Kind{Single, Double, Block} {
		got := k.Wrap(nil, 1)
		if got != nil {
			t.Errorf("%s.Wrap(nil) = %v; want nil", k, got)
		}
		got = k.Wrap([]string{}, 1)
		if len(got) != 0 {
			t.Errorf("%s.Wrap([]) = %v; want empty", k, got)
		}
	}
}

func TestSingleWrapSimple(t *testing.T) {
	got := Single.Wrap([]string{"hi"}, 1)
	want := []string{"┌────┐", "│ hi │", "└────┘"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Single.Wrap([hi], 1) = %q; want %q", got, want)
	}
	for i, row := range got {
		if utf8.RuneCountInString(row) != 6 {
			t.Errorf("row %d has %d runes; want 6", i, utf8.RuneCountInString(row))
		}
	}
}

func TestSingleWrapMultiline(t *testing.T) {
	got := Single.Wrap([]string{"ab", "cd"}, 1)
	want := []string{"┌────┐", "│ ab │", "│ cd │", "└────┘"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Single.Wrap([ab,cd], 1) = %q; want %q", got, want)
	}
	for i, row := range got {
		if utf8.RuneCountInString(row) != 6 {
			t.Errorf("row %d has %d runes; want 6", i, utf8.RuneCountInString(row))
		}
	}
}

func TestSingleWrapBlankRowInContent(t *testing.T) {
	got := Single.Wrap([]string{"ab", "", "cd"}, 1)
	want := []string{"┌────┐", "│ ab │", "│    │", "│ cd │", "└────┘"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Single.Wrap([ab,'',cd], 1) = %q; want %q", got, want)
	}
	for i, row := range got {
		if utf8.RuneCountInString(row) != 6 {
			t.Errorf("row %d has %d runes; want 6 (rectangular frame)", i, utf8.RuneCountInString(row))
		}
	}
}

func TestSingleWrapMultiMargin(t *testing.T) {
	got := Single.Wrap([]string{"ab"}, 3)
	want := []string{"┌────────┐", "│   ab   │", "└────────┘"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Single.Wrap([ab], 3) = %q; want %q", got, want)
	}
	for i, row := range got {
		if utf8.RuneCountInString(row) != 10 {
			t.Errorf("row %d has %d runes; want 10", i, utf8.RuneCountInString(row))
		}
	}
}

func TestDoubleWrap(t *testing.T) {
	got := Double.Wrap([]string{"ab", "cd"}, 1)
	want := []string{"╔════╗", "║ ab ║", "║ cd ║", "╚════╝"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Double.Wrap([ab,cd], 1) = %q; want %q", got, want)
	}
	for i, row := range got {
		if utf8.RuneCountInString(row) != 6 {
			t.Errorf("row %d has %d runes; want 6", i, utf8.RuneCountInString(row))
		}
	}
}

func TestBlockWrap(t *testing.T) {
	got := Block.Wrap([]string{"ab"}, 1)
	want := []string{"██████", "█ ab █", "██████"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Block.Wrap([ab], 1) = %q; want %q", got, want)
	}
	for i, row := range got {
		if utf8.RuneCountInString(row) != 6 {
			t.Errorf("row %d has %d runes; want 6", i, utf8.RuneCountInString(row))
		}
	}
}

func TestBlockWrapAllBlank(t *testing.T) {
	got := Block.Wrap([]string{"", ""}, 1)
	want := []string{"████", "█  █", "█  █", "████"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Block.Wrap(['',''], 1) = %q; want %q", got, want)
	}
	for i, row := range got {
		if utf8.RuneCountInString(row) != 4 {
			t.Errorf("row %d has %d runes; want 4", i, utf8.RuneCountInString(row))
		}
	}
}

func TestWrapRespectsRuneWidth(t *testing.T) {
	got := Single.Wrap([]string{"héllo"}, 1)
	want := []string{
		"┌───────┐",
		"│ héllo │",
		"└───────┘",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Single.Wrap([héllo], 1) = %q; want %q", got, want)
	}
	for i, row := range got {
		if w := utf8.RuneCountInString(row); w != 9 {
			t.Errorf("row %d has %d runes; want 9 (rune-counted, not byte-counted)", i, w)
		}
	}
}

func TestWrapEmptyRowsAllBlank(t *testing.T) {
	got := Single.Wrap([]string{"", ""}, 1)
	want := []string{"┌──┐", "│  │", "│  │", "└──┘"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Single.Wrap(['',''], 1) = %q; want %q", got, want)
	}
	for i, row := range got {
		if utf8.RuneCountInString(row) != 4 {
			t.Errorf("row %d has %d runes; want 4", i, utf8.RuneCountInString(row))
		}
	}
}

func TestWrapRectangularInvariant(t *testing.T) {
	rows := []string{"ab", "", "cd", " ", "ef"}
	for _, k := range []Kind{Single, Double, Block} {
		got := k.Wrap(rows, 2)
		width := utf8.RuneCountInString(got[0])
		for i, row := range got {
			if w := utf8.RuneCountInString(row); w != width {
				t.Errorf("%s.Wrap: row %d has %d runes; want %d (rectangular)", k, i, w, width)
			}
		}
	}
}

func TestWrapZeroMargin(t *testing.T) {
	got := Single.Wrap([]string{"abc", "de"}, 0)
	want := []string{"┌───┐", "│abc│", "│de │", "└───┘"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Single.Wrap([abc,de], 0) = %q; want %q", got, want)
	}
}
