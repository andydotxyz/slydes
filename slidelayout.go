package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const slideHeight = float32(240)

func (s *slide) layout(size fyne.Size) {
	scale := size.Height / slideHeight

	if s.paragraph == nil {
		s.layoutTitleSlide(size, scale)
	} else {
		s.layoutFallback(scale)
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
		if height > 0 {
			height += theme.InnerPadding() * scale
		}

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

func (s *slide) layoutFallback(scale float32) {
	pad := theme.Padding() * scale
	y := pad
	if s.heading != nil {
		s.heading.TextSize = theme.TextHeadingSize() * scale
		s.heading.Move(fyne.NewPos(pad, pad))
		s.heading.Refresh()
		y += s.heading.MinSize().Height + theme.InnerPadding()*scale
	}
	if s.subheading != nil {
		s.subheading.TextSize = theme.TextSubHeadingSize() * scale
		s.subheading.Move(fyne.NewPos(pad, y))
		s.subheading.Refresh()
		y += s.subheading.MinSize().Height + theme.InnerPadding()*scale
	}
	if s.paragraph != nil {
		s.paragraph.TextSize = theme.TextSize() * scale
		s.paragraph.Move(fyne.NewPos(pad, y))
		s.paragraph.Refresh()
	}
}
