package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type slideButton struct {
	widget.BaseWidget
	id      int
	content fyne.CanvasObject
	g       *gui
}

func (s *slideButton) CreateRenderer() fyne.WidgetRenderer {
	num := fmt.Sprintf(" %d", s.id+1)
	return widget.NewSimpleRenderer(container.NewStack(s.content, container.NewVBox(canvas.NewText(num, theme.BackgroundColor()))))
}

func (s *slideButton) Tapped(_ *fyne.PointEvent) {
	s.g.moveToSlide(s.id)
}

func (g *gui) newSlideButton(id int) fyne.CanvasObject {
	slide := newAspectContainer(newSlide(g.s.items[id], g.s))
	button := &slideButton{id: id, content: slide, g: g}
	button.ExtendBaseWidget(button)
	return button
}
