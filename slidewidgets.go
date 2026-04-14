package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type slideWidget interface {
	setScale(float32)
}

type bullet struct {
	widget.BaseWidget
	theme fyne.Theme

	content string
	indent  int
	scale   float32

	dot  *canvas.Circle
	text *canvas.Text
}

func newBullet(txt string, indent int, th fyne.Theme) *bullet {
	return &bullet{theme: th, content: txt, indent: indent, scale: 1}
}

func (b *bullet) CreateRenderer() fyne.WidgetRenderer {
	b.dot = canvas.NewCircle(b.theme.Color(colorNameBullet, theme.VariantLight))
	if b.content == "" {
		b.dot.FillColor = color.Transparent
	}

	b.text = canvas.NewText(b.content, b.theme.Color(colorNameBullet, theme.VariantLight))
	return widget.NewSimpleRenderer(container.NewWithoutLayout(b.dot, b.text))
}

func (b *bullet) Refresh() {
	if b.dot != nil {
		if b.content == "" {
			b.dot.FillColor = color.Transparent
		} else {
			b.dot.FillColor = b.theme.Color(colorNameBullet, theme.VariantLight)
		}
		b.dot.Refresh()
	}
	if b.text != nil {
		b.text.Color = b.theme.Color(colorNameBullet, theme.VariantLight)
		b.text.Refresh()
	}
}

func (b *bullet) indentOffset() float32 {
	return float32(b.indent) * theme.Padding() * 4 * b.scale
}

func (b *bullet) Resize(size fyne.Size) {
	off := b.indentOffset()
	b.dot.Move(fyne.NewPos(off, (size.Height-b.dot.Size().Height)/2))
	b.text.Move(fyne.NewPos(off+b.dot.Size().Width+theme.Padding()*b.scale, 0))
	b.text.Resize(fyne.NewSize(size.Width-off-b.dot.Size().Width-theme.Padding()*b.scale, size.Height))
}

func (b *bullet) MinSize() fyne.Size {
	if b.text == nil {
		return fyne.NewSize(14, 4)
	}

	return b.dot.Size().Add(b.text.MinSize()).AddWidthHeight(theme.Padding()*b.scale+b.indentOffset(), 0)
}

func (b *bullet) setScale(scale float32) {
	_ = test.WidgetRenderer(b)
	b.scale = scale

	b.dot.Resize(fyne.NewSize(5*scale, 5*scale))
	b.text.TextSize = theme.TextSize() * scale
}
