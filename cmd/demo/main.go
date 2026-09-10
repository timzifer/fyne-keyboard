// Command demo shows the two ways to attach the on screen keyboard: docked
// below the content with a Decorator, and floating next to a widget with a
// Popup. Ctrl+K toggles the docked keyboard.
package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	keyboard "github.com/timzifer/fyne-keyboard"
)

func main() {
	a := app.New()
	win := a.NewWindow("fyne-keyboard demo")

	text := widget.NewMultiLineEntry()
	text.SetPlaceHolder("Tap here, then press Ctrl+K or the button below")
	text.Wrapping = fyne.TextWrapWord

	amount := widget.NewEntry()
	amount.SetPlaceHolder("Amount – opens a numeric keypad")

	status := widget.NewLabel("")

	popup := keyboard.NewPopup(win, keyboard.WithNumpad())
	popup.Keyboard().OnKey = func(k keyboard.Key) {
		status.SetText("keypad: " + k.Text())
	}

	dec := keyboard.NewDecorator(win, nil,
		keyboard.WithLayoutNamed("de-DE"),
		keyboard.WithOffset(0.55),
	)
	dec.Keyboard().OnLayoutChange = func(l keyboard.Layout) {
		status.SetText("layout: " + l.Title)
	}

	layouts := widget.NewSelect(keyboard.LayoutNames(), func(name string) {
		if err := dec.SetLayout(keyboard.LayoutFor(name)); err != nil {
			status.SetText(fmt.Sprintf("layout error: %v", err))
		}
	})
	layouts.SetSelected("de-DE")

	styles := widget.NewSelect([]string{"text", "numpad", "full"}, func(name string) {
		switch name {
		case "numpad":
			dec.SetStyle(keyboard.StyleNumpad)
		case "full":
			dec.SetStyle(keyboard.StyleFull)
		default:
			dec.SetStyle(keyboard.StyleText)
		}
	})
	styles.SetSelected("text")

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Docked keyboard", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewGridWithColumns(3,
				widget.NewButton("Toggle (Ctrl+K)", dec.Toggle),
				layouts,
				styles,
			),
			widget.NewSeparator(),
			widget.NewLabelWithStyle("Floating keypad", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			amount,
			widget.NewButton("Keypad for the amount", func() { popup.ToggleFor(amount) }),
			status,
		),
		nil, nil, nil,
		text,
	)

	dec.SetContent(content)
	win.SetContent(dec.Content())
	win.Resize(fyne.NewSize(900, 640))

	// A main menu shortcut, not a canvas one: Fyne gives the focused widget
	// the first say, so an entry would swallow a plain key or a canvas
	// shortcut before the toggle ever sees it.
	keyboard.BindShortcut(win, "View", "Toggle keyboard",
		&desktop.CustomShortcut{KeyName: fyne.KeyK, Modifier: fyne.KeyModifierControl},
		dec.Toggle)

	win.ShowAndRun()
}
