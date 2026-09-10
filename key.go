package keyboard

import (
	"fmt"
	"strings"
	"unicode"

	"fyne.io/fyne/v2"
)

// KeyType classifies what pressing a [Key] does.
type KeyType uint8

const (
	// KeyChar types a rune into the focused widget via fyne.Focusable.TypedRune.
	KeyChar KeyType = iota
	// KeyControl sends a fyne.KeyEvent via fyne.Focusable.TypedKey.
	KeyControl
	// KeyModifier toggles a keyboard level such as shift or AltGr.
	KeyModifier
	// KeySpacer renders empty space; it is not tappable.
	KeySpacer
	// KeyDead arms a diacritic that combines with the next character key.
	KeyDead
	// KeyAction triggers keyboard behaviour such as hiding or switching layout.
	KeyAction
)

// Modifier identifies the level toggle a [KeyModifier] key operates.
type Modifier uint8

const (
	// ModNone is the zero value, used by non modifier keys.
	ModNone Modifier = iota
	// ModShift is a one shot upper level toggle.
	ModShift
	// ModCapsLock is a latching upper level toggle that only affects letters.
	ModCapsLock
	// ModAltGr selects the third (AltGr) level.
	ModAltGr
)

// Action identifies the behaviour a [KeyAction] key triggers.
type Action uint8

const (
	// ActionNone is the zero value, used by non action keys.
	ActionNone Action = iota
	// ActionHide hides the keyboard.
	ActionHide
	// ActionNextLayout switches to the next registered layout.
	ActionNextLayout
)

// Level is one of the character levels a layout can provide.
type Level uint8

const (
	// LevelBase is the unmodified level.
	LevelBase Level = iota
	// LevelShift is the level reached while shift (or caps lock) is active.
	LevelShift
	// LevelAltGr is the level reached while AltGr is active.
	LevelAltGr
)

// Key describes a single cap on the keyboard.
//
// The zero value is a character key with no rune, which renders as an empty
// cap. Width is measured in relative units; 0 is treated as 1.
type Key struct {
	// Rune is typed into the focused widget for KeyChar keys and holds the
	// diacritic for KeyDead keys.
	Rune rune
	// Runes is typed instead of Rune when set, one rune after the other. It
	// exists for caps such as the numeric keypad's "00".
	Runes string
	// Key is the key event sent for KeyControl keys.
	Key fyne.KeyName
	// Label is the cap text. Empty means "derive from Rune or Key".
	Label string
	// Width is the relative cap width in units. 0 means 1.
	Width float32
	// Type selects how a press is dispatched.
	Type KeyType
	// Mod is the toggled level for KeyModifier keys.
	Mod Modifier
	// Action is the triggered behaviour for KeyAction keys.
	Action Action
}

// Row is one horizontal line of keys.
type Row []Key

// Units returns the effective relative width of the key.
func (k Key) Units() float32 {
	if k.Width <= 0 {
		return 1
	}
	return k.Width
}

// Text returns the label to draw on the cap.
func (k Key) Text() string {
	if k.Label != "" {
		return k.Label
	}
	if k.Runes != "" {
		return k.Runes
	}
	if k.Rune != 0 {
		return string(k.Rune)
	}
	if k.Key != "" {
		return string(k.Key)
	}
	return ""
}

// Tappable reports whether the key reacts to a press.
func (k Key) Tappable() bool {
	return k.Type != KeySpacer
}

// CharKey builds a character key of one unit width.
func CharKey(r rune) Key { return Key{Rune: r, Type: KeyChar} }

// SequenceKey builds a character key that types more than one rune, such as
// the numeric keypad's "00".
func SequenceKey(runes string, width float32) Key {
	return Key{Runes: runes, Type: KeyChar, Width: width}
}

// CharKeyWidth builds a character key with an explicit relative width.
func CharKeyWidth(r rune, width float32) Key {
	return Key{Rune: r, Type: KeyChar, Width: width}
}

// ControlKey builds a key that sends a fyne.KeyEvent.
func ControlKey(name fyne.KeyName, label string, width float32) Key {
	return Key{Key: name, Label: label, Type: KeyControl, Width: width}
}

// ModifierKey builds a level toggle key.
func ModifierKey(mod Modifier, label string, width float32) Key {
	return Key{Label: label, Type: KeyModifier, Mod: mod, Width: width}
}

// SpacerKey builds inert space, used to offset a row.
func SpacerKey(width float32) Key { return Key{Type: KeySpacer, Width: width} }

// DeadKey builds a diacritic key that combines with the next character.
func DeadKey(r rune, label string) Key {
	if label == "" {
		label = string(r)
	}
	return Key{Rune: r, Label: label, Type: KeyDead}
}

// ActionKey builds a key that drives the keyboard itself.
func ActionKey(action Action, label string, width float32) Key {
	return Key{Label: label, Type: KeyAction, Action: action, Width: width}
}

// CharRow builds a row of one unit character keys, one per rune of s.
func CharRow(s string) Row {
	row := make(Row, 0, len(s))
	for _, r := range s {
		row = append(row, CharKey(r))
	}
	return row
}

// Layout is a complete keyboard description. It is plain data, so layouts can
// be written in Go, loaded from JSON, or generated at run time.
type Layout struct {
	// Name is the language tag identifying the layout, for example "de-DE".
	Name string
	// Title is the human readable name shown on the layout switch key.
	Title string
	// Rows is the base level and defines the geometry of the keyboard.
	Rows []Row
	// Shift is the optional shift level. Missing keys fall back to the upper
	// case form of the base rune.
	Shift []Row
	// AltGr is the optional third level. Missing keys fall back to the base.
	AltGr []Row
	// DeadKeys maps a diacritic to the characters it can combine with.
	DeadKeys map[rune]map[rune]rune
}

// Combine returns the rune produced by combining dead key d with r, and
// whether such a combination exists in this layout.
func (l Layout) Combine(d, r rune) (rune, bool) {
	table, ok := l.DeadKeys[d]
	if !ok {
		return 0, false
	}
	combined, ok := table[r]
	return combined, ok
}

// KeyAt resolves the key at the given row and column for a level, falling back
// to the base level when the requested level does not define the key.
func (l Layout) KeyAt(row, col int, level Level) Key {
	base := l.baseKey(row, col)
	if level == LevelBase || (base.Type != KeyChar && base.Type != KeyDead) {
		return base
	}

	rows := l.Shift
	if level == LevelAltGr {
		rows = l.AltGr
	}
	if over, ok := lookup(rows, row, col); ok && !isEmptyOverlay(over) {
		over.Width = base.Units()
		return over
	}
	if level == LevelAltGr || base.Type != KeyChar || base.Runes != "" {
		return base
	}

	upper := unicode.ToUpper(base.Rune)
	if upper == base.Rune {
		return base
	}
	base.Rune = upper
	if base.Label != "" {
		base.Label = strings.ToUpper(base.Label)
	}
	return base
}

// isEmptyOverlay reports whether an overlay slot says nothing, in which case
// the resolver falls back to the base key.
func isEmptyOverlay(k Key) bool {
	return k.Type == KeySpacer || (k.Rune == 0 && k.Runes == "" && k.Label == "" && k.Key == "")
}

func (l Layout) baseKey(row, col int) Key {
	k, _ := lookup(l.Rows, row, col)
	return k
}

func lookup(rows []Row, row, col int) (Key, bool) {
	if row < 0 || row >= len(rows) {
		return Key{}, false
	}
	r := rows[row]
	if col < 0 || col >= len(r) {
		return Key{}, false
	}
	return r[col], true
}

// Validate reports structural problems: a missing name, an empty base level,
// or an overlay level whose rows do not line up with the base rows.
func (l Layout) Validate() error {
	if l.Name == "" {
		return fmt.Errorf("keyboard: layout has no name")
	}
	if len(l.Rows) == 0 {
		return fmt.Errorf("keyboard: layout %q has no rows", l.Name)
	}
	for i, row := range l.Rows {
		if len(row) == 0 {
			return fmt.Errorf("keyboard: layout %q row %d is empty", l.Name, i)
		}
	}
	if err := l.validateOverlay("shift", l.Shift); err != nil {
		return err
	}
	return l.validateOverlay("altgr", l.AltGr)
}

func (l Layout) validateOverlay(name string, rows []Row) error {
	if len(rows) == 0 {
		return nil
	}
	if len(rows) != len(l.Rows) {
		return fmt.Errorf("keyboard: layout %q %s level has %d rows, base has %d",
			l.Name, name, len(rows), len(l.Rows))
	}
	for i, row := range rows {
		if len(row) != len(l.Rows[i]) {
			return fmt.Errorf("keyboard: layout %q %s row %d has %d keys, base has %d",
				l.Name, name, i, len(row), len(l.Rows[i]))
		}
	}
	return nil
}

// Clone returns a deep copy, so callers can mutate a built in layout safely.
func (l Layout) Clone() Layout {
	out := l
	out.Rows = cloneRows(l.Rows)
	out.Shift = cloneRows(l.Shift)
	out.AltGr = cloneRows(l.AltGr)
	if l.DeadKeys != nil {
		out.DeadKeys = make(map[rune]map[rune]rune, len(l.DeadKeys))
		for d, table := range l.DeadKeys {
			inner := make(map[rune]rune, len(table))
			for k, v := range table {
				inner[k] = v
			}
			out.DeadKeys[d] = inner
		}
	}
	return out
}

func cloneRows(rows []Row) []Row {
	if rows == nil {
		return nil
	}
	out := make([]Row, len(rows))
	for i, row := range rows {
		out[i] = append(Row(nil), row...)
	}
	return out
}
