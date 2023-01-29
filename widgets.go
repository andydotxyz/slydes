package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"image/color"
)

type slideButton struct {
	widget.BaseWidget
	id      int
	content fyne.CanvasObject
	g       *gui
}

func (s *slideButton) CreateRenderer() fyne.WidgetRenderer {
	num := fmt.Sprintf("%d:", s.id+1)
	return widget.NewSimpleRenderer(container.NewHBox(container.NewVBox(canvas.NewText(num, theme.ForegroundColor())), s.content))
}

func (s *slideButton) Tapped(_ *fyne.PointEvent) {
	s.g.s.current.Set(s.id)
}

func (g *gui) newSlideButton(id int) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.White)
	bg.StrokeColor = theme.PrimaryColor()
	c, _ := g.s.current.Get()
	if c == id {
		bg.StrokeWidth = 3
	} else {
		bg.StrokeWidth = 0
	}

	t := fmt.Sprintf("Slide %d", id+1)
	title := canvas.NewText(t, theme.BackgroundColor())
	title.TextSize = 8
	slide := newAspectContainer(bg, container.NewPadded(container.NewVBox(title)))

	button := &slideButton{id: id, content: slide, g: g}
	button.ExtendBaseWidget(button)
	return button
}
