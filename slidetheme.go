package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

func (s *slide) themeBackground() fyne.CanvasObject {
	bg := canvas.NewRectangle(s.parent.theme.Color(theme.ColorNameBackground, theme.VariantLight))
	top := canvas.NewRectangle(s.parent.theme.Color(colorNameHeaderBackground, theme.VariantLight))
	bottom := canvas.NewRectangle(s.parent.theme.Color(colorNameHeaderBackground, theme.VariantLight))
	return container.New(&backgroundLayout{s: s}, bg, top, bottom)
}

type backgroundLayout struct {
	s *slide
}

func (l *backgroundLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	objs[0].Resize(size)

	top := objs[1]
	bottom := objs[2]
	if l.s.variant == headingSlide {
		top.Resize(fyne.NewSize(size.Width, size.Height/4))
		top.Move(fyne.NewPos(0, size.Height*3/8))

		bottom.Hide()
		return
	}

	top.Resize(fyne.NewSize(size.Width, size.Height/6))
	top.Move(fyne.Position{})

	bottomHeight := size.Height / 18
	bottom.Show()
	bottom.Resize(fyne.NewSize(size.Width, bottomHeight))
	bottom.Move(fyne.NewPos(0, size.Height-bottomHeight))
}

func (l *backgroundLayout) MinSize([]fyne.CanvasObject) fyne.Size {
	return fyne.Size{}
}
