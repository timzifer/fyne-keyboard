package keyboard

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestNumpadShape(t *testing.T) {
	l := Numpad()

	if err := l.Validate(); err != nil {
		t.Fatalf("Validate() = %v", err)
	}
	for i, row := range l.Rows {
		var sum float32
		for _, key := range row {
			sum += key.Units()
		}
		if sum != numpadUnits {
			t.Errorf("row %d sums to %v units, want %v", i, sum, numpadUnits)
		}
	}
	if _, ok := LookupLayout("numpad"); ok {
		t.Error("the keypad is a style, it must not be in the language registry")
	}
}

func TestNumpadDoubleZeroTypesTwoRunes(t *testing.T) {
	k, entry := newTestKeyboard(t, Numpad())

	var doubleZero Key
	for _, row := range k.CurrentLayout().Rows {
		for _, key := range row {
			if key.Runes == "00" {
				doubleZero = key
			}
		}
	}
	if doubleZero.Runes != "00" {
		t.Fatal("the keypad has no double zero key")
	}

	press(k, CharKey('1'), doubleZero)
	if entry.Text != "100" {
		t.Errorf("entry text = %q, want %q", entry.Text, "100")
	}
}

func TestNumpadHasNoUpperLevel(t *testing.T) {
	l := Numpad()
	if got := l.KeyAt(1, 0, LevelShift).Rune; got != '7' {
		t.Errorf("digits must not change on shift, got %q", got)
	}
}

func TestFullAttachesTheKeypad(t *testing.T) {
	full := Full(QWERTZ_DE)

	if err := full.Validate(); err != nil {
		t.Fatalf("Validate() = %v", err)
	}
	if len(full.Rows) != len(QWERTZ_DE.Rows) {
		t.Fatalf("full layout has %d rows, want %d", len(full.Rows), len(QWERTZ_DE.Rows))
	}

	var want float32 = rowUnits + fullGapUnits + numpadUnits
	for i, row := range full.Rows {
		var sum float32
		for _, key := range row {
			sum += key.Units()
		}
		if sum != want {
			t.Errorf("row %d sums to %v units, want %v", i, sum, want)
		}
	}

	// The character block keeps its levels and dead keys.
	if got := full.KeyAt(0, 1, LevelShift).Rune; got != '!' {
		t.Errorf("shift level lost: got %q, want '!'", got)
	}
	if got, ok := full.Combine('^', 'a'); !ok || got != 'â' {
		t.Errorf("dead keys lost: %q, %v", got, ok)
	}
	// And it has exactly one hide key, from the character block.
	hides := 0
	for _, row := range full.Rows {
		for _, key := range row {
			if key.Action == ActionHide {
				hides++
			}
		}
	}
	if hides != 1 {
		t.Errorf("full layout has %d hide keys, want 1", hides)
	}
}

func TestComposeRejectsMismatchedRows(t *testing.T) {
	left := Layout{Name: "l", Rows: []Row{{CharKey('a')}}}
	right := Layout{Name: "r", Rows: []Row{{CharKey('b')}, {CharKey('c')}}}

	if _, err := Compose(left, right, 1); err == nil {
		t.Error("expected an error for mismatched row counts")
	}
}

func TestComposeMergesLevelsAndTitles(t *testing.T) {
	left := Layout{
		Name: "l", Title: "Left",
		Rows:  []Row{{CharKey('a')}},
		Shift: []Row{{CharKey('A')}},
	}
	right := Layout{Name: "r", Title: "Right", Rows: []Row{{CharKey('1')}}}

	out, err := Compose(left, right, 0.5)
	if err != nil {
		t.Fatalf("Compose() = %v", err)
	}
	if out.Name != "l+r" || out.Title != "Left + Right" {
		t.Errorf("name/title = %q/%q", out.Name, out.Title)
	}
	if len(out.Shift[0]) != len(out.Rows[0]) {
		t.Errorf("shift row has %d keys, base has %d", len(out.Shift[0]), len(out.Rows[0]))
	}
	if got := out.Rows[0][1]; got.Type != KeySpacer || got.Units() != 0.5 {
		t.Errorf("separator = %+v, want a 0.5 unit spacer", got)
	}
	if got := out.KeyAt(0, 2, LevelShift).Rune; got != '1' {
		t.Errorf("the right block should be unaffected by shift, got %q", got)
	}
}

func TestStyleApply(t *testing.T) {
	if got := StyleText.apply(QWERTY_US).Name; got != "en-US" {
		t.Errorf("StyleText = %q, want en-US", got)
	}
	if got := StyleNumpad.apply(QWERTY_US).Name; got != "numpad" {
		t.Errorf("StyleNumpad = %q, want numpad", got)
	}
	if got := StyleFull.apply(QWERTY_US).Name; got != "en-US+numpad" {
		t.Errorf("StyleFull = %q, want en-US+numpad", got)
	}
	if got := StyleFull.String(); got != "full" {
		t.Errorf("String() = %q, want full", got)
	}
}

func TestFullFallsBackForMismatchedLayouts(t *testing.T) {
	odd := Layout{Name: "odd", Rows: []Row{{CharKey('a')}}}
	if got := Full(odd); got.Name != "odd" {
		t.Errorf("Full() = %q, want the unchanged layout when the keypad does not fit", got.Name)
	}
}

func TestKeyboardStyleSwitching(t *testing.T) {
	k, _ := newTestKeyboard(t, QWERTZ_DE)

	k.SetStyle(StyleNumpad)
	if k.CurrentLayout().Name != "numpad" || k.BaseLayout().Name != "de-DE" {
		t.Errorf("current/base = %q/%q", k.CurrentLayout().Name, k.BaseLayout().Name)
	}

	// Switching the language while in keypad style keeps the style.
	if err := k.SetLayout(QWERTY_US); err != nil {
		t.Fatalf("SetLayout() = %v", err)
	}
	if k.CurrentLayout().Name != "numpad" || k.BaseLayout().Name != "en-US" {
		t.Errorf("current/base = %q/%q", k.CurrentLayout().Name, k.BaseLayout().Name)
	}

	k.SetStyle(StyleText)
	if k.CurrentLayout().Name != "en-US" {
		t.Errorf("current = %q, want en-US", k.CurrentLayout().Name)
	}
}

func TestNumpadStyleRendersInAWindow(t *testing.T) {
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })

	for _, style := range []Style{StyleText, StyleNumpad, StyleFull} {
		k := NewKeyboardWithStyle(nil, QWERTZ_DE, style)
		win := test.NewWindow(k)
		win.Resize(fyne.NewSize(1000, 320))
		win.Close()
	}
}
