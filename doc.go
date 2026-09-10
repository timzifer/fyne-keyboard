// Package keyboard provides an on screen keyboard for [Fyne] applications.
//
// Fyne ships no desktop on screen keyboard: the mobile Keyboardable interface
// only asks iOS or Android for their system keyboard, which leaves desktop,
// touch panel and kiosk applications without one. This package builds the
// keyboard from ordinary Fyne parts, so it works everywhere Fyne does.
//
// # How typing works
//
// Every focusable Fyne widget handles its own text input through
// fyne.Focusable: TypedRune for characters, TypedKey for control keys, with
// cursor, selection and OnChanged already taken care of. The keyboard
// therefore only has to send to whatever holds the canvas focus; it never
// needs to know which widget that is.
//
// Keeping the focus is the one delicate part. Fyne's drivers unfocus the
// current widget when a non focusable object is tapped, so every cap captures
// the focused widget on press (before the unfocus) and the keyboard restores
// it right before each keystroke. Focus is only ever changed from a tap
// handler, never from FocusGained or FocusLost, which the Fyne documentation
// warns can deadlock.
//
// # Attaching the keyboard
//
// [Decorator] docks the keyboard below an application's main widget, either in
// a draggable container.Split or fixed to the bottom edge:
//
//	dec := keyboard.NewDecorator(win, ui, keyboard.WithLayoutNamed("de-DE"))
//	win.SetContent(dec.Content())
//	keyboard.BindShortcut(win, "View", "Toggle keyboard",
//		&desktop.CustomShortcut{KeyName: fyne.KeyK, Modifier: fyne.KeyModifierControl},
//		dec.Toggle)
//
// [BindShortcut] uses the window's main menu because Fyne gives the focused
// widget the first say on key input: a plain key or a canvas shortcut is
// swallowed by the entry the user is typing in.
//
// [Popup] instead floats the keyboard next to a single widget, the way a
// tooltip is anchored, leaving the window layout untouched:
//
//	pop := keyboard.NewPopup(win, keyboard.WithNumpad())
//	pop.ShowFor(amountEntry)
//
// # Styles
//
// A keyboard shows the character block ([StyleText]), a numeric keypad with a
// double zero key ([StyleNumpad]), or both side by side ([StyleFull]).
//
// # Layouts
//
// A [Layout] is plain data: rows of [Key] values with relative widths, plus
// optional shift and AltGr levels and a dead key table. Rows are laid out by
// weight, so staggered rows and wide caps such as Tab, Shift and Space fall
// out of the data instead of pixel positions, and the keyboard scales with the
// window. Layouts can be written in Go, decoded from JSON with [ParseLayout],
// registered with [RegisterLayout] and looked up per language tag with
// [LayoutFor]. Since characters travel as runes, umlauts, accents and ß need
// no special handling.
//
// [Fyne]: https://fyne.io
package keyboard
