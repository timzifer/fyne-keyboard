package keyboard

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestBindShortcutCreatesAMenuEntry(t *testing.T) {
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })

	win := test.NewWindow(widget.NewLabel("x"))
	t.Cleanup(win.Close)

	called := 0
	sh := &desktop.CustomShortcut{KeyName: fyne.KeyK, Modifier: fyne.KeyModifierControl}
	BindShortcut(win, "View", "Keyboard", sh, func() { called++ })

	menu := win.MainMenu()
	if menu == nil || len(menu.Items) != 1 || menu.Items[0].Label != "View" {
		t.Fatalf("main menu = %+v", menu)
	}
	item := menu.Items[0].Items[0]
	if item.Label != "Keyboard" || item.Shortcut != fyne.Shortcut(sh) {
		t.Fatalf("menu item = %+v", item)
	}
	item.Action()
	if called != 1 {
		t.Errorf("the action ran %d times, want 1", called)
	}
}

func TestBindShortcutReusesAndAddsMenus(t *testing.T) {
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })

	win := test.NewWindow(widget.NewLabel("x"))
	t.Cleanup(win.Close)

	sh := func(k fyne.KeyName) *desktop.CustomShortcut {
		return &desktop.CustomShortcut{KeyName: k, Modifier: fyne.KeyModifierControl}
	}
	BindShortcut(win, "View", "Keyboard", sh(fyne.KeyK), func() {})
	BindShortcut(win, "View", "Keypad", sh(fyne.KeyN), func() {})
	BindShortcut(win, "Help", "About", sh(fyne.KeyH), func() {})

	menu := win.MainMenu()
	if len(menu.Items) != 2 {
		t.Fatalf("menu has %d top level entries, want 2", len(menu.Items))
	}
	if len(menu.Items[0].Items) != 2 {
		t.Errorf("the View menu has %d entries, want 2", len(menu.Items[0].Items))
	}
}

func TestBindShortcutIgnoresMissingArguments(t *testing.T) {
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })

	win := test.NewWindow(widget.NewLabel("x"))
	t.Cleanup(win.Close)

	BindShortcut(nil, "View", "Keyboard", &desktop.CustomShortcut{KeyName: fyne.KeyK}, func() {})
	BindShortcut(win, "View", "Keyboard", nil, func() {})
	BindShortcut(win, "View", "Keyboard", &desktop.CustomShortcut{KeyName: fyne.KeyK}, nil)

	if win.MainMenu() != nil {
		t.Error("nothing should have been installed")
	}
}

// A shortcut without a modifier is installed, but the driver never turns a
// bare key press into one, so it is only reachable through the menu.
func TestBindShortcutWithoutModifier(t *testing.T) {
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })

	win := test.NewWindow(widget.NewLabel("x"))
	t.Cleanup(win.Close)

	BindShortcut(win, "View", "Keyboard", &desktop.CustomShortcut{KeyName: fyne.KeyF2}, func() {})
	if win.MainMenu() == nil {
		t.Error("the menu entry should still be created")
	}
}
