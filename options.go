package keyboard

import (
	"time"

	"fyne.io/fyne/v2"
)

// Style selects which blocks a keyboard shows.
type Style uint8

const (
	// StyleText is the character block alone.
	StyleText Style = iota
	// StyleNumpad is the numeric keypad alone, including the double zero key.
	StyleNumpad
	// StyleFull is the character block with the numeric keypad attached.
	StyleFull
)

// apply turns the requested layout into the one the style actually shows.
func (s Style) apply(l Layout) Layout {
	switch s {
	case StyleNumpad:
		return Numpad()
	case StyleFull:
		return Full(l)
	default:
		return l
	}
}

// String implements fmt.Stringer.
func (s Style) String() string {
	switch s {
	case StyleNumpad:
		return "numpad"
	case StyleFull:
		return "full"
	default:
		return "text"
	}
}

// config holds everything the constructors of [Decorator] and [Popup] can be
// told. Options that do not apply to a given host are ignored by it.
type config struct {
	layout   Layout
	style    Style
	mode     DockMode
	offset   float64
	autoShow bool
	pollRate time.Duration
	size     fyne.Size
	gap      float32
}

func newConfig() *config {
	return &config{
		layout:   defaultLayout(),
		offset:   DefaultOffset,
		pollRate: 200 * time.Millisecond,
		gap:      8,
	}
}

func (c *config) apply(opts []Option) *config {
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Option configures a [Decorator] or a [Popup].
type Option func(*config)

// WithLayout sets the layout the keyboard starts with.
func WithLayout(l Layout) Option {
	return func(c *config) {
		if err := l.Validate(); err != nil {
			fyne.LogError("keyboard: ignoring invalid layout", err)
			return
		}
		c.layout = l
	}
}

// WithLayoutNamed sets the start layout by registry name, for example "de-DE".
func WithLayoutNamed(name string) Option {
	return func(c *config) { c.layout = LayoutFor(name) }
}

// WithStyle selects the character block, the numeric keypad, or both.
func WithStyle(style Style) Option {
	return func(c *config) { c.style = style }
}

// WithNumpad is shorthand for WithStyle(StyleNumpad): a numeric keypad only.
func WithNumpad() Option { return WithStyle(StyleNumpad) }

// WithFullKeyboard is shorthand for WithStyle(StyleFull): the character block
// with the numeric keypad attached.
func WithFullKeyboard() Option { return WithStyle(StyleFull) }

// WithDockMode selects split or border docking. [Decorator] only.
func WithDockMode(mode DockMode) Option {
	return func(c *config) { c.mode = mode }
}

// WithOffset sets the split offset used while the keyboard is shown. Values
// are clamped to 0.1 - 0.95. [Decorator] with [DockSplit] only.
func WithOffset(offset float64) Option {
	return func(c *config) { c.offset = clamp(offset, 0.1, 0.95) }
}

// WithAutoShow makes the keyboard appear whenever a text input takes the focus.
func WithAutoShow(auto bool) Option {
	return func(c *config) { c.autoShow = auto }
}

// WithAutoShowInterval sets how often the focus is sampled for auto show.
// Fyne has no focus change callback, so auto show polls the canvas.
func WithAutoShowInterval(every time.Duration) Option {
	return func(c *config) {
		if every > 0 {
			c.pollRate = every
		}
	}
}

// WithPopupSize fixes the size of a [Popup]. Without it the popup sizes itself
// from the canvas and the layout's proportions.
func WithPopupSize(size fyne.Size) Option {
	return func(c *config) { c.size = size }
}

// WithPopupGap sets the distance in pixels between a [Popup] and the widget it
// is anchored to.
func WithPopupGap(gap float32) Option {
	return func(c *config) { c.gap = gap }
}

func clamp(v, lo, hi float64) float64 {
	switch {
	case v < lo:
		return lo
	case v > hi:
		return hi
	default:
		return v
	}
}
