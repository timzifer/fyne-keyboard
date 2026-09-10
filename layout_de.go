package keyboard

import "fyne.io/fyne/v2"

// QWERTZ_DE is the German layout, including the AltGr level and the
// circumflex, acute and grave dead keys.
var QWERTZ_DE = deLayout()

func init() { MustRegisterLayout(QWERTZ_DE) }

func deLayout() Layout {
	deadKeys := map[rune]map[rune]rune{
		'^': accents("aeiouAEIOU", "âêîôûÂÊÎÔÛ"),
		'´': accents("aeiouyAEIOUY", "áéíóúýÁÉÍÓÚÝ"),
		'`': accents("aeiouAEIOU", "àèìòùÀÈÌÒÙ"),
	}

	rows := []Row{
		row(DeadKey('^', ""), "1234567890ß", DeadKey('´', ""), ControlKey(fyne.KeyBackspace, "⌫", 2)),
		row(ControlKey(fyne.KeyTab, "Tab", 1.5), "qwertzuiopü+#", SpacerKey(0.5)),
		row(ModifierKey(ModCapsLock, "Caps", 1.75), "asdfghjklöä", ControlKey(fyne.KeyReturn, "⏎", 2.25)),
		row(ModifierKey(ModShift, "⇧", 2.25), "yxcvbnm,.-", ModifierKey(ModShift, "⇧", 2.75)),
		bottomRow(),
	}

	shift := map[rune]rune{
		'^': '°', '1': '!', '2': '"', '3': '§', '4': '$', '5': '%',
		'6': '&', '7': '/', '8': '(', '9': ')', '0': '=', 'ß': '?', '´': '`',
		'+': '*', '#': '\'', ',': ';', '.': ':', '-': '_',
	}

	altGr := map[rune]rune{
		'1': '¹', '2': '²', '3': '³', '7': '{', '8': '[', '9': ']', '0': '}',
		'ß': '\\', 'q': '@', 'e': '€', 'm': 'µ', '+': '~', '<': '|',
	}

	return Layout{
		Name:     "de-DE",
		Title:    "Deutsch (QWERTZ)",
		Rows:     rows,
		Shift:    overlayLevel(rows, shift, deadKeys),
		AltGr:    overlayLevel(rows, altGr, deadKeys),
		DeadKeys: deadKeys,
	}
}
