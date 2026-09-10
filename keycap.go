package keyboard

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// keyCap is a single tappable cap. It deliberately does not implement
// fyne.Focusable: the canvas focus has to stay on the widget being typed into.
// Focus is captured on press (before the driver unfocuses on a tap of a non
// focusable object) via desktop.Mouseable and mobile.Touchable.
type keyCap struct {
	widget.BaseWidget

	kb  *Keyboard
	key Key

	active  bool
	hovered bool
	pressed bool
}

var (
	_ fyne.Widget            = (*keyCap)(nil)
	_ fyne.Tappable          = (*keyCap)(nil)
	_ desktop.Mouseable      = (*keyCap)(nil)
	_ desktop.Hoverable      = (*keyCap)(nil)
	_ desktop.Cursorable     = (*keyCap)(nil)
	_ mobile.Touchable       = (*keyCap)(nil)
	_ fyne.SecondaryTappable = (*keyCap)(nil)
)

func newKeyCap(kb *Keyboard, key Key) *keyCap {
	c := &keyCap{kb: kb, key: key}
	c.ExtendBaseWidget(c)
	return c
}

// setKey swaps the key this cap represents, for example when a modifier
// changes the active level or the layout is switched.
func (c *keyCap) setKey(key Key, active bool) {
	if c.key == key && c.active == active {
		return
	}
	c.key = key
	c.active = active
	c.Refresh()
}

func (c *keyCap) Tapped(*fyne.PointEvent) {
	if !c.key.Tappable() {
		return
	}
	c.kb.Press(c.key)
}

// TappedSecondary is ignored, but implementing it keeps a right click from
// falling through to whatever is behind the keyboard.
func (c *keyCap) TappedSecondary(*fyne.PointEvent) {}

func (c *keyCap) MouseDown(*desktop.MouseEvent) {
	c.kb.captureFocus()
	c.setPressed(true)
}

func (c *keyCap) MouseUp(*desktop.MouseEvent) { c.setPressed(false) }

func (c *keyCap) MouseIn(*desktop.MouseEvent) {
	if !c.key.Tappable() {
		return
	}
	c.hovered = true
	c.Refresh()
}

func (c *keyCap) MouseMoved(*desktop.MouseEvent) {}

func (c *keyCap) MouseOut() {
	if !c.hovered && !c.pressed {
		return
	}
	c.hovered = false
	c.pressed = false
	c.Refresh()
}

func (c *keyCap) TouchDown(*mobile.TouchEvent) {
	c.kb.captureFocus()
	c.setPressed(true)
}

func (c *keyCap) TouchUp(*mobile.TouchEvent)     { c.setPressed(false) }
func (c *keyCap) TouchCancel(*mobile.TouchEvent) { c.setPressed(false) }

func (c *keyCap) Cursor() desktop.Cursor { return desktop.DefaultCursor }

func (c *keyCap) setPressed(pressed bool) {
	if !c.key.Tappable() || c.pressed == pressed {
		return
	}
	c.pressed = pressed
	c.Refresh()
}

func (c *keyCap) CreateRenderer() fyne.WidgetRenderer {
	c.ExtendBaseWidget(c)
	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = theme.Size(theme.SizeNameSelectionRadius)
	label := canvas.NewText(c.key.Text(), theme.Color(theme.ColorNameForeground))
	label.Alignment = fyne.TextAlignCenter
	label.TextSize = theme.TextSize()
	r := &keyCapRenderer{cap: c, bg: bg, label: label}
	r.applyState()
	return r
}

type keyCapRenderer struct {
	cap   *keyCap
	bg    *canvas.Rectangle
	label *canvas.Text
}

var _ fyne.WidgetRenderer = (*keyCapRenderer)(nil)

func (r *keyCapRenderer) Destroy() {}

func (r *keyCapRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.bg.Move(fyne.NewPos(0, 0))
	textHeight := r.label.MinSize().Height
	r.label.Resize(fyne.NewSize(size.Width, textHeight))
	r.label.Move(fyne.NewPos(0, (size.Height-textHeight)/2))
}

func (r *keyCapRenderer) MinSize() fyne.Size {
	pad := theme.InnerPadding()
	min := r.label.MinSize()
	// Caps stay finger friendly even when their label is a single narrow rune.
	side := theme.TextSize() + pad
	return fyne.NewSize(fyne.Max(min.Width+pad, side), fyne.Max(min.Height+pad, side))
}

func (r *keyCapRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.label}
}

func (r *keyCapRenderer) Refresh() {
	r.applyState()
	r.Layout(r.cap.Size())
	canvas.Refresh(r.cap)
}

func (r *keyCapRenderer) applyState() {
	c := r.cap
	r.label.Text = c.key.Text()
	r.label.TextSize = theme.TextSize()
	r.bg.CornerRadius = theme.Size(theme.SizeNameSelectionRadius)

	if !c.key.Tappable() {
		r.bg.FillColor = color.Transparent
		r.label.Color = color.Transparent
		r.bg.Refresh()
		r.label.Refresh()
		return
	}

	switch {
	case c.pressed:
		r.bg.FillColor = theme.Color(theme.ColorNamePressed)
	case c.active:
		r.bg.FillColor = theme.Color(theme.ColorNamePrimary)
	case c.hovered:
		r.bg.FillColor = theme.Color(theme.ColorNameHover)
	default:
		r.bg.FillColor = theme.Color(theme.ColorNameButton)
	}

	if c.active {
		r.label.Color = theme.Color(theme.ColorNameForegroundOnPrimary)
	} else {
		r.label.Color = theme.Color(theme.ColorNameForeground)
	}
	r.bg.Refresh()
	r.label.Refresh()
}
