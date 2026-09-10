package keyboard

import "fyne.io/fyne/v2"

// AZERTY_FR is the French layout, including the circumflex and diaeresis dead
// keys and the AltGr level for @, € and the bracket family.
var AZERTY_FR = frLayout()

func init() { MustRegisterLayout(AZERTY_FR) }

func frLayout() Layout {
	deadKeys := map[rune]map[rune]rune{
		'^': accents("aeiouAEIOU", "âêîôûÂÊÎÔÛ"),
		'¨': accents("aeiouyAEIOUY", "äëïöüÿÄËÏÖÜŸ"),
	}

	rows := []Row{
		row("²&é\"'(-è_çà)=", ControlKey(fyne.KeyBackspace, "⌫", 2)),
		row(ControlKey(fyne.KeyTab, "Tab", 1.5), "azertyuiop", DeadKey('^', ""), "$*", SpacerKey(0.5)),
		row(ModifierKey(ModCapsLock, "Caps", 1.75), "qsdfghjklmù", ControlKey(fyne.KeyReturn, "⏎", 2.25)),
		row(ModifierKey(ModShift, "⇧", 2.25), "wxcvbn,;:!", ModifierKey(ModShift, "⇧", 2.75)),
		bottomRow(),
	}

	shift := map[rune]rune{
		'²': '~', '&': '1', 'é': '2', '"': '3', '\'': '4', '(': '5', '-': '6',
		'è': '7', '_': '8', 'ç': '9', 'à': '0', ')': '°', '=': '+',
		'^': '¨', '$': '£', '*': 'µ', 'ù': '%',
		',': '?', ';': '.', ':': '/', '!': '§',
	}

	altGr := map[rune]rune{
		'&': '¹', 'é': '~', '"': '#', '\'': '{', '(': '[', '-': '|',
		'è': '`', '_': '\\', 'à': '@', ')': ']', '=': '}',
		'e': '€',
	}

	return Layout{
		Name:     "fr-FR",
		Title:    "Français (AZERTY)",
		Rows:     rows,
		Shift:    overlayLevel(rows, shift, deadKeys),
		AltGr:    overlayLevel(rows, altGr, deadKeys),
		DeadKeys: deadKeys,
	}
}
