package keyboard

import (
	"math"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

func rects(n int) []fyne.CanvasObject {
	out := make([]fyne.CanvasObject, n)
	for i := range out {
		out[i] = canvas.NewRectangle(nil)
	}
	return out
}

func TestKeyRowLayoutSharesWidthByUnits(t *testing.T) {
	l := &keyRowLayout{units: []float32{1, 2, 1}}
	objects := rects(3)
	size := fyne.NewSize(404, 40)

	l.Layout(objects, size)

	pad := theme.Padding()
	avail := size.Width - 2*pad
	want := []float32{avail * 0.25, avail * 0.5, avail * 0.25}
	for i, obj := range objects {
		if math.Abs(float64(obj.Size().Width-want[i])) > 1 {
			t.Errorf("object %d width = %v, want about %v", i, obj.Size().Width, want[i])
		}
		if obj.Size().Height != size.Height {
			t.Errorf("object %d height = %v, want %v", i, obj.Size().Height, size.Height)
		}
	}
}

func TestKeyRowLayoutFillsTheRowExactly(t *testing.T) {
	l := &keyRowLayout{units: []float32{1.5, 1, 1, 1, 1, 1, 2.25, 0.5}}
	objects := rects(len(l.units))
	size := fyne.NewSize(777, 40)

	l.Layout(objects, size)

	last := objects[len(objects)-1]
	right := last.Position().X + last.Size().Width
	if math.Abs(float64(right-size.Width)) > 1 {
		t.Errorf("row ends at %v, want %v", right, size.Width)
	}

	pad := theme.Padding()
	for i := 1; i < len(objects); i++ {
		prev := objects[i-1]
		gap := objects[i].Position().X - (prev.Position().X + prev.Size().Width)
		if math.Abs(float64(gap-pad)) > 1 {
			t.Errorf("gap before object %d = %v, want %v", i, gap, pad)
		}
	}
}

func TestKeyRowLayoutHandlesMissingUnits(t *testing.T) {
	l := &keyRowLayout{units: []float32{0, -1}}
	objects := rects(2)

	l.Layout(objects, fyne.NewSize(100, 10))

	if objects[0].Size().Width != objects[1].Size().Width {
		t.Errorf("non positive units should count as one: %v vs %v",
			objects[0].Size().Width, objects[1].Size().Width)
	}
}

func TestKeyRowLayoutMinSizeKeepsRatios(t *testing.T) {
	wide := canvas.NewRectangle(nil)
	wide.SetMinSize(fyne.NewSize(40, 20))
	narrow := canvas.NewRectangle(nil)
	narrow.SetMinSize(fyne.NewSize(10, 30))

	l := &keyRowLayout{units: []float32{2, 1}}
	min := l.MinSize([]fyne.CanvasObject{wide, narrow})

	// The wide object needs 20 per unit, so three units need 60 plus the gap.
	want := fyne.NewSize(60+theme.Padding(), 30)
	if math.Abs(float64(min.Width-want.Width)) > 0.01 || min.Height != want.Height {
		t.Errorf("MinSize() = %v, want %v", min, want)
	}
}

func TestKeyRowLayoutEmpty(t *testing.T) {
	l := &keyRowLayout{}
	l.Layout(nil, fyne.NewSize(10, 10))
	if got := l.MinSize(nil); got != (fyne.Size{}) {
		t.Errorf("MinSize(nil) = %v, want zero", got)
	}
}
