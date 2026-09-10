# Contributing

Thanks for taking the time. Bug reports, layouts for more languages and pull
requests are all welcome.

## Getting started

```
git clone https://github.com/timzifer/fyne-keyboard
cd fyne-keyboard
go test ./...
go run ./cmd/demo
```

On Linux you need the Fyne build dependencies (OpenGL and X11 headers) for the
demo; see the [Fyne getting started guide](https://docs.fyne.io/started/).

## Before opening a pull request

- `gofmt -l .` prints nothing
- `go vet ./...` is clean
- `go test ./...` passes, and new behaviour comes with a test
- exported identifiers have doc comments that say *why*, not just *what*

CI runs the same checks on Linux, macOS and Windows, plus `golangci-lint`, and
builds against both the minimum versions named in `go.mod` (Go 1.22, Fyne
v2.6) and the latest Fyne release. Please do not raise either floor without a
reason — say which API needs it in the pull request.

## Adding a layout

A layout is plain data, so a new language is a small file:

1. Copy `layout_us.go` to `layout_xx.go` and adapt the rows.
2. Keep every row at the same unit total (`rowUnits`) — that is what makes the
   rows line up, since each row is scaled to the full width on its own.
3. Express the shift and AltGr levels with `overlayLevel`; letters get their
   upper case form automatically, so only punctuation and digits need entries.
4. Dead keys go in `DeadKeys`; `accents("aeiou", "âêîôû")` builds a table.
5. Register it in `init` with `MustRegisterLayout`.

The shared tests in `layouts_test.go` check every registered layout, so a new
one is covered as soon as it is registered. Please add a couple of cases for
the characters that are specific to your language.

## Reporting a bug

Please include the Fyne version, the operating system, the layout in use and,
if you can, a small program that reproduces the problem.

## Code of conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md).
