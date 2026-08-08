package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

// maxColumnBullets is how many bullets fit in a single column before the
// content is split across two.
const maxColumnBullets = 6

func layoutContent(objs []fyne.CanvasObject, scale float32, size fyne.Size, pos fyne.Position) {
	splitAt := -1
	for i, o := range objs {
		if _, ok := o.(*canvas.Image); ok {
			splitAt = i
		}
	}

	// With no image to sit alongside, a long bullet list takes the second
	// column instead.
	wrapAt := -1
	if splitAt == -1 {
		wrapAt = columnBreak(objs)
	}

	pad := theme.Padding() * scale
	width := size.Width
	if (splitAt > -1 && len(objs) > 1) || wrapAt > -1 {
		width = (width - pad) / 2
	}
	x := pos.X
	y := pos.Y
	if splitAt == 0 {
		x = x + width + pad
	}

	leftEdge := x
	inline := false
	for i, o := range objs {
		if i == wrapAt {
			leftEdge = pos.X + width + pad
			x = leftEdge
			y = pos.Y
			inline = false
		}

		switch t := o.(type) {
		case *canvas.Text:
			t.TextSize = theme.TextSize() * scale

			if len(t.Text) > 0 && t.Text[len(t.Text)-1] != '\r' {
				inline = true
			}
		case slideWidget:
			t.setScale(scale)
		case *fyne.Container:
			if len(t.Objects) == 2 {
				if t, ok := t.Objects[1].(*canvas.Text); ok {
					t.TextSize = theme.TextSize() * scale
					inline = true
				}
			}
		}

		if splitAt == i {
			o.Resize(fyne.NewSize(width, size.Height))
			if splitAt == 0 {
				o.Move(fyne.NewPos(pos.X, pos.Y))
			} else {
				o.Move(fyne.NewPos(x+width+pad, pos.Y))
			}
		} else {
			o.Move(fyne.NewPos(x, y))
			if inline {
				o.Resize(o.MinSize())
				x += o.MinSize().Width

				inline = false
			} else {
				o.Resize(fyne.NewSize(width, o.MinSize().Height))
				x = leftEdge
				y += o.MinSize().Height + pad
			}
		}
	}
}

// columnBreak returns the index of the first object of a second column, or
// -1 if the content fits in one. The bullets are shared as evenly as possible.
func columnBreak(objs []fyne.CanvasObject) int {
	total := 0
	for _, o := range objs {
		if _, ok := o.(*bullet); ok {
			total++
		}
	}
	if total <= maxColumnBullets {
		return -1
	}

	half := (total + 1) / 2
	seen := 0
	for i, o := range objs {
		if _, ok := o.(*bullet); !ok {
			continue
		}
		if seen == half {
			return i
		}
		seen++
	}
	return -1
}
