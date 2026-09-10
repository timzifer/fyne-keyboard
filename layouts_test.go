package keyboard

import (
	"math"
	"testing"
)

func TestBuiltInLayoutsAreValid(t *testing.T) {
	for _, l := range Layouts() {
		t.Run(l.Name, func(t *testing.T) {
			if err := l.Validate(); err != nil {
				t.Errorf("Validate() = %v", err)
			}
			if l.Title == "" {
				t.Error("layout has no title")
			}
		})
	}
}

// Every row is scaled to the full keyboard width on its own, so rows only line
// up when their unit totals match.
func TestBuiltInLayoutRowsHaveEqualWidth(t *testing.T) {
	for _, l := range Layouts() {
		t.Run(l.Name, func(t *testing.T) {
			for i, row := range l.Rows {
				var sum float32
				for _, key := range row {
					sum += key.Units()
				}
				if math.Abs(float64(sum-rowUnits)) > 0.001 {
					t.Errorf("row %d sums to %v units, want %v", i, sum, rowUnits)
				}
			}
		})
	}
}

func TestBuiltInLayoutsHaveTheEssentialKeys(t *testing.T) {
	for _, l := range Layouts() {
		t.Run(l.Name, func(t *testing.T) {
			var space, shift, hide bool
			for _, row := range l.Rows {
				for _, key := range row {
					switch {
					case key.Type == KeyChar && key.Rune == ' ':
						space = true
					case key.Type == KeyModifier && key.Mod == ModShift:
						shift = true
					case key.Type == KeyAction && key.Action == ActionHide:
						hide = true
					}
				}
			}
			if !space || !shift || !hide {
				t.Errorf("missing keys: space=%v shift=%v hide=%v", space, shift, hide)
			}
		})
	}
}

func TestOverlayLevelKeepsShapeAndMarksDeadKeys(t *testing.T) {
	deadKeys := map[rune]map[rune]rune{'`': {'a': 'à'}}
	rows := []Row{{CharKey('a'), CharKey('´'), ControlKey("Tab", "Tab", 1.5)}}

	level := overlayLevel(rows, map[rune]rune{'´': '`'}, deadKeys)

	if len(level) != 1 || len(level[0]) != 3 {
		t.Fatalf("overlay shape = %d rows, first row %d keys", len(level), len(level[0]))
	}
	if level[0][0] != (Key{}) {
		t.Errorf("unmapped keys should stay empty, got %+v", level[0][0])
	}
	if got := level[0][1]; got.Type != KeyDead || got.Rune != '`' {
		t.Errorf("mapped dead key = %+v, want a dead '`'", got)
	}
	if level[0][2] != (Key{}) {
		t.Errorf("control keys have no overlay, got %+v", level[0][2])
	}
}

func TestAccents(t *testing.T) {
	table := accents("ae", "âê")
	if table['a'] != 'â' || table['e'] != 'ê' || len(table) != 2 {
		t.Errorf("accents() = %v", table)
	}
	if got := accents("abc", "â"); len(got) != 1 {
		t.Errorf("a short replacement string should stop early, got %v", got)
	}
}

func TestRowConcatenation(t *testing.T) {
	got := row(SpacerKey(0.5), "ab", Row{CharKey('c')})
	want := Row{SpacerKey(0.5), CharKey('a'), CharKey('b'), CharKey('c')}

	if len(got) != len(want) {
		t.Fatalf("row() = %d keys, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("key %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestRowRejectsUnsupportedParts(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("row() should panic on an unsupported part")
		}
	}()
	row(42)
}

func TestGermanLevels(t *testing.T) {
	l := QWERTZ_DE

	cases := []struct {
		name        string
		row, col    int
		level       Level
		wantRune    rune
		wantKeyType KeyType
	}{
		{"digit base", 0, 1, LevelBase, '1', KeyChar},
		{"digit shift", 0, 1, LevelShift, '!', KeyChar},
		{"digit altgr", 0, 1, LevelAltGr, '¹', KeyChar},
		{"eszett shift", 0, 11, LevelShift, '?', KeyChar},
		{"circumflex dead", 0, 0, LevelBase, '^', KeyDead},
		{"acute shifts to grave", 0, 12, LevelShift, '`', KeyDead},
		{"umlaut upper", 1, 11, LevelShift, 'Ü', KeyChar},
		{"euro on altgr", 1, 3, LevelAltGr, '€', KeyChar},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := l.KeyAt(tc.row, tc.col, tc.level)
			if got.Rune != tc.wantRune || got.Type != tc.wantKeyType {
				t.Errorf("KeyAt(%d,%d,%v) = %q/%v, want %q/%v",
					tc.row, tc.col, tc.level, got.Rune, got.Type, tc.wantRune, tc.wantKeyType)
			}
		})
	}
}

func TestFrenchDigitsNeedShift(t *testing.T) {
	l := AZERTY_FR

	if got := l.KeyAt(0, 1, LevelBase).Rune; got != '&' {
		t.Errorf("base = %q, want '&'", got)
	}
	if got := l.KeyAt(0, 1, LevelShift).Rune; got != '1' {
		t.Errorf("shift = %q, want '1'", got)
	}
}
