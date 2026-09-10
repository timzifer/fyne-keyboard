package keyboard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"fyne.io/fyne/v2"
)

// Layouts round trip through JSON so they can ship as data files or be edited
// without recompiling. Runes are encoded as one character strings, key types
// and modifiers as lower case names.

var keyTypeNames = map[KeyType]string{
	KeyChar:     "char",
	KeyControl:  "control",
	KeyModifier: "modifier",
	KeySpacer:   "spacer",
	KeyDead:     "dead",
	KeyAction:   "action",
}

var modifierNames = map[Modifier]string{
	ModShift:    "shift",
	ModCapsLock: "caps",
	ModAltGr:    "altgr",
}

var actionNames = map[Action]string{
	ActionHide:       "hide",
	ActionNextLayout: "nextLayout",
}

type keyJSON struct {
	Rune   string  `json:"rune,omitempty"`
	Key    string  `json:"key,omitempty"`
	Label  string  `json:"label,omitempty"`
	Width  float32 `json:"width,omitempty"`
	Type   string  `json:"type,omitempty"`
	Mod    string  `json:"modifier,omitempty"`
	Action string  `json:"action,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (k Key) MarshalJSON() ([]byte, error) {
	out := keyJSON{
		Key:    string(k.Key),
		Label:  k.Label,
		Width:  k.Width,
		Type:   keyTypeNames[k.Type],
		Mod:    modifierNames[k.Mod],
		Action: actionNames[k.Action],
	}
	if k.Rune != 0 {
		out.Rune = string(k.Rune)
	}
	return json.Marshal(out)
}

// UnmarshalJSON implements json.Unmarshaler. The key type may be omitted; it
// is then inferred from the fields that are present.
func (k *Key) UnmarshalJSON(data []byte) error {
	var in keyJSON
	if err := json.Unmarshal(data, &in); err != nil {
		return err
	}

	out := Key{Key: fyne.KeyName(in.Key), Label: in.Label, Width: in.Width}
	if in.Rune != "" {
		r, size := utf8.DecodeRuneInString(in.Rune)
		if r == utf8.RuneError || size != len(in.Rune) {
			return fmt.Errorf("keyboard: %q is not a single rune", in.Rune)
		}
		out.Rune = r
	}

	var err error
	if out.Mod, err = lookupName("modifier", in.Mod, modifierNames, ModNone); err != nil {
		return err
	}
	if out.Action, err = lookupName("action", in.Action, actionNames, ActionNone); err != nil {
		return err
	}
	if in.Type == "" {
		out.Type = inferKeyType(out)
	} else if out.Type, err = lookupName("key type", in.Type, keyTypeNames, KeyChar); err != nil {
		return err
	}

	*k = out
	return nil
}

func inferKeyType(k Key) KeyType {
	switch {
	case k.Action != ActionNone:
		return KeyAction
	case k.Mod != ModNone:
		return KeyModifier
	case k.Key != "":
		return KeyControl
	case k.Rune == 0 && k.Label == "":
		return KeySpacer
	default:
		return KeyChar
	}
}

func lookupName[T comparable](kind, name string, names map[T]string, zero T) (T, error) {
	if name == "" {
		return zero, nil
	}
	for value, candidate := range names {
		if candidate == name {
			return value, nil
		}
	}
	return zero, fmt.Errorf("keyboard: unknown %s %q", kind, name)
}

type layoutJSON struct {
	Name     string                       `json:"name"`
	Title    string                       `json:"title,omitempty"`
	Rows     []Row                        `json:"rows"`
	Shift    []Row                        `json:"shift,omitempty"`
	AltGr    []Row                        `json:"altgr,omitempty"`
	DeadKeys map[string]map[string]string `json:"deadKeys,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (l Layout) MarshalJSON() ([]byte, error) {
	out := layoutJSON{Name: l.Name, Title: l.Title, Rows: l.Rows, Shift: l.Shift, AltGr: l.AltGr}
	if len(l.DeadKeys) > 0 {
		out.DeadKeys = make(map[string]map[string]string, len(l.DeadKeys))
		for dead, table := range l.DeadKeys {
			inner := make(map[string]string, len(table))
			for in, result := range table {
				inner[string(in)] = string(result)
			}
			out.DeadKeys[string(dead)] = inner
		}
	}
	return json.Marshal(out)
}

// UnmarshalJSON implements json.Unmarshaler.
func (l *Layout) UnmarshalJSON(data []byte) error {
	var in layoutJSON
	if err := json.Unmarshal(data, &in); err != nil {
		return err
	}

	out := Layout{Name: in.Name, Title: in.Title, Rows: in.Rows, Shift: in.Shift, AltGr: in.AltGr}
	if len(in.DeadKeys) > 0 {
		out.DeadKeys = make(map[rune]map[rune]rune, len(in.DeadKeys))
		for dead, table := range in.DeadKeys {
			d, err := singleRune(dead)
			if err != nil {
				return err
			}
			inner := make(map[rune]rune, len(table))
			for from, to := range table {
				f, err := singleRune(from)
				if err != nil {
					return err
				}
				t, err := singleRune(to)
				if err != nil {
					return err
				}
				inner[f] = t
			}
			out.DeadKeys[d] = inner
		}
	}

	*l = out
	return nil
}

func singleRune(s string) (rune, error) {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError || size != len(s) {
		return 0, fmt.Errorf("keyboard: %q is not a single rune", s)
	}
	return r, nil
}

// ParseLayout decodes a layout from JSON and validates it.
func ParseLayout(data []byte) (Layout, error) {
	var l Layout
	if err := json.Unmarshal(data, &l); err != nil {
		return Layout{}, err
	}
	if err := l.Validate(); err != nil {
		return Layout{}, err
	}
	return l, nil
}

// ReadLayout decodes a layout from a JSON stream and validates it.
func ReadLayout(r io.Reader) (Layout, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return Layout{}, err
	}
	return ParseLayout(data)
}

// LoadLayoutFile reads and validates a layout from a JSON file.
func LoadLayoutFile(path string) (Layout, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Layout{}, err
	}
	return ParseLayout(data)
}
