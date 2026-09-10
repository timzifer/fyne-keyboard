package keyboard

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Keyboard is the on screen keyboard widget. It renders a [Layout] as rows of
// caps and forwards presses to whatever currently holds the canvas focus.
//
// The widget itself is not focusable and neither are its caps, so typing on it
// never steals focus from the widget being typed into.
type Keyboard struct {
	widget.BaseWidget

	// OnKey is called after every press, with the resolved key. Useful for
	// click feedback or logging.
	OnKey func(Key)
	// OnLayoutChange is called after the layout was switched.
	OnLayoutChange func(Layout)
	// OnHide is called when a key with ActionHide is pressed.
	OnHide func()

	canvas  fyne.Canvas
	base    Layout
	current Layout
	style   Style

	shift bool
	caps  bool
	altGr bool
	dead  rune

	target fyne.Focusable

	grid   *fyne.Container
	caps2d [][]*keyCap
}

var _ fyne.Widget = (*Keyboard)(nil)

// NewKeyboard creates a keyboard for the given canvas and layout. The canvas
// is used to restore focus before every keystroke and may be nil in tests, in
// which case the target must be set with [Keyboard.SetTarget].
func NewKeyboard(c fyne.Canvas, l Layout) *Keyboard {
	return newKeyboard(c, l, StyleText)
}

// NewKeyboardWithStyle creates a keyboard in the given style: the plain
// character block, a numeric keypad, or both side by side.
func NewKeyboardWithStyle(c fyne.Canvas, l Layout, style Style) *Keyboard {
	return newKeyboard(c, l, style)
}

func newKeyboard(c fyne.Canvas, l Layout, style Style) *Keyboard {
	if len(l.Rows) == 0 {
		l = defaultLayout()
	}
	k := &Keyboard{canvas: c, base: l, style: style}
	k.current = style.apply(l)
	k.ExtendBaseWidget(k)
	k.build()
	return k
}

// CurrentLayout returns the layout in use, with the style applied.
func (k *Keyboard) CurrentLayout() Layout { return k.current }

// BaseLayout returns the layout as it was set, before the style was applied.
func (k *Keyboard) BaseLayout() Layout { return k.base }

// Style returns the keyboard style.
func (k *Keyboard) Style() Style { return k.style }

// SetStyle switches between the character block, the numeric keypad and the
// full size keyboard, keeping the current layout.
func (k *Keyboard) SetStyle(style Style) {
	if style == k.style {
		return
	}
	k.style = style
	k.reload(k.base)
}

// SetLayout swaps the layout and rebuilds the caps. It reports an error and
// keeps the previous layout if the new one is structurally invalid.
func (k *Keyboard) SetLayout(l Layout) error {
	if err := l.Validate(); err != nil {
		return err
	}
	k.reload(l)
	return nil
}

func (k *Keyboard) reload(l Layout) {
	k.base = l
	k.current = k.style.apply(l)
	k.shift, k.altGr, k.dead = false, false, 0
	k.build()
	k.Refresh()
	if k.OnLayoutChange != nil {
		k.OnLayoutChange(k.current)
	}
}

// SetCanvas points the keyboard at the canvas whose focus it should follow.
func (k *Keyboard) SetCanvas(c fyne.Canvas) { k.canvas = c }

// Target returns the widget presses are sent to: the explicitly set target, or
// else whatever the canvas reports as focused.
func (k *Keyboard) Target() fyne.Focusable {
	if k.target != nil {
		return k.target
	}
	if k.canvas != nil {
		return k.canvas.Focused()
	}
	return nil
}

// SetTarget pins the widget that presses are sent to. Passing nil returns to
// following the canvas focus.
func (k *Keyboard) SetTarget(f fyne.Focusable) { k.target = f }

// Shift reports whether the one shot shift level is armed.
func (k *Keyboard) Shift() bool { return k.shift }

// CapsLock reports whether caps lock is latched.
func (k *Keyboard) CapsLock() bool { return k.caps }

// AltGr reports whether the AltGr level is active.
func (k *Keyboard) AltGr() bool { return k.altGr }

// Dead returns the armed dead key, or 0 when none is pending.
func (k *Keyboard) Dead() rune { return k.dead }

// Level returns the character level the caps currently show.
func (k *Keyboard) Level() Level {
	switch {
	case k.altGr:
		return LevelAltGr
	case k.shift != k.caps:
		return LevelShift
	default:
		return LevelBase
	}
}

// captureFocus remembers the focused widget. It is called on press, before the
// driver unfocuses because a non focusable object was tapped.
func (k *Keyboard) captureFocus() {
	if k.canvas == nil {
		return
	}
	if f := k.canvas.Focused(); f != nil {
		k.target = f
	}
}

// Press dispatches a key as if the user had tapped that cap.
func (k *Keyboard) Press(key Key) {
	deadBefore := k.dead
	switch key.Type {
	case KeySpacer:
		return
	case KeyModifier:
		k.toggleModifier(key.Mod)
	case KeyAction:
		k.runAction(key.Action)
	case KeyDead:
		k.flushDead()
		k.dead = key.Rune
		k.consumeShift()
	case KeyControl:
		k.flushDead()
		k.typeKey(key.Key)
		k.consumeShift()
	case KeyChar:
		if key.Runes != "" {
			for _, r := range key.Runes {
				k.typeChar(r)
			}
		} else {
			k.typeChar(key.Rune)
		}
		k.consumeShift()
	}
	if k.dead != deadBefore {
		k.refreshCaps()
	}
	if k.OnKey != nil {
		k.OnKey(key)
	}
}

func (k *Keyboard) toggleModifier(mod Modifier) {
	switch mod {
	case ModShift:
		k.shift = !k.shift
	case ModCapsLock:
		k.caps = !k.caps
	case ModAltGr:
		k.altGr = !k.altGr
	default:
		return
	}
	k.refreshCaps()
}

func (k *Keyboard) runAction(action Action) {
	switch action {
	case ActionHide:
		if k.OnHide != nil {
			k.OnHide()
		}
	case ActionNextLayout:
		if next, ok := nextLayout(k.base.Name); ok {
			_ = k.SetLayout(next)
		}
	}
}

func (k *Keyboard) typeChar(r rune) {
	if r == 0 {
		return
	}
	if k.dead != 0 {
		dead := k.dead
		k.dead = 0
		if combined, ok := k.current.Combine(dead, r); ok {
			k.typeRune(combined)
			return
		}
		k.typeRune(dead)
	}
	k.typeRune(r)
}

// flushDead emits a pending diacritic on its own, which is what a real layout
// does when the dead key is followed by something it cannot combine with.
func (k *Keyboard) flushDead() {
	if k.dead == 0 {
		return
	}
	dead := k.dead
	k.dead = 0
	k.typeRune(dead)
}

func (k *Keyboard) typeRune(r rune) {
	if target := k.focus(); target != nil {
		target.TypedRune(r)
	}
}

func (k *Keyboard) typeKey(name fyne.KeyName) {
	if name == "" {
		return
	}
	if target := k.focus(); target != nil {
		target.TypedKey(&fyne.KeyEvent{Name: name})
	}
}

// focus restores the canvas focus to the target and returns it. Focus is only
// ever changed from a tap handler, never from FocusGained or FocusLost, which
// the Fyne docs warn can deadlock.
func (k *Keyboard) focus() fyne.Focusable {
	target := k.Target()
	if target == nil {
		return nil
	}
	if k.canvas != nil && k.canvas.Focused() != target {
		k.canvas.Focus(target)
	}
	return target
}

// consumeShift clears the one shot shift level after a character was typed.
func (k *Keyboard) consumeShift() {
	if !k.shift && !k.altGr {
		return
	}
	k.shift = false
	k.altGr = false
	k.refreshCaps()
}

// build creates the cap widgets for the current layout. It reuses the grid
// container: the renderer holds a reference to it, so replacing the container
// would leave the previous keyboard on screen after a layout or style change.
func (k *Keyboard) build() {
	rows := make([]fyne.CanvasObject, 0, len(k.current.Rows))
	k.caps2d = make([][]*keyCap, len(k.current.Rows))

	for r, row := range k.current.Rows {
		caps := make([]*keyCap, len(row))
		objects := make([]fyne.CanvasObject, len(row))
		units := make([]float32, len(row))
		for c, key := range row {
			kc := newKeyCap(k, key)
			caps[c] = kc
			objects[c] = kc
			units[c] = key.Units()
		}
		k.caps2d[r] = caps
		rows = append(rows, container.New(&keyRowLayout{units: units}, objects...))
	}

	if k.grid == nil {
		k.grid = container.New(layout.NewGridLayoutWithRows(len(rows)), rows...)
	} else {
		k.grid.Layout = layout.NewGridLayoutWithRows(len(rows))
		k.grid.Objects = rows
		k.grid.Refresh()
	}
	k.refreshCaps()
}

// refreshCaps re-resolves every cap for the active level and modifier state.
func (k *Keyboard) refreshCaps() {
	level := k.Level()
	for r, caps := range k.caps2d {
		for c, kc := range caps {
			kc.setKey(k.current.KeyAt(r, c, level), k.modifierActive(k.current.Rows[r][c]))
		}
	}
}

func (k *Keyboard) modifierActive(key Key) bool {
	switch {
	case key.Type == KeyModifier && key.Mod == ModShift:
		return k.shift
	case key.Type == KeyModifier && key.Mod == ModCapsLock:
		return k.caps
	case key.Type == KeyModifier && key.Mod == ModAltGr:
		return k.altGr
	case key.Type == KeyDead:
		return k.dead == key.Rune
	default:
		return false
	}
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer.
func (k *Keyboard) CreateRenderer() fyne.WidgetRenderer {
	k.ExtendBaseWidget(k)
	if k.grid == nil {
		k.build()
	}
	return widget.NewSimpleRenderer(k.grid)
}
