package main

import (
	"fyne.io/fyne/v2"
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
		y += s.heading.MinSize().Height //+ pad
	}
	subPad := float32(0)
	if s.subheading != nil {
		skip++
		s.subheading.TextSize = theme.TextSubHeadingSize() * scale
		s.subheading.Move(fyne.NewPos(pad, y))
		s.subheading.Refresh()
		subPad = s.subheading.MinSize().Height
	}

	contentSize := size.SubtractWidthHeight(pad*2, size.Height/9*2+subPad+pad*2)
	contentPos := fyne.NewPos(pad, size.Height/6+subPad+pad)
	layoutContent(s.content.Objects[skip:], scale, contentSize, contentPos)
}

func (s *slide) layoutImage(size fyne.Size) {
	for _, o := range s.content.Objects[1:] {
		o.Resize(size)
	}
}
