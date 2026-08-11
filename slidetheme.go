package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

// themeBackground builds the slide chrome: a base fill and the header and footer
// bands. Any backdrop the slide brings of its own covers the base fill and sits
// beneath the bands, so a heading and the footer stay legible over it.
func (s *slide) themeBackground(backdrop ...fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(s.parent.theme.Color(theme.ColorNameBackground, theme.VariantLight))
	objs := append([]fyne.CanvasObject{bg}, backdrop...)

	top := canvas.NewRectangle(s.parent.theme.Color(colorNameHeaderBackground, theme.VariantLight))
	bgCol := s.parent.theme.Color(colorNameFooterBackground, theme.VariantLight)
	if bgCol == color.Transparent {
		bgCol = top.FillColor
	}
	bottom := canvas.NewRectangle(bgCol)
	return container.New(&backgroundLayout{s: s}, append(objs, top, bottom)...)
}

type backgroundLayout struct {
	s *slide
}

func (l *backgroundLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	// the base fill and any backdrop over it cover the whole slide
	for _, o := range objs[:len(objs)-2] {
		o.Resize(size)
	}

	top := objs[len(objs)-2]
	bottom := objs[len(objs)-1]
	if l.s.variant == imageSlide { // a full page image carries no chrome
		top.Hide()
		bottom.Hide()
		return
	}

	top.Show()
	if l.s.variant == headingSlide {
		top.Resize(fyne.NewSize(size.Width, size.Height/4))
		top.Move(fyne.NewPos(0, size.Height*3/8))

		bottom.Hide()
		return
	}

	top.Resize(fyne.NewSize(size.Width, size.Height/6))
	top.Move(fyne.Position{})

	bottomHeight := size.Height / 14
	bottom.Show()
	bottom.Resize(fyne.NewSize(size.Width, bottomHeight))
	bottom.Move(fyne.NewPos(0, size.Height-bottomHeight))
}

func (l *backgroundLayout) MinSize([]fyne.CanvasObject) fyne.Size {
	return fyne.Size{}
}
