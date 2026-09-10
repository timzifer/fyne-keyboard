package keyboard

import "testing"

func testLayout(name string) Layout {
	return Layout{Name: name, Title: name, Rows: []Row{{CharKey('a')}}}
}

// withRegistry swaps in an isolated registry for the duration of a test.
func withRegistry(t *testing.T, layouts ...Layout) {
	t.Helper()
	registry.Lock()
	saved := registry.byName
	savedDefault := defaultLayoutName
	registry.byName = map[string]Layout{}
	registry.Unlock()

	t.Cleanup(func() {
		registry.Lock()
		registry.byName = saved
		defaultLayoutName = savedDefault
		registry.Unlock()
	})

	for _, l := range layouts {
		MustRegisterLayout(l)
	}
}

func TestRegisterRejectsInvalidLayout(t *testing.T) {
	withRegistry(t)
	if err := RegisterLayout(Layout{Rows: []Row{{CharKey('a')}}}); err == nil {
		t.Error("expected an error for a layout without a name")
	}
	if len(LayoutNames()) != 0 {
		t.Error("an invalid layout must not be registered")
	}
}

func TestRegisterStoresACopy(t *testing.T) {
	withRegistry(t)
	l := testLayout("xx-XX")
	MustRegisterLayout(l)

	l.Rows[0][0] = CharKey('z')
	stored, _ := LookupLayout("xx-XX")
	if stored.Rows[0][0].Rune != 'a' {
		t.Error("the registry kept a reference instead of a copy")
	}
}

func TestLookupMatchesLanguagePrefix(t *testing.T) {
	withRegistry(t, testLayout("de-DE"), testLayout("en-US"))

	if l, ok := LookupLayout("de"); !ok || l.Name != "de-DE" {
		t.Errorf(`LookupLayout("de") = %q, %v; want de-DE, true`, l.Name, ok)
	}
	if l, ok := LookupLayout("de_AT"); !ok || l.Name != "de-DE" {
		t.Errorf(`LookupLayout("de_AT") = %q, %v; want de-DE, true`, l.Name, ok)
	}
	if _, ok := LookupLayout("ja-JP"); ok {
		t.Error("unknown languages must not match")
	}
}

func TestLayoutForFallsBackToDefault(t *testing.T) {
	withRegistry(t, testLayout("de-DE"), testLayout("en-US"))
	SetDefaultLayout("de-DE")

	if got := LayoutFor("ja-JP").Name; got != "de-DE" {
		t.Errorf("LayoutFor(unknown) = %q, want the default de-DE", got)
	}
	if got := LayoutFor("en-US").Name; got != "en-US" {
		t.Errorf("LayoutFor(en-US) = %q, want en-US", got)
	}

	SetDefaultLayout("nope")
	if got := LayoutFor("ja-JP").Name; got != "de-DE" {
		t.Errorf("a missing default should fall back to the first layout, got %q", got)
	}
}

func TestLayoutForEmptyRegistry(t *testing.T) {
	withRegistry(t)
	if got := LayoutFor("de-DE"); got.Name != "" {
		t.Errorf("LayoutFor on an empty registry = %+v, want the zero layout", got)
	}
}

func TestNextLayoutWraps(t *testing.T) {
	withRegistry(t, testLayout("aa"), testLayout("bb"), testLayout("cc"))

	for _, tc := range []struct{ from, want string }{
		{"aa", "bb"},
		{"cc", "aa"},
		{"unknown", "aa"},
	} {
		got, ok := nextLayout(tc.from)
		if !ok || got.Name != tc.want {
			t.Errorf("nextLayout(%q) = %q, %v; want %q", tc.from, got.Name, ok, tc.want)
		}
	}
}

func TestNextLayoutNeedsTwo(t *testing.T) {
	withRegistry(t, testLayout("only"))
	if _, ok := nextLayout("only"); ok {
		t.Error("a single layout cannot be switched away from")
	}
}

func TestBuiltInLayoutsAreRegistered(t *testing.T) {
	for _, name := range []string{"de-DE", "en-US", "fr-FR"} {
		if _, ok := LookupLayout(name); !ok {
			t.Errorf("built in layout %q is not registered", name)
		}
	}
}
