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
	s.g.moveToSlide(s.id)
}

func (g *gui) newSlideButton(id int) fyne.CanvasObject {
	border := canvas.NewRectangle(color.Transparent)
	border.StrokeColor = theme.PrimaryColor()
	c, _ := g.s.current.Get()
	if c == id {
		border.StrokeWidth = 2
	} else {
		border.StrokeWidth = 0
	}

	slide := newAspectContainer(newSlide(widget.NewRichTextFromMarkdown(g.s.items[id])), border)
	button := &slideButton{id: id, content: slide, g: g}
	button.ExtendBaseWidget(button)
	return button
}
