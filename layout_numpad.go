package keyboard

import (
	"fmt"

	"fyne.io/fyne/v2"
)

// numpadUnits is the relative width of the numeric keypad block.
const numpadUnits = 5

// fullGapUnits is the gap between the character block and the keypad block of
// a full size keyboard.
const fullGapUnits = 0.5

// NUMPAD is a standalone numeric keypad, including a double zero key. It is
// the layout to use for numeric input fields. It is deliberately not in the
// layout registry: it is a style, not a language, so the layout switch key
// does not cycle into it.
var NUMPAD = numpadLayout()

func numpadLayout() Layout {
	return Layout{
		Name:  "numpad",
		Title: "Numpad",
		Rows:  numpadRows(true),
	}
}

// numpadRows is the keypad block: five rows of five units, so it can stand on
// its own or be appended to a character block. Standalone it needs its own
// hide key, attached to a character block it does not.
func numpadRows(standalone bool) []Row {
	last := CharKey('=')
	if standalone {
		last = ActionKey(ActionHide, "Hide", 1)
	}
	return []Row{
		row(
			ControlKey(fyne.KeyBackspace, "⌫", 1),
			ControlKey(fyne.KeyLeft, "◀", 1),
			ControlKey(fyne.KeyRight, "▶", 1),
			ControlKey(fyne.KeyTab, "Tab", 1),
			last,
		),
		row("789", CharKey('/'), CharKey('*')),
		row("456", CharKey('-'), CharKey('+')),
		row("123", CharKey('.'), ControlKey(fyne.KeyReturn, "⏎", 1)),
		row(CharKeyWidth('0', 2), SequenceKey("00", 2), CharKey(',')),
	}
}

// Numpad returns the standalone keypad layout.
func Numpad() Layout { return NUMPAD.Clone() }

// Full returns a full size keyboard: the character block of main with the
// numeric keypad attached on the right.
func Full(main Layout) Layout {
	keypad := Layout{Name: "numpad", Title: "Numpad", Rows: numpadRows(false)}
	full, err := Compose(main, keypad, fullGapUnits)
	if err != nil {
		// Both blocks are five rows tall, so this cannot happen for the built
		// in layouts; a caller's own layout may differ, and then the plain
		// character block is the safer answer.
		fyne.LogError("keyboard: cannot attach the keypad", err)
		return main
	}
	return full
}

// Compose places right next to left, separated by gapUnits of empty space.
// Both layouts must have the same number of rows. Overlay levels and dead key
// tables are merged, so shift and AltGr keep working on the left block.
func Compose(left, right Layout, gapUnits float32) (Layout, error) {
	if len(left.Rows) != len(right.Rows) {
		return Layout{}, fmt.Errorf("keyboard: cannot compose %q (%d rows) with %q (%d rows)",
			left.Name, len(left.Rows), right.Name, len(right.Rows))
	}

	out := Layout{
		Name:  left.Name + "+" + right.Name,
		Title: composeTitle(left, right),
		Rows:  joinRows(left.Rows, right.Rows, SpacerKey(gapUnits)),
	}
	if len(left.Shift) > 0 || len(right.Shift) > 0 {
		out.Shift = joinRows(levelOrBlank(left.Shift, left.Rows), levelOrBlank(right.Shift, right.Rows), Key{})
	}
	if len(left.AltGr) > 0 || len(right.AltGr) > 0 {
		out.AltGr = joinRows(levelOrBlank(left.AltGr, left.Rows), levelOrBlank(right.AltGr, right.Rows), Key{})
	}
	out.DeadKeys = mergeDeadKeys(left.DeadKeys, right.DeadKeys)

	if err := out.Validate(); err != nil {
		return Layout{}, err
	}
	return out, nil
}

func composeTitle(left, right Layout) string {
	switch {
	case left.Title == "":
		return right.Title
	case right.Title == "":
		return left.Title
	default:
		return left.Title + " + " + right.Title
	}
}

func joinRows(left, right []Row, separator Key) []Row {
	out := make([]Row, len(left))
	for i := range left {
		joined := make(Row, 0, len(left[i])+1+len(right[i]))
		joined = append(joined, left[i]...)
		joined = append(joined, separator)
		joined = append(joined, right[i]...)
		out[i] = joined
	}
	return out
}

// levelOrBlank returns an overlay level, or one of empty keys shaped like the
// base rows when the level is not defined.
func levelOrBlank(level, base []Row) []Row {
	if len(level) == len(base) {
		return level
	}
	out := make([]Row, len(base))
	for i, row := range base {
		out[i] = make(Row, len(row))
	}
	return out
}

func mergeDeadKeys(tables ...map[rune]map[rune]rune) map[rune]map[rune]rune {
	var out map[rune]map[rune]rune
	for _, table := range tables {
		for dead, combos := range table {
			if out == nil {
				out = map[rune]map[rune]rune{}
			}
			if out[dead] == nil {
				out[dead] = map[rune]rune{}
			}
			for from, to := range combos {
				out[dead][from] = to
			}
		}
	}
	return out
}
