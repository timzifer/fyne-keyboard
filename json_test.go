package keyboard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
)

func TestLayoutJSONRoundTrip(t *testing.T) {
	for _, original := range Layouts() {
		t.Run(original.Name, func(t *testing.T) {
			data, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("Marshal() = %v", err)
			}
			decoded, err := ParseLayout(data)
			if err != nil {
				t.Fatalf("ParseLayout() = %v", err)
			}
			if !reflect.DeepEqual(original, decoded) {
				t.Error("layout changed while round tripping through JSON")
			}
		})
	}
}

func TestParseLayoutInfersKeyTypes(t *testing.T) {
	const src = `{
		"name": "test",
		"rows": [[
			{"rune": "a"},
			{"key": "BackSpace", "label": "back", "width": 2},
			{"modifier": "shift", "label": "Shift"},
			{"action": "hide", "label": "Hide"},
			{}
		]]
	}`

	l, err := ParseLayout([]byte(src))
	if err != nil {
		t.Fatalf("ParseLayout() = %v", err)
	}

	want := Row{
		{Rune: 'a', Type: KeyChar},
		{Key: fyne.KeyBackspace, Label: "back", Width: 2, Type: KeyControl},
		{Label: "Shift", Type: KeyModifier, Mod: ModShift},
		{Label: "Hide", Type: KeyAction, Action: ActionHide},
		{Type: KeySpacer},
	}
	if !reflect.DeepEqual(l.Rows[0], want) {
		t.Errorf("rows = %+v\nwant %+v", l.Rows[0], want)
	}
}

func TestParseLayoutErrors(t *testing.T) {
	cases := map[string]string{
		"broken json":     `{`,
		"multi rune":      `{"name":"t","rows":[[{"rune":"ab"}]]}`,
		"unknown type":    `{"name":"t","rows":[[{"rune":"a","type":"magic"}]]}`,
		"unknown mod":     `{"name":"t","rows":[[{"modifier":"hyper"}]]}`,
		"unknown action":  `{"name":"t","rows":[[{"action":"explode"}]]}`,
		"invalid layout":  `{"name":"t","rows":[]}`,
		"dead key string": `{"name":"t","rows":[[{"rune":"a"}]],"deadKeys":{"^^":{"a":"â"}}}`,
		"dead value":      `{"name":"t","rows":[[{"rune":"a"}]],"deadKeys":{"^":{"a":"ab"}}}`,
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseLayout([]byte(src)); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestParseLayoutDeadKeys(t *testing.T) {
	const src = `{"name":"t","rows":[[{"rune":"^","type":"dead"},{"rune":"a"}]],
		"deadKeys":{"^":{"a":"â"}}}`

	l, err := ParseLayout([]byte(src))
	if err != nil {
		t.Fatalf("ParseLayout() = %v", err)
	}
	if l.Rows[0][0].Type != KeyDead {
		t.Errorf("key type = %v, want KeyDead", l.Rows[0][0].Type)
	}
	if got, ok := l.Combine('^', 'a'); !ok || got != 'â' {
		t.Errorf("Combine() = %q, %v; want â, true", got, ok)
	}
}

func TestReadAndLoadLayout(t *testing.T) {
	data, err := json.Marshal(QWERTZ_DE)
	if err != nil {
		t.Fatalf("Marshal() = %v", err)
	}

	if l, err := ReadLayout(strings.NewReader(string(data))); err != nil || l.Name != "de-DE" {
		t.Errorf("ReadLayout() = %q, %v", l.Name, err)
	}

	path := filepath.Join(t.TempDir(), "de-DE.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if l, err := LoadLayoutFile(path); err != nil || l.Name != "de-DE" {
		t.Errorf("LoadLayoutFile() = %q, %v", l.Name, err)
	}
	if _, err := LoadLayoutFile(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Error("expected an error for a missing file")
	}
}
