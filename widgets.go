package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"

	"image/color"
)

func (g *gui) newSlideButton(id int) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.White)
	bg.StrokeColor = theme.PrimaryColor()
	c, _ := g.current.Get()
	if c == id {
		bg.StrokeWidth = 3
	} else {
		bg.StrokeWidth = 0
	}

	t := fmt.Sprintf("Slide %d", id+1)
	title := canvas.NewText(t, theme.BackgroundColor())
	title.TextSize = 8
	slide := newAspectContainer(bg, container.NewPadded(container.NewVBox(title)))
	num := fmt.Sprintf("%d:", id+1)
	return container.NewHBox(container.NewVBox(canvas.NewText(num, theme.ForegroundColor())), slide)
}
