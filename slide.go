package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type slideType int

const (
	otherSlide slideType = iota
	headingSlide
	imageSlide
)

type slide struct {
	widget.BaseWidget
	variant slideType

	content             *fyne.Container
	bg                  fyne.CanvasObject
	heading, subheading *canvas.Text
}

func newSlide(in *widget.RichText) *slide {
	s := &slide{}
	s.ExtendBaseWidget(s)
	s.bg = s.themeBackground()
	items := []fyne.CanvasObject{s.bg}
	s.heading = nil
	s.subheading = nil
	s.addContent(&items, in.Segments)
	s.content = container.NewWithoutLayout(items...)
	return s
}

func (s *slide) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.content)
}

func (s *slide) Resize(size fyne.Size) {
	s.bg.Resize(size)

	s.layout(size)
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
				s.heading = canvas.NewText(seg.Text, color.Black)
				s.heading.TextStyle.Bold = true
				s.themeText(s.heading, seg.Style)

				s.variant = headingSlide
				*items = append(*items, s.heading)
			case widget.RichTextStyleSubHeading:
				if s.subheading != nil {
					s.subheading = nil
				}
				s.subheading = canvas.NewText(seg.Text, color.Black)
				s.subheading.TextStyle.Bold = true
				s.themeText(s.subheading, seg.Style)

				s.variant = headingSlide
				*items = append(*items, s.subheading)
			default:
				text := canvas.NewText(seg.Text, color.Black)
				s.themeText(text, seg.Style)

				s.variant = otherSlide
				*items = append(*items, text)
			}
		case *widget.ListSegment:
			s.addContent(items, seg.Items)
		case *widget.ParagraphSegment:
			s.addContent(items, seg.Texts)
		case *widget.ImageSegment:
			img := canvas.NewImageFromFile(seg.Source.Path())
			*items = append(*items, img)
			if s.heading == nil {
				s.variant = imageSlide
			} else {
				s.variant = otherSlide
			}
		}
	}
}

func (s *slide) setSource(rich *widget.RichText) {
	items := []fyne.CanvasObject{s.bg}
	s.heading = nil
	s.subheading = nil
	s.addContent(&items, rich.Segments)
	s.content.Objects = items
	s.content.Refresh()
	s.Resize(s.Size())
}
