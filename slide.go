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
	parent  *slides

	content             *fyne.Container
	bg                  fyne.CanvasObject
	heading, subheading *canvas.Text
}

func newSlide(data string, parent *slides) *slide {
	s := &slide{parent: parent}
	s.ExtendBaseWidget(s)
	s.bg = s.themeBackground()
	items := []fyne.CanvasObject{s.bg}
	s.heading = nil
	s.subheading = nil
	s.addContent(&items, parent.parseMarkdown(data))
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

func (s *slide) addContent(items *[]fyne.CanvasObject, in content) {
	if in.bgpath != "" {
		img := canvas.NewImageFromFile(in.bgpath)
		img.ScaleMode = canvas.ImageScaleFastest
		*items = append(*items, img)
		s.variant = imageSlide
		return
	}

	if in.heading != "" {
		s.heading = canvas.NewText(in.heading, color.Black)
		s.heading.TextStyle.Bold = true
		s.themeText(s.heading, widget.RichTextStyleHeading)

		s.variant = headingSlide
		*items = append(*items, s.heading)
	}
	if in.subheading != "" {
		s.subheading = canvas.NewText(in.subheading, color.Black)
		s.subheading.TextStyle.Bold = true
		s.themeText(s.subheading, widget.RichTextStyleSubHeading)

		s.variant = headingSlide
		*items = append(*items, s.subheading)
	}

	if len(in.content) > 0 {
		s.variant = otherSlide
		*items = append(*items, in.content...)
	}
}

func (s *slide) setSource(data string) {
	items := []fyne.CanvasObject{s.bg}
	s.heading = nil
	s.subheading = nil
	s.addContent(&items, s.parent.parseMarkdown(data))
	s.content.Objects = items
	s.content.Refresh()
	s.Resize(s.Size())
}
