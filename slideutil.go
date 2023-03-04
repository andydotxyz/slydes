package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

func layoutContent(objs []fyne.CanvasObject, scale float32, size fyne.Size, pos fyne.Position) {
	splitAt := -1
	for i, o := range objs {
		if _, ok := o.(*canvas.Image); ok {
			splitAt = i
		}
	}

	pad := theme.Padding() * scale
	width := size.Width
	if splitAt > -1 && len(objs) > 1 {
		width = (width - pad) / 2
	}
	x := pos.X
	y := pos.Y
	if splitAt == 0 {
		x = x + width + pad
	}
	for i, o := range objs {
		switch t := o.(type) {
		case *canvas.Text:
			t.TextSize = theme.TextSize() * scale
		case slideWidget:
			t.setScale(scale)
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
			o.Resize(fyne.NewSize(width, o.MinSize().Height))
			y += o.MinSize().Height + pad
		}
	}
}
