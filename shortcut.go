package keyboard

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// BindShortcut installs a keyboard shortcut that toggles (or otherwise drives)
// the on screen keyboard, as an entry in the window's main menu.
//
// The main menu is used because it is the only place a shortcut reliably
// fires while a text field has the focus, which is exactly the situation an on
// screen keyboard is for. Fyne's key handling gives the focused widget the
// first say: a plain key such as F2 is delivered to the focused Entry and
// swallowed there, and a canvas shortcut registered with
// fyne.Canvas.AddShortcut is only consulted when nothing has the focus. Main
// menu shortcuts are checked before the focused widget, so they still work.
//
// The shortcut must carry a modifier: the desktop driver only turns a key
// press into a [desktop.CustomShortcut] when one is held.
//
//	keyboard.BindShortcut(win, "View", "On screen keyboard",
//		&desktop.CustomShortcut{KeyName: fyne.KeyK, Modifier: fyne.KeyModifierControl},
//		dec.Toggle)
func BindShortcut(win fyne.Window, menuLabel, itemLabel string, sh *desktop.CustomShortcut, action func()) {
	if win == nil || sh == nil || action == nil {
		return
	}
	if sh.Modifier == 0 {
		fyne.LogError("keyboard: a shortcut without a modifier never fires, "+
			"the menu entry will still work", nil)
	}

	item := fyne.NewMenuItem(itemLabel, action)
	item.Shortcut = sh

	menu := win.MainMenu()
	if menu == nil {
		win.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu(menuLabel, item)))
		return
	}
	for _, existing := range menu.Items {
		if existing.Label == menuLabel {
			existing.Items = append(existing.Items, item)
			win.SetMainMenu(menu)
			return
		}
	}
	menu.Items = append(menu.Items, fyne.NewMenu(menuLabel, item))
	win.SetMainMenu(menu)
}
