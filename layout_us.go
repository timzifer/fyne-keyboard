package keyboard

import "fyne.io/fyne/v2"

// QWERTY_US is the US English layout.
var QWERTY_US = usLayout()

func init() { MustRegisterLayout(QWERTY_US) }

func usLayout() Layout {
	rows := []Row{
		row("`1234567890-=", ControlKey(fyne.KeyBackspace, "⌫", 2)),
		row(ControlKey(fyne.KeyTab, "Tab", 1.5), `qwertyuiop[]\`, SpacerKey(0.5)),
		row(ModifierKey(ModCapsLock, "Caps", 1.75), "asdfghjkl;'", ControlKey(fyne.KeyReturn, "⏎", 2.25)),
		row(ModifierKey(ModShift, "⇧", 2.25), "zxcvbnm,./", ModifierKey(ModShift, "⇧", 2.75)),
		bottomRow(),
	}

	shift := map[rune]rune{
		'`': '~', '1': '!', '2': '@', '3': '#', '4': '$', '5': '%',
		'6': '^', '7': '&', '8': '*', '9': '(', '0': ')', '-': '_', '=': '+',
		'[': '{', ']': '}', '\\': '|', ';': ':', '\'': '"',
		',': '<', '.': '>', '/': '?',
	}

	return Layout{
		Name:  "en-US",
		Title: "English (QWERTY)",
		Rows:  rows,
		Shift: overlayLevel(rows, shift, nil),
	}
}
