package keyboard

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// keyRowLayout arranges the caps of one keyboard row side by side, giving each
// object a share of the available width proportional to its relative unit
// width. Fractional units at the start or end of a row produce the staggered
// look of a real keyboard without any pixel maths in the layout data.
type keyRowLayout struct {
	units []float32
}

var _ fyne.Layout = (*keyRowLayout)(nil)

func (l *keyRowLayout) unitsFor(i int) float32 {
	if i < 0 || i >= len(l.units) || l.units[i] <= 0 {
		return 1
	}
	return l.units[i]
}

func (l *keyRowLayout) total() float32 {
	var sum float32
	for i := range l.units {
		sum += l.unitsFor(i)
	}
	return sum
}

// Layout distributes width proportionally. Positions are accumulated in
// float64 so that rounding does not drift across a long row.
func (l *keyRowLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	pad := theme.Padding()
	sum := l.total()
	if sum <= 0 {
		return
	}
	gaps := pad * float32(len(objects)-1)
	avail := float64(size.Width - gaps)
	if avail < 0 {
		avail = 0
	}

	var x float64
	for i, obj := range objects {
		share := avail * float64(l.unitsFor(i)) / float64(sum)
		next := x + share
		// Snap both edges to the pixel grid so neighbouring caps meet exactly.
		left := float32(int(x + 0.5))
		right := float32(int(next + 0.5))
		obj.Move(fyne.NewPos(left+float32(i)*pad, 0))
		obj.Resize(fyne.NewSize(right-left, size.Height))
		x = next
	}
}

// MinSize is the smallest size at which every cap still fits its own minimum,
// keeping the requested width ratios intact.
func (l *keyRowLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.Size{}
	}
	sum := l.total()
	var perUnit, height float32
	for i, obj := range objects {
		min := obj.MinSize()
		if w := min.Width / l.unitsFor(i); w > perUnit {
			perUnit = w
		}
		if min.Height > height {
			height = min.Height
		}
	}
	gaps := theme.Padding() * float32(len(objects)-1)
	return fyne.NewSize(perUnit*sum+gaps, height)
}
