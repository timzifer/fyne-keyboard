package keyboard

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func newTestWindow(t *testing.T) (fyne.Window, *widget.Entry) {
	t.Helper()
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })

	entry := widget.NewEntry()
	win := test.NewWindow(entry)
	win.Resize(fyne.NewSize(800, 600))
	t.Cleanup(win.Close)
	return win, entry
}

func TestDecoratorStartsHidden(t *testing.T) {
	win, _ := newTestWindow(t)
	d := NewDecorator(win, widget.NewLabel("content"))

	if d.Visible() {
		t.Error("the keyboard should start hidden")
	}
	if d.Keyboard().Visible() {
		t.Error("the keyboard widget should start hidden")
	}
	split, ok := d.Content().(*container.Split)
	if !ok {
		t.Fatalf("Content() = %T, want *container.Split", d.Content())
	}
	if split.Offset != 1 {
		t.Errorf("offset = %v, want 1 while hidden", split.Offset)
	}
}

func TestDecoratorToggle(t *testing.T) {
	win, _ := newTestWindow(t)
	d := NewDecorator(win, widget.NewLabel("content"), WithOffset(0.4))
	split := d.Content().(*container.Split)

	d.Toggle()
	if !d.Visible() || !d.Keyboard().Visible() {
		t.Fatal("Toggle() should have shown the keyboard")
	}
	if split.Offset != 0.4 {
		t.Errorf("offset = %v, want 0.4", split.Offset)
	}

	d.Toggle()
	if d.Visible() || d.Keyboard().Visible() {
		t.Error("Toggle() should have hidden the keyboard again")
	}
	if split.Offset != 1 {
		t.Errorf("offset = %v, want 1 after hiding", split.Offset)
	}
}

func TestDecoratorRemembersFocusOnShow(t *testing.T) {
	win, entry := newTestWindow(t)
	d := NewDecorator(win, entry)
	win.SetContent(d.Content())
	win.Canvas().Focus(entry)

	d.Show()
	win.Canvas().Unfocus()
	d.Keyboard().Press(CharKey('z'))

	if entry.Text != "z" {
		t.Errorf("entry text = %q, want %q", entry.Text, "z")
	}
}

func TestDecoratorHideKeyRuns(t *testing.T) {
	win, _ := newTestWindow(t)
	d := NewDecorator(win, widget.NewLabel("content"))

	d.Show()
	d.Keyboard().Press(ActionKey(ActionHide, "Hide", 1))
	if d.Visible() {
		t.Error("the hide key should close the keyboard")
	}
}

func TestDecoratorBorderMode(t *testing.T) {
	win, _ := newTestWindow(t)
	d := NewDecorator(win, widget.NewLabel("content"), WithDockMode(DockBorder))

	if _, ok := d.Content().(*fyne.Container); !ok {
		t.Fatalf("Content() = %T, want *fyne.Container", d.Content())
	}
	d.Show()
	if !d.Keyboard().Visible() {
		t.Error("border mode should still show the keyboard")
	}
	d.Hide()
	if d.Keyboard().Visible() {
		t.Error("border mode should still hide the keyboard")
	}
}

func TestDecoratorOffsetIsClamped(t *testing.T) {
	win, _ := newTestWindow(t)
	d := NewDecorator(win, widget.NewLabel("content"), WithOffset(9))
	d.Show()

	if got := d.Content().(*container.Split).Offset; got != 0.95 {
		t.Errorf("offset = %v, want the clamped 0.95", got)
	}

	d.SetOffset(-1)
	if got := d.Content().(*container.Split).Offset; got != 0.1 {
		t.Errorf("offset = %v, want the clamped 0.1", got)
	}
}

func TestDecoratorOptions(t *testing.T) {
	win, _ := newTestWindow(t)
	d := NewDecorator(win, widget.NewLabel("content"),
		WithLayoutNamed("de-DE"), WithFullKeyboard())

	if got := d.Keyboard().BaseLayout().Name; got != "de-DE" {
		t.Errorf("base layout = %q, want de-DE", got)
	}
	if got := d.Keyboard().Style(); got != StyleFull {
		t.Errorf("style = %v, want full", got)
	}

	d.SetStyle(StyleNumpad)
	if got := d.Keyboard().CurrentLayout().Name; got != "numpad" {
		t.Errorf("layout = %q, want numpad", got)
	}
	if err := d.SetLayout(QWERTY_US); err != nil {
		t.Errorf("SetLayout() = %v", err)
	}
}

func TestDecoratorAutoShowFollowsTextInput(t *testing.T) {
	win, entry := newTestWindow(t)
	d := NewDecorator(win, entry, WithAutoShow(true))
	win.SetContent(d.Content())
	// Stop the poller and drive followFocus by hand: under the test driver
	// fyne.Do runs on the calling goroutine instead of queueing onto the main
	// thread, so a live ticker would touch the canvas concurrently with the
	// test itself.
	d.StopAutoShow()

	win.Canvas().Focus(entry)
	d.followFocus()
	if !d.Visible() {
		t.Fatal("focusing an entry should show the keyboard")
	}

	win.Canvas().Unfocus()
	d.followFocus()
	if !d.Visible() {
		t.Error("a nil focus must not hide the keyboard while typing")
	}

	button := widget.NewButton("ok", func() {})
	win.SetContent(container.NewVBox(d.Content(), button))
	win.Canvas().Focus(button)
	d.followFocus()
	if d.Visible() {
		t.Error("focusing a non text widget should hide the keyboard")
	}
}

func TestDecoratorAutoShowPollerStartsAndStops(t *testing.T) {
	win, entry := newTestWindow(t)
	// An interval no test will ever reach: this checks the lifecycle of the
	// polling goroutine, not what it does.
	d := NewDecorator(win, entry, WithAutoShow(true), WithAutoShowInterval(time.Hour))

	if d.stopPoll == nil {
		t.Fatal("WithAutoShow should have started the poller")
	}
	d.startAutoShow() // a second call must not start a second goroutine
	d.StopAutoShow()
	if d.stopPoll != nil {
		t.Error("StopAutoShow should have cleared the poller")
	}
	d.StopAutoShow() // stopping twice is harmless
}

func TestDecoratorWithoutWindow(t *testing.T) {
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })

	d := NewDecorator(nil, widget.NewLabel("content"), WithAutoShow(true))
	d.Show()
	d.followFocus()
	d.Hide()
	d.StopAutoShow()
}
