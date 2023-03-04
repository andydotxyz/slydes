package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

const slideHeight = float32(240)

func (s *slide) layout(size fyne.Size) {
	scale := size.Height / slideHeight

	switch s.variant {
	case headingSlide:
		s.layoutTitleSlide(size, scale)
	case imageSlide:
		s.layoutImage(size)
	default:
		s.layoutFallback(size, scale)
	}
}

func (s *slide) layoutTitleSlide(size fyne.Size, scale float32) {
	height := float32(0)
	if s.heading != nil {
		s.heading.TextSize = theme.TextHeadingSize() * scale
		s.heading.Alignment = fyne.TextAlignCenter

		headHeight := s.heading.MinSize().Height
		height = headHeight
		s.heading.Resize(fyne.NewSize(size.Width, headHeight))
		s.heading.Refresh()
	}
	if s.subheading != nil {
		s.subheading.TextSize = theme.TextSubHeadingSize() * scale
		s.subheading.Alignment = fyne.TextAlignCenter

		subHeight := s.subheading.MinSize().Height
		height += subHeight
		s.subheading.Resize(fyne.NewSize(size.Width, subHeight))
		s.subheading.Refresh()
	}
	y := (size.Height - height) / 2
	if s.heading != nil {
		s.heading.Move(fyne.NewPos(0, y))
	}
	if s.subheading != nil {
		subHeight := s.subheading.MinSize().Height
		s.subheading.Move(fyne.NewPos(0, y+height-subHeight))
	}
}

func (s *slide) layoutFallback(size fyne.Size, scale float32) {
	skip := 1
	pad := theme.Padding() * scale
	y := pad
	if s.heading != nil {
		skip++
		s.heading.TextSize = theme.TextHeadingSize() * scale
		s.heading.Move(fyne.NewPos(pad, pad))
		s.heading.Refresh()
		y += s.heading.MinSize().Height + theme.InnerPadding()*scale
	}
	if s.subheading != nil {
		skip++
		s.subheading.TextSize = theme.TextSubHeadingSize() * scale
		s.subheading.Move(fyne.NewPos(pad, y))
		s.subheading.Refresh()
		y += s.subheading.MinSize().Height + theme.InnerPadding()*scale
	}

	// TODO split/layout not just stack
	for _, o := range s.content.Objects[skip:] {
		switch t := o.(type) {
		case *canvas.Image:
			t.FillMode = canvas.ImageFillContain
			t.SetMinSize(fyne.NewSize(128*scale, 80*scale)) // TODO remove once we layout properly
		case *canvas.Text:
			t.TextSize = theme.TextSize() * scale
		case slideWidget:
			t.setScale(scale)
		}
		o.Move(fyne.NewPos(pad, y))
		o.Resize(o.MinSize())
		y += o.MinSize().Height + theme.Padding()*scale
	}
}

func (s *slide) layoutImage(size fyne.Size) {
	for _, o := range s.content.Objects[1:] {
		o.Resize(size)
	}
}
