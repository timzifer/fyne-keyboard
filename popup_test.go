package keyboard

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestPopupShowForPinsTheTarget(t *testing.T) {
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })

	entry := widget.NewEntry()
	win := test.NewWindow(container.NewVBox(entry))
	win.Resize(fyne.NewSize(800, 600))
	t.Cleanup(win.Close)

	p := NewPopup(win)
	p.ShowFor(entry)

	if !p.Visible() {
		t.Fatal("ShowFor() should open the popup")
	}
	if p.Anchor() != fyne.CanvasObject(entry) {
		t.Error("the anchor was not remembered")
	}
	if p.Keyboard().Target() != fyne.Focusable(entry) {
		t.Error("the entry should be pinned as the keyboard target")
	}

	p.Keyboard().Press(CharKey('q'))
	if entry.Text != "q" {
		t.Errorf("entry text = %q, want %q", entry.Text, "q")
	}

	p.Hide()
	if p.Visible() {
		t.Error("Hide() should close the popup")
	}
}

func TestPopupToggleFor(t *testing.T) {
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })

	first := widget.NewEntry()
	second := widget.NewEntry()
	win := test.NewWindow(container.NewVBox(first, second))
	win.Resize(fyne.NewSize(800, 600))
	t.Cleanup(win.Close)

	p := NewPopup(win)
	p.ToggleFor(first)
	if !p.Visible() {
		t.Fatal("the first toggle should open the popup")
	}

	p.ToggleFor(second)
	if !p.Visible() || p.Anchor() != fyne.CanvasObject(second) {
		t.Error("toggling for another widget should move the popup, not close it")
	}

	p.ToggleFor(second)
	if p.Visible() {
		t.Error("toggling the same widget again should close the popup")
	}
}

func TestPopupHideKeyClosesIt(t *testing.T) {
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })

	entry := widget.NewEntry()
	win := test.NewWindow(entry)
	win.Resize(fyne.NewSize(800, 600))
	t.Cleanup(win.Close)

	p := NewPopup(win, WithNumpad())
	p.ShowFor(entry)
	p.Keyboard().Press(ActionKey(ActionHide, "Hide", 1))

	if p.Visible() {
		t.Error("the hide key should close the popup")
	}
}

func TestPopupPositionStaysInsideTheCanvas(t *testing.T) {
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })

	entry := widget.NewEntry()
	win := test.NewWindow(container.NewVBox(widget.NewLabel("top"), entry))
	win.Resize(fyne.NewSize(400, 300))
	t.Cleanup(win.Close)

	p := NewPopup(win)
	size := p.sizeFor()
	pos := p.positionFor(entry, size)
	canvasSize := win.Canvas().Size()

	if pos.X < 0 || pos.Y < 0 {
		t.Errorf("position %v is off the canvas", pos)
	}
	if pos.X+size.Width > canvasSize.Width+1 {
		t.Errorf("popup right edge %v exceeds the canvas width %v", pos.X+size.Width, canvasSize.Width)
	}
	if pos.Y+size.Height > canvasSize.Height+1 {
		t.Errorf("popup bottom edge %v exceeds the canvas height %v", pos.Y+size.Height, canvasSize.Height)
	}
}

func TestPopupSizeFollowsTheStyle(t *testing.T) {
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })

	win := test.NewWindow(widget.NewLabel("content"))
	win.Resize(fyne.NewSize(1200, 800))
	t.Cleanup(win.Close)

	text := NewPopup(win).sizeFor()
	numpad := NewPopup(win, WithNumpad()).sizeFor()
	full := NewPopup(win, WithFullKeyboard()).sizeFor()

	if numpad.Width >= text.Width {
		t.Errorf("the keypad (%v) should be narrower than the text block (%v)", numpad.Width, text.Width)
	}
	if full.Width <= text.Width {
		t.Errorf("the full keyboard (%v) should be wider than the text block (%v)", full.Width, text.Width)
	}

	fixed := fyne.NewSize(321, 123)
	p := NewPopup(win, WithPopupSize(fixed))
	if got := p.sizeFor(); got != fixed {
		t.Errorf("sizeFor() = %v, want the fixed %v", got, fixed)
	}
}

func TestPopupAutoShow(t *testing.T) {
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })

	entry := widget.NewEntry()
	win := test.NewWindow(entry)
	win.Resize(fyne.NewSize(800, 600))
	t.Cleanup(win.Close)

	p := NewPopup(win, WithAutoShow(true))
	// See the decorator test: the test driver runs fyne.Do inline, so the
	// poller is stopped and followFocus called directly.
	p.StopAutoShow()

	win.Canvas().Focus(entry)
	p.followFocus()
	if !p.Visible() {
		t.Fatal("focusing an entry should open the popup")
	}

	// While the popup is up the canvas reports the overlay focus, so polling
	// must not fight the popup that is already open.
	p.followFocus()
	if !p.Visible() {
		t.Error("the popup closed itself while open")
	}
}

func TestPopupAutoShowPollerStartsAndStops(t *testing.T) {
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })

	win := test.NewWindow(widget.NewEntry())
	t.Cleanup(win.Close)

	p := NewPopup(win, WithAutoShow(true), WithAutoShowInterval(time.Hour))
	if p.stopPoll == nil {
		t.Fatal("WithAutoShow should have started the poller")
	}
	p.startAutoShow()
	p.StopAutoShow()
	if p.stopPoll != nil {
		t.Error("StopAutoShow should have cleared the poller")
	}
	p.StopAutoShow()
}

func TestPopupWithoutWindowIsInert(t *testing.T) {
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })

	p := NewPopup(nil)
	p.ShowFor(widget.NewEntry())
	p.ShowAt(fyne.NewPos(0, 0))
	p.Hide()

	if p.Visible() {
		t.Error("a popup without a window cannot be visible")
	}
}
