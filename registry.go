package keyboard

import (
	"sort"
	"strings"
	"sync"
)

var registry = struct {
	sync.RWMutex
	byName map[string]Layout
}{byName: map[string]Layout{}}

// DefaultLayoutName is the layout used when none was requested and no match
// was found. It is set to the first registered layout at init time and can be
// changed with [SetDefaultLayout].
var defaultLayoutName = "en-US"

// RegisterLayout adds a layout to the registry under its Name, replacing any
// layout registered under the same name. It reports an error for a
// structurally invalid layout.
func RegisterLayout(l Layout) error {
	if err := l.Validate(); err != nil {
		return err
	}
	registry.Lock()
	defer registry.Unlock()
	registry.byName[l.Name] = l.Clone()
	return nil
}

// MustRegisterLayout is [RegisterLayout] for package level initialisation; it
// panics on an invalid layout.
func MustRegisterLayout(l Layout) {
	if err := RegisterLayout(l); err != nil {
		panic(err)
	}
}

// LookupLayout returns the layout registered for name. A full tag such as
// "de-DE" is matched first, then any layout whose language prefix matches, so
// "de" finds "de-DE".
func LookupLayout(name string) (Layout, bool) {
	registry.RLock()
	defer registry.RUnlock()

	if l, ok := registry.byName[name]; ok {
		return l.Clone(), true
	}
	lang := strings.ToLower(primaryLanguage(name))
	if lang == "" {
		return Layout{}, false
	}
	for _, key := range sortedNamesLocked() {
		if strings.ToLower(primaryLanguage(key)) == lang {
			l := registry.byName[key]
			return l.Clone(), true
		}
	}
	return Layout{}, false
}

// LayoutFor returns the layout for name, falling back to the default layout
// when name is unknown. Use [LookupLayout] when the miss matters.
func LayoutFor(name string) Layout {
	if l, ok := LookupLayout(name); ok {
		return l
	}
	return defaultLayout()
}

// LayoutNames returns the names of all registered layouts, sorted.
func LayoutNames() []string {
	registry.RLock()
	defer registry.RUnlock()
	return sortedNamesLocked()
}

// Layouts returns all registered layouts, sorted by name.
func Layouts() []Layout {
	registry.RLock()
	defer registry.RUnlock()
	names := sortedNamesLocked()
	out := make([]Layout, 0, len(names))
	for _, name := range names {
		out = append(out, registry.byName[name].Clone())
	}
	return out
}

// SetDefaultLayout selects the layout returned by [LayoutFor] on a miss.
func SetDefaultLayout(name string) {
	registry.Lock()
	defer registry.Unlock()
	defaultLayoutName = name
}

func defaultLayout() Layout {
	registry.RLock()
	name := defaultLayoutName
	l, ok := registry.byName[name]
	if !ok {
		if names := sortedNamesLocked(); len(names) > 0 {
			l, ok = registry.byName[names[0]], true
		}
	}
	registry.RUnlock()
	if ok {
		return l.Clone()
	}
	return Layout{}
}

// nextLayout returns the layout following current in registration order,
// wrapping around at the end.
func nextLayout(current string) (Layout, bool) {
	registry.RLock()
	defer registry.RUnlock()

	names := sortedNamesLocked()
	if len(names) < 2 {
		return Layout{}, false
	}
	for i, name := range names {
		if name == current {
			l := registry.byName[names[(i+1)%len(names)]]
			return l.Clone(), true
		}
	}
	l := registry.byName[names[0]]
	return l.Clone(), true
}

func sortedNamesLocked() []string {
	names := make([]string, 0, len(registry.byName))
	for name := range registry.byName {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func primaryLanguage(tag string) string {
	if i := strings.IndexAny(tag, "-_"); i >= 0 {
		return tag[:i]
	}
	return tag
}
