package keyboard

import (
	"testing"

	"fyne.io/fyne/v2"
)

func TestKeyUnitsAndText(t *testing.T) {
	if got := (Key{}).Units(); got != 1 {
		t.Errorf("zero width should default to 1 unit, got %v", got)
	}
	if got := CharKeyWidth('a', 2.5).Units(); got != 2.5 {
		t.Errorf("Units() = %v, want 2.5", got)
	}

	cases := []struct {
		name string
		key  Key
		want string
	}{
		{"rune", CharKey('q'), "q"},
		{"label wins", Key{Rune: ' ', Label: "Space"}, "Space"},
		{"key name", ControlKey(fyne.KeyBackspace, "", 1), string(fyne.KeyBackspace)},
		{"empty", Key{}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.key.Text(); got != tc.want {
				t.Errorf("Text() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestKeyAtLevels(t *testing.T) {
	l := Layout{
		Name: "test",
		Rows: []Row{{
			CharKey('a'),
			CharKey('1'),
			ControlKey(fyne.KeyBackspace, "⌫", 2),
			CharKey('e'),
		}},
		Shift: []Row{{
			Key{},        // no override: falls back to upper case
			CharKey('!'), // explicit override
			Key{},        // ignored, base is a control key
			Key{},        // upper case fallback again
		}},
		AltGr: []Row{{
			Key{},
			Key{},
			Key{},
			CharKey('€'),
		}},
	}

	cases := []struct {
		name  string
		col   int
		level Level
		want  rune
	}{
		{"base letter", 0, LevelBase, 'a'},
		{"shift letter derived", 0, LevelShift, 'A'},
		{"shift digit override", 1, LevelShift, '!'},
		{"altgr falls back to base", 1, LevelAltGr, '1'},
		{"altgr override", 3, LevelAltGr, '€'},
		{"shift keeps letter case rule", 3, LevelShift, 'E'},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := l.KeyAt(0, tc.col, tc.level).Rune; got != tc.want {
				t.Errorf("KeyAt(0, %d, %v).Rune = %q, want %q", tc.col, tc.level, got, tc.want)
			}
		})
	}

	if got := l.KeyAt(0, 2, LevelShift); got.Key != fyne.KeyBackspace {
		t.Errorf("control keys must ignore the shift level, got %+v", got)
	}
	if got := l.KeyAt(9, 9, LevelBase); got != (Key{}) {
		t.Errorf("out of range lookup should return the zero key, got %+v", got)
	}
}

func TestKeyAtKeepsBaseWidth(t *testing.T) {
	l := Layout{
		Name:  "test",
		Rows:  []Row{{CharKeyWidth('a', 2)}},
		Shift: []Row{{CharKey('A')}},
	}
	if got := l.KeyAt(0, 0, LevelShift).Units(); got != 2 {
		t.Errorf("overlay key should keep the base width, got %v", got)
	}
}

func TestLayoutValidate(t *testing.T) {
	base := []Row{{CharKey('a'), CharKey('b')}}

	cases := []struct {
		name    string
		layout  Layout
		wantErr bool
	}{
		{"ok", Layout{Name: "x", Rows: base}, false},
		{"no name", Layout{Rows: base}, true},
		{"no rows", Layout{Name: "x"}, true},
		{"empty row", Layout{Name: "x", Rows: []Row{{}}}, true},
		{"shift row count", Layout{Name: "x", Rows: base, Shift: []Row{{}, {}}}, true},
		{"shift key count", Layout{Name: "x", Rows: base, Shift: []Row{{CharKey('A')}}}, true},
		{"altgr key count", Layout{Name: "x", Rows: base, AltGr: []Row{{CharKey('@')}}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.layout.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestLayoutCloneIsDeep(t *testing.T) {
	original := Layout{
		Name:     "x",
		Rows:     []Row{{CharKey('a')}},
		DeadKeys: map[rune]map[rune]rune{'^': {'a': 'â'}},
	}
	clone := original.Clone()
	clone.Rows[0][0] = CharKey('z')
	clone.DeadKeys['^']['a'] = 'x'

	if original.Rows[0][0].Rune != 'a' {
		t.Error("clone shares the rows slice with the original")
	}
	if original.DeadKeys['^']['a'] != 'â' {
		t.Error("clone shares the dead key table with the original")
	}
}

func TestLayoutCombine(t *testing.T) {
	l := QWERTZ_DE
	if got, ok := l.Combine('^', 'a'); !ok || got != 'â' {
		t.Errorf("Combine('^','a') = %q, %v; want 'â', true", got, ok)
	}
	if _, ok := l.Combine('^', 'q'); ok {
		t.Error("^ should not combine with q")
	}
	if _, ok := l.Combine('%', 'a'); ok {
		t.Error("% is not a dead key")
	}
}
