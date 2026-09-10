package keyboard

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Popup shows the keyboard as a floating overlay next to a widget, the way a
// tooltip is anchored: the window layout does not change at all.
//
// Two consequences of Fyne's overlay handling are worth knowing. While the
// popup is up it is the only layer that receives taps, so a tap anywhere
// outside it dismisses the popup instead of reaching the widget below; the
// next tap then works normally. And because the canvas reports the overlay's
// focus, the popup pins the widget it was opened for as the keyboard target,
// rather than following the focus while it is open.
type Popup struct {
	win      fyne.Window
	keyboard *Keyboard
	popup    *widget.PopUp

	size     fyne.Size
	gap      float32
	pollRate time.Duration

	anchor   fyne.CanvasObject
	stopPoll chan struct{}
}

// NewPopup creates a floating keyboard for the given window. It is hidden
// until [Popup.ShowFor] or [Popup.ShowAt] is called.
func NewPopup(win fyne.Window, opts ...Option) *Popup {
	cfg := newConfig().apply(opts)

	var canvas fyne.Canvas
	if win != nil {
		canvas = win.Canvas()
	}

	p := &Popup{
		win:      win,
		keyboard: NewKeyboardWithStyle(canvas, cfg.layout, cfg.style),
		size:     cfg.size,
		gap:      cfg.gap,
		pollRate: cfg.pollRate,
	}
	p.keyboard.OnHide = p.Hide

	if canvas != nil {
		p.popup = widget.NewPopUp(p.keyboard, canvas)
		p.popup.Hide()
	}
	if cfg.autoShow {
		p.startAutoShow()
	}
	return p
}

// Keyboard returns the embedded keyboard, for layout switching and callbacks.
func (p *Popup) Keyboard() *Keyboard { return p.keyboard }

// Visible reports whether the popup is currently shown.
func (p *Popup) Visible() bool { return p.popup != nil && p.popup.Visible() }

// Anchor returns the widget the popup was last opened for.
func (p *Popup) Anchor() fyne.CanvasObject { return p.anchor }

// ShowFor opens the keyboard next to obj: below it when there is room, above
// it otherwise. When obj can take the focus it also becomes the keyboard
// target, so an application can open the keyboard for an entry that the user
// has not tapped yet.
func (p *Popup) ShowFor(obj fyne.CanvasObject) {
	if p.popup == nil || obj == nil {
		return
	}
	p.anchor = obj

	if focusable, ok := obj.(fyne.Focusable); ok {
		p.keyboard.SetTarget(focusable)
		if canvas := p.canvas(); canvas != nil && canvas.Focused() != focusable {
			canvas.Focus(focusable)
		}
	} else {
		p.keyboard.captureFocus()
	}

	size := p.sizeFor()
	p.showAt(p.positionFor(obj, size), size)
}

// ShowAt opens the keyboard at an absolute canvas position, keeping whatever
// target the keyboard already has.
func (p *Popup) ShowAt(pos fyne.Position) {
	if p.popup == nil {
		return
	}
	p.keyboard.captureFocus()
	p.showAt(pos, p.sizeFor())
}

func (p *Popup) showAt(pos fyne.Position, size fyne.Size) {
	p.popup.Resize(size)
	p.popup.ShowAtPosition(pos)
}

// Hide closes the popup.
func (p *Popup) Hide() {
	if p.popup != nil {
		p.popup.Hide()
	}
}

// ToggleFor opens the popup for obj, or closes it when it is already open for
// that same widget.
func (p *Popup) ToggleFor(obj fyne.CanvasObject) {
	if p.Visible() && p.anchor == obj {
		p.Hide()
		return
	}
	p.ShowFor(obj)
}

// SetLayout switches the keyboard layout at run time.
func (p *Popup) SetLayout(l Layout) error { return p.keyboard.SetLayout(l) }

// SetStyle switches between the character block, the keypad and both.
func (p *Popup) SetStyle(style Style) { p.keyboard.SetStyle(style) }

// SetSize fixes the popup size. A zero size restores automatic sizing.
func (p *Popup) SetSize(size fyne.Size) { p.size = size }

func (p *Popup) canvas() fyne.Canvas {
	if p.win == nil {
		return nil
	}
	return p.win.Canvas()
}

// sizeFor returns the configured size, or one derived from the layout's
// proportions and capped to the canvas.
func (p *Popup) sizeFor() fyne.Size {
	if p.size.Width > 0 && p.size.Height > 0 {
		return p.size
	}

	l := p.keyboard.CurrentLayout()
	units := maxRowUnits(l)
	rows := float32(len(l.Rows))
	if units <= 0 || rows == 0 {
		return p.keyboard.MinSize()
	}

	capSize := fyne.Max(theme.TextSize()*3, 36)
	size := fyne.NewSize(units*capSize, rows*capSize)

	// Fitting the canvas wins over the keyboard's own minimum size: the caps
	// scale down with the row layout, an oversized popup would not fit at all.
	if canvas := p.canvas(); canvas != nil {
		available := canvas.Size()
		maxWidth := available.Width - 2*theme.Padding()
		maxHeight := available.Height * 0.75
		if scale := fyne.Min(maxWidth/size.Width, maxHeight/size.Height); scale < 1 && scale > 0 {
			size = fyne.NewSize(size.Width*scale, size.Height*scale)
		}
	}
	return size
}

// positionFor anchors the popup below obj, flipping above it when the lower
// edge of the canvas is in the way, and keeps it inside the canvas.
func (p *Popup) positionFor(obj fyne.CanvasObject, size fyne.Size) fyne.Position {
	driver := fyne.CurrentApp().Driver()
	origin := driver.AbsolutePositionForObject(obj)

	pos := fyne.NewPos(origin.X, origin.Y+obj.Size().Height+p.gap)

	canvas := p.canvas()
	if canvas == nil {
		return pos
	}
	available := canvas.Size()
	pad := theme.Padding()

	if pos.Y+size.Height > available.Height-pad {
		if above := origin.Y - p.gap - size.Height; above >= pad {
			pos.Y = above
		} else {
			pos.Y = fyne.Max(pad, available.Height-pad-size.Height)
		}
	}
	if pos.X+size.Width > available.Width-pad {
		pos.X = available.Width - pad - size.Width
	}
	pos.X = fyne.Max(pad, pos.X)
	return pos
}

// startAutoShow opens the popup whenever a text input takes the focus.
func (p *Popup) startAutoShow() {
	if p.win == nil || p.stopPoll != nil {
		return
	}
	stop := make(chan struct{})
	p.stopPoll = stop

	go func() {
		ticker := time.NewTicker(p.pollRate)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				fyne.Do(p.followFocus)
			}
		}
	}()
}

// StopAutoShow ends the focus polling started by [WithAutoShow].
func (p *Popup) StopAutoShow() {
	if p.stopPoll == nil {
		return
	}
	close(p.stopPoll)
	p.stopPoll = nil
}

func (p *Popup) followFocus() {
	canvas := p.canvas()
	if canvas == nil || p.Visible() {
		return
	}
	focused, ok := canvas.Focused().(mobile.Keyboardable)
	if !ok {
		return
	}
	if obj, ok := focused.(fyne.CanvasObject); ok {
		p.ShowFor(obj)
	}
}

// maxRowUnits returns the widest row of a layout in relative units.
func maxRowUnits(l Layout) float32 {
	var widest float32
	for _, row := range l.Rows {
		var sum float32
		for _, key := range row {
			sum += key.Units()
		}
		if sum > widest {
			widest = sum
		}
	}
	return widest
}
