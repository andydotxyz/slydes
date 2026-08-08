package main

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// bulletPositions renders a slide of the given markdown at 16:9 and returns the
// laid out position of each bullet, in order.
func bulletPositions(t *testing.T, md string) []fyne.Position {
	t.Helper()
	test.NewTempApp(t)

	s := newSlide(md, 0, newSlides())
	s.Resize(fyne.NewSize(1600, 900))

	var pos []fyne.Position
	for _, o := range s.content.Objects {
		if b, ok := o.(*bullet); ok {
			pos = append(pos, b.Position())
		}
	}
	return pos
}

func TestLayoutSingleColumn(t *testing.T) {
	pos := bulletPositions(t, "# Head\n\n- a\n- b\n- c\n- d\n- e\n- f\n")

	if len(pos) != 6 {
		t.Fatalf("expected 6 bullets, got %d", len(pos))
	}
	for i, p := range pos {
		if p.X != pos[0].X {
			t.Errorf("bullet %d should share the first column, got x %g want %g", i, p.X, pos[0].X)
		}
		if i > 0 && p.Y <= pos[i-1].Y {
			t.Errorf("bullet %d should sit below bullet %d, got y %g and %g", i, i-1, p.Y, pos[i-1].Y)
		}
	}
}

func TestLayoutTwoColumns(t *testing.T) {
	pos := bulletPositions(t, "# Head\n\n- a\n- b\n- c\n- d\n- e\n- f\n- g\n")

	if len(pos) != 7 {
		t.Fatalf("expected 7 bullets, got %d", len(pos))
	}
	// 7 bullets split 4 / 3, the extra one going to the first column.
	for i, p := range pos[1:4] {
		if p.X != pos[0].X {
			t.Errorf("bullet %d should be in the first column, got x %g want %g", i+1, p.X, pos[0].X)
		}
	}
	if pos[4].X <= pos[0].X {
		t.Fatalf("bullet 4 should start the second column, got x %g", pos[4].X)
	}
	if pos[4].Y != pos[0].Y {
		t.Errorf("the second column should start at the top, got y %g want %g", pos[4].Y, pos[0].Y)
	}
	for i, p := range pos[5:] {
		if p.X != pos[4].X {
			t.Errorf("bullet %d should be in the second column, got x %g want %g", i+5, p.X, pos[4].X)
		}
	}
}
