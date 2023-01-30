package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

func (s *slide) themeBackground() fyne.CanvasObject {
	bg := canvas.NewRectangle(color.White)
	top := canvas.NewRectangle(color.Gray{Y: 0xC0})
	bottom := canvas.NewRectangle(color.Gray{Y: 0xC0})
	return container.New(&backgroundLayout{}, bg, top, bottom)
}

func (s *slide) themeText(text *canvas.Text) {
	text.Color = color.Black
}

type backgroundLayout struct{}

func (a *backgroundLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	objs[0].Resize(size)

	top := objs[1]
	top.Resize(fyne.NewSize(size.Width, size.Height/6))

	bottom := objs[2]
	bottomHeight := size.Height / 18
	bottom.Resize(fyne.NewSize(size.Width, bottomHeight))
	bottom.Move(fyne.NewPos(0, size.Height-bottomHeight))
}

func (a *backgroundLayout) MinSize([]fyne.CanvasObject) fyne.Size {
	return fyne.Size{}
}
