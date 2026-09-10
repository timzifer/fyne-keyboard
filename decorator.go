package keyboard

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/mobile"
)

// DockMode selects how the keyboard is attached to the content.
type DockMode uint8

const (
	// DockSplit puts the keyboard below the content in a container.Split, so
	// the user can drag the divider to resize it.
	DockSplit DockMode = iota
	// DockBorder docks the keyboard to the bottom edge at its minimum size.
	// There is no divider, which is usually the calmer choice for kiosks.
	DockBorder
)

// DefaultOffset is the split offset used when the keyboard is shown: the
// content keeps the upper 60% of the window.
const DefaultOffset = 0.6

// Decorator wraps an application's main widget and adds an on screen keyboard
// below it. Use [Decorator.Content] as the window content.
//
// For a keyboard that floats next to a single widget instead of taking a share
// of the window, see [Popup].
type Decorator struct {
	win      fyne.Window
	keyboard *Keyboard
	holder   *fyne.Container
	root     fyne.CanvasObject
	split    *container.Split

	offset   float64
	pollRate time.Duration

	visible  bool
	stopPoll chan struct{}
}

// NewDecorator wraps content with an on screen keyboard. The keyboard starts
// hidden; call [Decorator.Show] or bind [Decorator.Toggle] to a shortcut.
func NewDecorator(win fyne.Window, content fyne.CanvasObject, opts ...Option) *Decorator {
	cfg := newConfig().apply(opts)

	var canvas fyne.Canvas
	if win != nil {
		canvas = win.Canvas()
	}

	d := &Decorator{
		win:      win,
		holder:   container.NewStack(),
		keyboard: NewKeyboardWithStyle(canvas, cfg.layout, cfg.style),
		offset:   cfg.offset,
		pollRate: cfg.pollRate,
	}
	d.keyboard.OnHide = d.Hide
	d.keyboard.Hide()
	d.SetContent(content)

	switch cfg.mode {
	case DockBorder:
		d.root = container.NewBorder(nil, d.keyboard, nil, nil, d.holder)
	default:
		d.split = container.NewVSplit(d.holder, d.keyboard)
		d.split.SetOffset(1)
		d.root = d.split
	}

	if cfg.autoShow {
		d.startAutoShow()
	}
	return d
}

// Content returns the object to hand to fyne.Window.SetContent.
func (d *Decorator) Content() fyne.CanvasObject { return d.root }

// SetContent replaces the decorated application widget. It exists so that the
// content can refer back to the decorator, for example with a button that
// calls [Decorator.Toggle]: build the decorator with a nil content first, then
// hand it the finished widget tree.
func (d *Decorator) SetContent(content fyne.CanvasObject) {
	if content == nil {
		d.holder.Objects = nil
	} else {
		d.holder.Objects = []fyne.CanvasObject{content}
	}
	d.holder.Refresh()
}

// DecoratedContent returns the application widget being decorated.
func (d *Decorator) DecoratedContent() fyne.CanvasObject {
	if len(d.holder.Objects) == 0 {
		return nil
	}
	return d.holder.Objects[0]
}

// Keyboard returns the embedded keyboard, for layout switching and callbacks.
func (d *Decorator) Keyboard() *Keyboard { return d.keyboard }

// Visible reports whether the keyboard is currently shown.
func (d *Decorator) Visible() bool { return d.visible }

// Show raises the keyboard, remembering the widget that currently has focus so
// the first keystroke lands in the right place.
func (d *Decorator) Show() {
	d.keyboard.captureFocus()
	if d.visible {
		return
	}
	d.visible = true
	d.keyboard.Show()
	if d.split != nil {
		d.split.SetOffset(d.offset)
	}
	d.root.Refresh()
}

// Hide drops the keyboard and gives the whole area back to the content.
func (d *Decorator) Hide() {
	if !d.visible {
		return
	}
	d.visible = false
	d.keyboard.Hide()
	if d.split != nil {
		d.split.SetOffset(1)
	}
	d.root.Refresh()
}

// Toggle shows the keyboard when hidden and hides it when shown.
func (d *Decorator) Toggle() {
	if d.visible {
		d.Hide()
		return
	}
	d.Show()
}

// SetOffset changes the split offset used while the keyboard is shown.
func (d *Decorator) SetOffset(offset float64) {
	d.offset = clamp(offset, 0.1, 0.95)
	if d.visible && d.split != nil {
		d.split.SetOffset(d.offset)
	}
}

// SetLayout switches the keyboard layout at run time.
func (d *Decorator) SetLayout(l Layout) error { return d.keyboard.SetLayout(l) }

// SetStyle switches between the character block, the keypad and both.
func (d *Decorator) SetStyle(style Style) { d.keyboard.SetStyle(style) }

// startAutoShow samples the canvas focus and follows text inputs. Fyne offers
// no focus change notification, so this polls; the sampling itself runs on the
// UI thread via fyne.Do.
func (d *Decorator) startAutoShow() {
	if d.win == nil || d.stopPoll != nil {
		return
	}
	stop := make(chan struct{})
	d.stopPoll = stop

	go func() {
		ticker := time.NewTicker(d.pollRate)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				fyne.Do(d.followFocus)
			}
		}
	}()
}

// StopAutoShow ends the focus polling started by [WithAutoShow].
func (d *Decorator) StopAutoShow() {
	if d.stopPoll == nil {
		return
	}
	close(d.stopPoll)
	d.stopPoll = nil
}

func (d *Decorator) followFocus() {
	if d.win == nil {
		return
	}
	canvas := d.win.Canvas()
	if canvas == nil {
		return
	}
	focused := canvas.Focused()
	if _, ok := focused.(mobile.Keyboardable); ok {
		d.Show()
		return
	}
	// A nil focus is not a reason to hide: tapping a cap unfocuses the entry
	// for the moment between press and release, and the keyboard restores the
	// focus itself on the next keystroke. Only a different widget taking the
	// focus means the user has moved on.
	if focused != nil {
		d.Hide()
	}
}
