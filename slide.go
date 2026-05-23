package main

import (
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
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
	index   int

	content             *fyne.Container
	footer              *fyne.Container
	bg                  fyne.CanvasObject
	heading, subheading *canvas.Text

	footerLeft, footerCenter, footerRight *canvas.Text
}

func newSlide(data string, index int, parent *slides) *slide {
	s := &slide{parent: parent, index: index}
	s.ExtendBaseWidget(s)

	s.bg = s.themeBackground()
	items := []fyne.CanvasObject{s.bg}
	s.heading = nil
	s.subheading = nil
	if data == "" {
		s.bg = canvas.NewRectangle(color.Black)
		s.content = container.NewWithoutLayout()
	} else {
		s.addContent(&items, parent.parseMarkdown(data))
		s.content = container.NewWithoutLayout(items...)
	}

	s.makeFooter()
	if data == "" || s.variant == imageSlide {
		s.footer.Hide()
	}
	return s
}

// makeFooter builds the three footer labels: presenter name (left), configurable
// text (centre) and slide number (right).
func (s *slide) makeFooter() {
	s.footerLeft = canvas.NewText("", color.Black)
	s.footerLeft.Alignment = fyne.TextAlignLeading
	s.footerCenter = canvas.NewText("", color.Black)
	s.footerCenter.Alignment = fyne.TextAlignCenter
	s.footerRight = canvas.NewText("", color.Black)
	s.footerRight.Alignment = fyne.TextAlignTrailing

	s.footer = container.NewWithoutLayout(s.footerLeft, s.footerCenter, s.footerRight)
	s.updateFooterText()
}

// updateFooterText refreshes the footer labels from the current config and index.
func (s *slide) updateFooterText() {
	s.footerLeft.Text = presenterName(s.parent.config)
	s.footerCenter.Text = s.parent.config.Footer
	s.footerRight.Text = strconv.Itoa(s.index + 1)
}

// hideFooter permanently hides the footer, for contexts such as slide thumbnails.
func (s *slide) hideFooter() {
	s.footer.Hide()
}

func (s *slide) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewWithoutLayout(s.content, s.footer))
}

func (s *slide) Resize(size fyne.Size) {
	s.bg.Resize(size)
	s.content.Resize(size)
	s.footer.Resize(size)

	s.layout(size)
	s.layoutFooter(size)
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
		s.heading = canvas.NewText(in.heading, s.parent.theme.Color(colorNameHeader, theme.VariantLight))
		s.heading.TextStyle.Bold = true
		s.variant = headingSlide
		*items = append(*items, s.heading)
	}
	if in.subheading != "" {
		s.subheading = canvas.NewText(in.subheading, s.parent.theme.Color(colorNameSubHeader, theme.VariantLight))
		s.subheading.TextStyle.Bold = true

		s.variant = headingSlide
		*items = append(*items, s.subheading)
	}

	if len(in.content) > 0 {
		s.variant = otherSlide
		*items = append(*items, in.content...)
	}

	for _, o := range *items {
		if t, ok := o.(*canvas.Text); ok {
			if t == s.heading || t == s.subheading {
				continue
			}
			t.Color = s.parent.theme.Color(theme.ColorNameForeground, theme.VariantLight)
		}
	}
}

func (s *slide) setSource(data string, index int) {
	s.index = index
	s.updateFooterText()

	if data == "" {
		s.bg = canvas.NewRectangle(color.Black)
		s.content.Objects = []fyne.CanvasObject{}
		s.footer.Hide()
		s.Refresh()
		return
	}

	s.bg = s.themeBackground()
	items := []fyne.CanvasObject{s.bg}
	s.heading = nil
	s.subheading = nil
	s.addContent(&items, s.parent.parseMarkdown(data))
	s.content.Objects = items
	if s.variant == imageSlide {
		s.footer.Hide()
	} else {
		s.footer.Show()
	}
	s.content.Refresh()
	s.Resize(s.Size())
}

// footerColor matches the progress bar: header slides use the header background
// colour, all others use the standard background colour.
func (s *slide) footerColor() color.Color {
	v := fyne.CurrentApp().Settings().ThemeVariant()
	if s.variant == headingSlide {
		return s.parent.theme.Color(colorNameHeaderBackground, v)
	}
	return s.parent.theme.Color(theme.ColorNameBackground, v)
}
