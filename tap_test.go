package keyboard

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// These tests drive the keyboard the way a user does: by tapping the caps that
// are actually on screen, after the widget has been rendered.

func renderedKeyboard(t *testing.T) (*Decorator, *widget.Entry, fyne.Window) {
	t.Helper()
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })

	entry := widget.NewEntry()
	win := test.NewWindow(nil)
	t.Cleanup(win.Close)

	d := NewDecorator(win, entry)
	win.SetContent(d.Content())
	win.Resize(fyne.NewSize(900, 600))
	win.Canvas().Focus(entry)
	d.Show()

	return d, entry, win
}

// capFor finds the cap that currently shows the given label.
func capFor(t *testing.T, k *Keyboard, label string) *keyCap {
	t.Helper()
	for _, row := range k.caps2d {
		for _, kc := range row {
			if kc.key.Text() == label {
				return kc
			}
		}
	}
	t.Fatalf("no cap labelled %q", label)
	return nil
}

func tapCap(t *testing.T, k *Keyboard, label string) {
	t.Helper()
	test.Tap(capFor(t, k, label))
}

func TestTappingShiftTypesUpperCase(t *testing.T) {
	d, entry, _ := renderedKeyboard(t)
	k := d.Keyboard()

	tapCap(t, k, "⇧")
	if !k.Shift() {
		t.Fatal("tapping shift did not arm the shift level")
	}
	if got := capFor(t, k, "A"); got == nil {
		t.Fatal("the caps did not switch to the shift level")
	}

	tapCap(t, k, "A")
	tapCap(t, k, "b")
	if entry.Text != "Ab" {
		t.Errorf("entry text = %q, want %q", entry.Text, "Ab")
	}
}

func TestTappingCapsLockLatches(t *testing.T) {
	d, entry, _ := renderedKeyboard(t)
	k := d.Keyboard()

	tapCap(t, k, "Caps")
	tapCap(t, k, "A")
	tapCap(t, k, "B")
	if entry.Text != "AB" {
		t.Errorf("entry text = %q, want %q", entry.Text, "AB")
	}

	tapCap(t, k, "Caps")
	tapCap(t, k, "c")
	if entry.Text != "ABc" {
		t.Errorf("entry text = %q, want %q", entry.Text, "ABc")
	}
}

// Rebuilding the keyboard must reuse the container the renderer holds,
// otherwise the previous keyboard stays on screen and its caps keep answering
// taps while the state changes somewhere invisible.
func TestStyleChangeAfterRenderReachesTheScreen(t *testing.T) {
	d, entry, _ := renderedKeyboard(t)
	k := d.Keyboard()
	grid := k.grid

	d.SetStyle(StyleNumpad)

	if k.grid != grid {
		t.Error("the grid container was replaced, the renderer still shows the old one")
	}
	if len(k.caps2d) != len(k.CurrentLayout().Rows) {
		t.Errorf("caps grid has %d rows, layout has %d", len(k.caps2d), len(k.CurrentLayout().Rows))
	}
	if !containsCap(k, "00") {
		t.Fatal("the keypad caps are not on the keyboard")
	}

	tapCap(t, k, "7")
	tapCap(t, k, "00")
	if entry.Text != "700" {
		t.Errorf("entry text = %q, want %q", entry.Text, "700")
	}
}

func TestLayoutChangeAfterRenderReachesTheScreen(t *testing.T) {
	d, entry, _ := renderedKeyboard(t)
	k := d.Keyboard()

	if err := d.SetLayout(QWERTZ_DE); err != nil {
		t.Fatalf("SetLayout() = %v", err)
	}
	if !containsCap(k, "ü") {
		t.Fatal("the German caps are not on the keyboard")
	}

	tapCap(t, k, "z")
	tapCap(t, k, "ü")
	if entry.Text != "zü" {
		t.Errorf("entry text = %q, want %q", entry.Text, "zü")
	}

	// The caps of the previous layout must be gone from the container.
	if countCapObjects(k) != countCaps(k) {
		t.Errorf("the grid holds %d caps, the layout has %d", countCapObjects(k), countCaps(k))
	}
}

func TestTappingDeadKeyThenLetter(t *testing.T) {
	d, entry, _ := renderedKeyboard(t)
	k := d.Keyboard()
	if err := d.SetLayout(QWERTZ_DE); err != nil {
		t.Fatalf("SetLayout() = %v", err)
	}

	tapCap(t, k, "^")
	tapCap(t, k, "a")
	if entry.Text != "â" {
		t.Errorf("entry text = %q, want %q", entry.Text, "â")
	}
}

func TestTappingRestoresFocusAfterTheDriverUnfocuses(t *testing.T) {
	d, entry, win := renderedKeyboard(t)
	k := d.Keyboard()

	kc := capFor(t, k, "a")
	kc.MouseDown(nil)      // the driver reports the press ...
	win.Canvas().Unfocus() // ... and then unfocuses the entry
	kc.MouseUp(nil)
	test.Tap(kc)

	if entry.Text != "a" {
		t.Errorf("entry text = %q, want %q", entry.Text, "a")
	}
}

func containsCap(k *Keyboard, label string) bool {
	for _, row := range k.caps2d {
		for _, kc := range row {
			if kc.key.Text() == label {
				return true
			}
		}
	}
	return false
}

func countCaps(k *Keyboard) int {
	n := 0
	for _, row := range k.caps2d {
		n += len(row)
	}
	return n
}

func countCapObjects(k *Keyboard) int {
	n := 0
	for _, row := range k.grid.Objects {
		if c, ok := row.(*fyne.Container); ok {
			n += len(c.Objects)
		}
	}
	return n
}

func TestKeyboardStaysUsableInsideASplit(t *testing.T) {
	d, _, _ := renderedKeyboard(t)

	split, ok := d.Content().(*container.Split)
	if !ok {
		t.Fatalf("Content() = %T", d.Content())
	}
	if split.Trailing != fyne.CanvasObject(d.Keyboard()) {
		t.Error("the keyboard is not the trailing half of the split")
	}
	if !d.Keyboard().Visible() {
		t.Error("the keyboard should be visible after Show()")
	}
}
