# Security Policy

## Supported versions

The latest released version is supported. Fixes are released as new patch
versions.

## Reporting a vulnerability

Please report security issues privately through GitHub's
["Report a vulnerability"](https://github.com/timzifer/fyne-keyboard/security/advisories/new)
form rather than in a public issue. Include what an attacker could do and, if
possible, a small reproducer.

You can expect an acknowledgement within a week.

## Scope

This package renders keys and forwards them to the focused Fyne widget. It does
not store, log or transmit what is typed. Note that an on screen keyboard is
visible on screen by design: it is not a defence against shoulder surfing, and
password entries show the same caps as any other field.
