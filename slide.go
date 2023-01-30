package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const slideHeight = float32(180)

type slide struct {
	widget.BaseWidget

	content             *fyne.Container
	bg                  *canvas.Rectangle
	heading, subheading *canvas.Text
	paragraph           *canvas.Text
}

func newSlide(in *widget.RichText) *slide {
	s := &slide{}
	s.ExtendBaseWidget(s)
	s.bg = canvas.NewRectangle(color.White)
	items := []fyne.CanvasObject{s.bg}
	s.heading = nil
	s.subheading = nil
	s.paragraph = nil
	s.addContent(&items, in.Segments)
	s.content = container.NewWithoutLayout(items...)
	return s
}

func (s *slide) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.content)
}

func (s *slide) Resize(size fyne.Size) {
	s.bg.Resize(size)

	scale := size.Height / slideHeight
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
	s.BaseWidget.Resize(size)
}

func (s *slide) MinSize() fyne.Size {
	return fyne.NewSize(80, 45) // TODO de-duplicate
}

func (s *slide) addContent(items *[]fyne.CanvasObject, segs []widget.RichTextSegment) {
	for _, item := range segs {
		switch seg := item.(type) {
		case *widget.TextSegment:
			switch seg.Style {
			case widget.RichTextStyleHeading:
				if s.heading != nil {
					s.heading = nil
				}
				s.heading = canvas.NewText(seg.Text, theme.BackgroundColor())
				s.heading.TextStyle.Bold = true
				*items = append(*items, s.heading)
			case widget.RichTextStyleSubHeading:
				if s.subheading != nil {
					s.subheading = nil
				}
				s.subheading = canvas.NewText(seg.Text, theme.BackgroundColor())
				s.subheading.TextStyle.Bold = true
				*items = append(*items, s.subheading)
			default:
				if s.paragraph != nil {
					continue
				}
				s.paragraph = canvas.NewText(seg.Text, theme.BackgroundColor())
				*items = append(*items, s.paragraph)
			}
		case *widget.ListSegment:
			s.addContent(items, seg.Items)
		case *widget.ParagraphSegment:
			s.addContent(items, seg.Texts)
		}
	}
}

func (s *slide) setSource(rich *widget.RichText) {
	items := []fyne.CanvasObject{s.bg}
	s.heading = nil
	s.subheading = nil
	s.paragraph = nil
	s.addContent(&items, rich.Segments)
	s.content.Objects = items
	s.content.Refresh()
	s.Resize(s.Size())
}
