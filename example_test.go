package keyboard_test

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"

	keyboard "github.com/timzifer/fyne-keyboard"
)

// The decorator docks the keyboard below the application's main widget and a
// shortcut toggles it.
func ExampleNewDecorator() {
	var win fyne.Window // your window
	var ui fyne.CanvasObject

	dec := keyboard.NewDecorator(win, ui,
		keyboard.WithLayoutNamed("de-DE"),
		keyboard.WithOffset(0.6),
	)
	_ = dec.Content() // hand this to win.SetContent

	// A main menu shortcut still fires while an entry has the focus, which a
	// plain key or a canvas shortcut does not.
	keyboard.BindShortcut(win, "View", "Toggle keyboard",
		&desktop.CustomShortcut{KeyName: fyne.KeyK, Modifier: fyne.KeyModifierControl},
		dec.Toggle)
}

// The popup floats the keyboard next to one widget, without changing the
// window layout. A numeric keypad is a good fit for an amount field.
func ExampleNewPopup() {
	var win fyne.Window
	var amount fyne.CanvasObject

	pop := keyboard.NewPopup(win, keyboard.WithNumpad())
	pop.ShowFor(amount)
}

// Layouts are plain data, so an application can register its own.
func ExampleRegisterLayout() {
	own := keyboard.Layout{
		Name:  "xx-XX",
		Title: "Digits",
		Rows: []keyboard.Row{
			keyboard.CharRow("123"),
			keyboard.CharRow("456"),
		},
	}
	if err := keyboard.RegisterLayout(own); err != nil {
		panic(err)
	}
	fmt.Println(keyboard.LayoutFor("xx-XX").Title)
	// Output: Digits
}

// A key can also type more than one rune, which is how the keypad's double
// zero works.
func ExampleSequenceKey() {
	key := keyboard.SequenceKey("00", 2)
	fmt.Println(key.Text(), key.Units())
	// Output: 00 2
}
