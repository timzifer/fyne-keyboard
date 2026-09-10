# fyne-keyboard

[![CI](https://github.com/timzifer/fyne-keyboard/actions/workflows/ci.yml/badge.svg)](https://github.com/timzifer/fyne-keyboard/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/timzifer/fyne-keyboard.svg)](https://pkg.go.dev/github.com/timzifer/fyne-keyboard)
[![Go Report Card](https://goreportcard.com/badge/github.com/timzifer/fyne-keyboard)](https://goreportcard.com/report/github.com/timzifer/fyne-keyboard)

An on screen keyboard for [Fyne](https://fyne.io) applications — for desktops
without a hardware keyboard, touch panels and kiosks.

Fyne has no built in desktop OSK: the mobile `Keyboardable` interface only asks
iOS or Android for their system keyboard. This package builds one from ordinary
Fyne parts, with no dependencies beyond Fyne itself, and works everywhere Fyne
does.

```go
import keyboard "github.com/timzifer/fyne-keyboard"
```

```
go get github.com/timzifer/fyne-keyboard
```

## Two ways to attach it

**Docked** — the keyboard takes a share of the window, below your main widget:

```go
dec := keyboard.NewDecorator(win, ui, keyboard.WithLayoutNamed("de-DE"))
win.SetContent(dec.Content())

keyboard.BindShortcut(win, "View", "Toggle keyboard",
    &desktop.CustomShortcut{KeyName: fyne.KeyK, Modifier: fyne.KeyModifierControl},
    dec.Toggle)
```

`BindShortcut` puts the shortcut in the window's main menu on purpose. Fyne
gives the focused widget the first say on key input, so a plain key such as F2
is delivered to the focused entry and swallowed there, and a canvas shortcut
(`Canvas().AddShortcut`) is only consulted when *nothing* has the focus —
which is never the case while someone is typing. Main menu shortcuts are
checked first, and they need a modifier: the desktop driver only builds a
`desktop.CustomShortcut` when one is held.

**Floating** — the keyboard is anchored next to one widget like a tooltip, and
the window layout does not move at all:

```go
pop := keyboard.NewPopup(win, keyboard.WithNumpad())
amount.OnTapped = func() { pop.ShowFor(amount) } // or pop.ToggleFor(amount)
```

Run `go run ./cmd/demo` to see both, plus live layout and style switching.

## Styles

| Style | What it shows |
| --- | --- |
| `StyleText` (default) | the character block |
| `StyleNumpad` | a numeric keypad, including a `00` key |
| `StyleFull` | character block with the keypad attached on the right |

```go
dec := keyboard.NewDecorator(win, ui, keyboard.WithFullKeyboard())
dec.SetStyle(keyboard.StyleNumpad) // switches at run time
```

## Options

| Option | Effect |
| --- | --- |
| `WithLayout(Layout)` / `WithLayoutNamed(string)` | the layout to start with |
| `WithStyle(Style)`, `WithNumpad()`, `WithFullKeyboard()` | which blocks to show |
| `WithDockMode(DockSplit \| DockBorder)` | draggable divider, or fixed to the bottom edge (`Decorator`) |
| `WithOffset(float64)` | split ratio while shown (`Decorator`) |
| `WithAutoShow(bool)` | open automatically when a text input takes the focus |
| `WithAutoShowInterval(time.Duration)` | how often the focus is sampled |
| `WithPopupSize(fyne.Size)`, `WithPopupGap(float32)` | size and anchor distance (`Popup`) |

## Layouts and i18n

A layout is plain data — the language *is* the layout, not a patch on top of
one. Built in: `QWERTZ_DE`, `QWERTY_US`, `AZERTY_FR` and the `NUMPAD` keypad.

```go
kb := dec.Keyboard()
kb.SetLayout(keyboard.LayoutFor("fr-FR"))   // live switch, "fr" also matches
keyboard.RegisterLayout(myLayout)           // bring your own, no fork needed
layout, err := keyboard.LoadLayoutFile("kiosk.json")
```

Rows are described by weight, not by pixels:

```go
row(ModifierKey(ModShift, "⇧", 2.25), "yxcvbnm,.-", ModifierKey(ModShift, "⇧", 2.75))
```

Each row is scaled to the full keyboard width and shares it out in proportion
to the key widths, so staggered rows, wide caps and half unit offsets follow
from the data, and everything scales with the window. Keep the unit total equal
across rows — that is what makes them line up.

Levels and dead keys are part of the layout: a `Shift` and an `AltGr` overlay
(letters fall back to their upper case form automatically) and a dead key table
so `^` + `a` becomes `â`. Text travels as runes, so umlauts, accents and ß need
no special handling.

Layouts round trip through JSON, which is handy for kiosk deployments that
should not need a rebuild:

```json
{
  "name": "xx-XX",
  "title": "Digits",
  "rows": [[{"rune": "1"}, {"rune": "2"}, {"key": "BackSpace", "label": "⌫", "width": 2}]]
}
```

## How it keeps the focus

Every focusable Fyne widget handles its own input through `fyne.Focusable`, so
the keyboard just sends `TypedRune` and `TypedKey` to whatever has the canvas
focus — cursor, selection and `OnChanged` come for free.

The subtle part is that Fyne's drivers unfocus the current widget as soon as a
non focusable object is tapped. Every cap therefore captures the focused widget
on press, *before* that unfocus happens, and the keyboard restores it right
before each keystroke. Caps are not focusable themselves, and focus is only
ever changed from a tap handler — never from `FocusGained`/`FocusLost`, which
the Fyne docs warn can deadlock.

Two things to know about `Popup`, both consequences of how Fyne handles
overlays: while it is open it is the only layer that receives taps, so a tap
outside dismisses it rather than reaching the widget below, and it pins the
widget it was opened for instead of following the focus.

## Requirements

Go 1.24+ and Fyne v2.8. Desktop and touch, no external dependencies.

## Contributing

Issues and pull requests are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)
