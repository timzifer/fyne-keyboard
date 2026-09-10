package keyboard

import "fyne.io/fyne/v2"

// The built in layouts share one geometry: five rows of 15 relative units.
// Keeping the row totals equal is what makes the rows line up, since every row
// is scaled to the full keyboard width on its own.
const rowUnits = 15

// bottomRow is the control row shared by all built in layouts.
func bottomRow() Row {
	return Row{
		ActionKey(ActionNextLayout, "🌐", 1.75),
		ModifierKey(ModAltGr, "AltGr", 1.5),
		Key{Rune: ' ', Label: "Space", Width: 6.5, Type: KeyChar},
		ControlKey(fyne.KeyLeft, "◀", 1.25),
		ControlKey(fyne.KeyRight, "▶", 1.25),
		ActionKey(ActionHide, "Hide", 2.75),
	}
}

// overlayLevel derives a shift or AltGr level from the base rows using a rune
// map. Every position the map does not mention stays empty, which makes the
// resolver fall back to the base key (or to its upper case form for letters).
// Runes that are dead keys in the layout produce dead keys on the overlay too.
func overlayLevel(rows []Row, subs map[rune]rune, deadKeys map[rune]map[rune]rune) []Row {
	out := make([]Row, len(rows))
	for i, row := range rows {
		level := make(Row, len(row))
		for j, key := range row {
			r, ok := subs[key.Rune]
			if !ok || key.Rune == 0 {
				continue
			}
			if _, dead := deadKeys[r]; dead {
				level[j] = DeadKey(r, "")
				continue
			}
			level[j] = CharKey(r)
		}
		out[i] = level
	}
	return out
}

// accents builds a dead key table: every rune of base combines with the rune
// at the same index in combined.
func accents(base, combined string) map[rune]rune {
	from := []rune(base)
	to := []rune(combined)
	table := make(map[rune]rune, len(from))
	for i, r := range from {
		if i >= len(to) {
			break
		}
		table[r] = to[i]
	}
	return table
}

// row concatenates keys and rows into a single row, so a layout can be spelled
// out as "this control key, then these characters, then that one".
func row(parts ...any) Row {
	var out Row
	for _, part := range parts {
		switch p := part.(type) {
		case Key:
			out = append(out, p)
		case Row:
			out = append(out, p...)
		case string:
			out = append(out, CharRow(p)...)
		default:
			panic("keyboard: unsupported row part")
		}
	}
	return out
}
