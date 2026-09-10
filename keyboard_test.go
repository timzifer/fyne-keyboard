package keyboard

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// newTestKeyboard wires a keyboard to an entry inside a test window, the same
// way an application would.
func newTestKeyboard(t *testing.T, l Layout) (*Keyboard, *widget.Entry) {
	t.Helper()
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })

	entry := widget.NewEntry()
	win := test.NewWindow(entry)
	t.Cleanup(win.Close)
	win.Canvas().Focus(entry)

	return NewKeyboard(win.Canvas(), l), entry
}

func press(k *Keyboard, keys ...Key) {
	for _, key := range keys {
		k.Press(key)
	}
}

func TestTypingReachesFocusedEntry(t *testing.T) {
	k, entry := newTestKeyboard(t, QWERTY_US)

	press(k, CharKey('h'), CharKey('i'))
	if entry.Text != "hi" {
		t.Errorf("entry text = %q, want %q", entry.Text, "hi")
	}
}

func TestTypingRestoresLostFocus(t *testing.T) {
	k, entry := newTestKeyboard(t, QWERTY_US)
	k.captureFocus()
	// The driver unfocuses when a non focusable cap is tapped.
	k.canvas.Unfocus()

	press(k, CharKey('x'))
	if entry.Text != "x" {
		t.Errorf("entry text = %q, want %q", entry.Text, "x")
	}
	if k.canvas.Focused() != fyne.Focusable(entry) {
		t.Error("focus was not restored to the entry")
	}
}

func TestControlKeysAreForwarded(t *testing.T) {
	k, entry := newTestKeyboard(t, QWERTY_US)

	press(k, CharKey('a'), CharKey('b'))
	press(k, ControlKey(fyne.KeyBackspace, "⌫", 2))
	if entry.Text != "a" {
		t.Errorf("entry text = %q, want %q", entry.Text, "a")
	}
}

func TestShiftIsOneShot(t *testing.T) {
	k, entry := newTestKeyboard(t, QWERTY_US)
	shift := ModifierKey(ModShift, "⇧", 2.25)

	press(k, shift)
	if k.Level() != LevelShift {
		t.Fatalf("level = %v, want LevelShift", k.Level())
	}
	press(k, k.CurrentLayout().KeyAt(2, 1, k.Level())) // 'a' row, first letter
	press(k, k.CurrentLayout().KeyAt(2, 2, k.Level()))

	if entry.Text != "As" {
		t.Errorf("entry text = %q, want %q", entry.Text, "As")
	}
	if k.Shift() {
		t.Error("shift should have been consumed by the first character")
	}
}

func TestCapsLockLatches(t *testing.T) {
	k, entry := newTestKeyboard(t, QWERTY_US)
	caps := ModifierKey(ModCapsLock, "Caps", 1.75)

	press(k, caps)
	press(k, k.CurrentLayout().KeyAt(2, 1, k.Level()), k.CurrentLayout().KeyAt(2, 2, k.Level()))
	if entry.Text != "AS" {
		t.Errorf("entry text = %q, want %q", entry.Text, "AS")
	}
	if !k.CapsLock() {
		t.Error("caps lock should stay latched")
	}

	press(k, caps)
	press(k, k.CurrentLayout().KeyAt(2, 1, k.Level()))
	if entry.Text != "ASa" {
		t.Errorf("entry text = %q, want %q", entry.Text, "ASa")
	}
}

func TestShiftAndCapsCancelOut(t *testing.T) {
	k, entry := newTestKeyboard(t, QWERTY_US)

	press(k, ModifierKey(ModCapsLock, "Caps", 1.75), ModifierKey(ModShift, "⇧", 2.25))
	if k.Level() != LevelBase {
		t.Fatalf("level = %v, want LevelBase", k.Level())
	}
	press(k, k.CurrentLayout().KeyAt(2, 1, k.Level()))
	if entry.Text != "a" {
		t.Errorf("entry text = %q, want %q", entry.Text, "a")
	}
}

func TestAltGrLevel(t *testing.T) {
	k, entry := newTestKeyboard(t, QWERTZ_DE)

	press(k, ModifierKey(ModAltGr, "AltGr", 1.5))
	if k.Level() != LevelAltGr {
		t.Fatalf("level = %v, want LevelAltGr", k.Level())
	}
	// Row 1 column 3 is 'e' in QWERTZ, which is € on the AltGr level.
	press(k, k.CurrentLayout().KeyAt(1, 3, k.Level()))
	if entry.Text != "€" {
		t.Errorf("entry text = %q, want %q", entry.Text, "€")
	}
	if k.AltGr() {
		t.Error("AltGr should be released after a character")
	}
}

func TestDeadKeyCombines(t *testing.T) {
	k, entry := newTestKeyboard(t, QWERTZ_DE)
	circumflex := DeadKey('^', "")

	press(k, circumflex)
	if k.Dead() != '^' {
		t.Fatalf("dead = %q, want '^'", k.Dead())
	}
	if entry.Text != "" {
		t.Errorf("a dead key must not type anything yet, got %q", entry.Text)
	}

	press(k, CharKey('a'))
	if entry.Text != "â" {
		t.Errorf("entry text = %q, want %q", entry.Text, "â")
	}
	if k.Dead() != 0 {
		t.Error("dead key should be cleared after combining")
	}
}

func TestDeadKeyWithoutCombinationEmitsBoth(t *testing.T) {
	k, entry := newTestKeyboard(t, QWERTZ_DE)

	press(k, DeadKey('^', ""), CharKey('q'))
	if entry.Text != "^q" {
		t.Errorf("entry text = %q, want %q", entry.Text, "^q")
	}
}

func TestDeadKeyFlushedByControlKey(t *testing.T) {
	k, entry := newTestKeyboard(t, QWERTZ_DE)

	press(k, DeadKey('^', ""), ControlKey(fyne.KeyReturn, "⏎", 2.25))
	if entry.Text != "^" {
		t.Errorf("entry text = %q, want %q", entry.Text, "^")
	}
}

func TestSpacerIsInert(t *testing.T) {
	k, entry := newTestKeyboard(t, QWERTY_US)
	called := false
	k.OnKey = func(Key) { called = true }

	press(k, SpacerKey(0.5))
	if entry.Text != "" || called {
		t.Error("a spacer must not type anything or fire OnKey")
	}
}

func TestActionsRunCallbacks(t *testing.T) {
	k, _ := newTestKeyboard(t, QWERTY_US)
	hidden := false
	k.OnHide = func() { hidden = true }

	press(k, ActionKey(ActionHide, "Hide", 2.75))
	if !hidden {
		t.Error("ActionHide should call OnHide")
	}

	changed := ""
	k.OnLayoutChange = func(l Layout) { changed = l.Name }
	press(k, ActionKey(ActionNextLayout, "🌐", 1.75))
	if changed == "" || changed == "en-US" {
		t.Errorf("layout switch did not move on, now %q", k.CurrentLayout().Name)
	}
}

func TestSetLayoutRejectsInvalid(t *testing.T) {
	k, _ := newTestKeyboard(t, QWERTY_US)

	if err := k.SetLayout(Layout{Name: "broken"}); err == nil {
		t.Fatal("expected an error for a layout without rows")
	}
	if k.CurrentLayout().Name != "en-US" {
		t.Errorf("layout changed despite the error, now %q", k.CurrentLayout().Name)
	}

	if err := k.SetLayout(QWERTZ_DE); err != nil {
		t.Fatalf("SetLayout(QWERTZ_DE) = %v", err)
	}
	if k.CurrentLayout().Name != "de-DE" {
		t.Errorf("layout = %q, want de-DE", k.CurrentLayout().Name)
	}
}

func TestKeyboardWithoutTargetDoesNotPanic(t *testing.T) {
	k := NewKeyboard(nil, QWERTY_US)
	press(k, CharKey('a'), ControlKey(fyne.KeyReturn, "⏎", 1), DeadKey('^', ""))
}

func TestCapsReflectLevelChange(t *testing.T) {
	k, _ := newTestKeyboard(t, QWERTY_US)

	before := k.caps2d[2][1].key.Rune
	press(k, ModifierKey(ModShift, "⇧", 2.25))
	after := k.caps2d[2][1].key.Rune

	if before != 'a' || after != 'A' {
		t.Errorf("cap label did not follow the level: %q -> %q", before, after)
	}
	if !k.caps2d[3][0].active {
		t.Error("the shift cap should render as active")
	}
}

func TestRendererBuildsForEveryLayout(t *testing.T) {
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })

	for _, l := range Layouts() {
		k := NewKeyboard(nil, l)
		win := test.NewWindow(k)
		win.Resize(fyne.NewSize(800, 300))
		win.Close()
	}
}
