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

	content string
	scale   float32

	dot  *canvas.Circle
	text *canvas.Text
}

func newBullet(txt string) *bullet {
	return &bullet{content: txt, scale: 1}
}

func (b *bullet) CreateRenderer() fyne.WidgetRenderer {
	b.dot = canvas.NewCircle(color.Black)
	b.text = canvas.NewText(b.content, color.Black)
	return widget.NewSimpleRenderer(container.NewWithoutLayout(b.dot, b.text))
}

func (b *bullet) Resize(size fyne.Size) {
	b.dot.Move(fyne.NewPos(0, (size.Height - b.dot.Size().Height)/2))
	b.text.Move(fyne.NewPos(b.dot.Size().Width + theme.Padding()*b.scale, 0))
	b.text.Resize(fyne.NewSize(size.Width - b.dot.Size().Width - theme.Padding()*b.scale, size.Height))
}

func (b *bullet) MinSize() fyne.Size {
	if b.text == nil {
		return fyne.NewSize(14, 4)
	}

	return b.dot.Size().Add(b.text.MinSize()).AddWidthHeight(theme.Padding()*b.scale, 0)
}

func (b *bullet) setScale(scale float32) {
	_ = test.WidgetRenderer(b)
	b.scale = scale

	b.dot.Resize(fyne.NewSize(5*scale, 5*scale))
	b.text.TextSize = theme.TextSize() * scale
}