package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// slideRatio is the aspect ratio (width / height) every slide is rendered at.
const slideRatio = float32(16) / float32(9)

type aspectLayout struct {
	ratio float32
}

func (a *aspectLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	width, height := size.Width, size.Height
	if width > height*a.ratio {
		width = height * a.ratio
	} else {
		height = width / a.ratio
	}

	inner := fyne.NewSize(width, height)
	pos := fyne.NewPos((size.Width-width)/2, (size.Height-height)/2)
	for _, o := range objs {
		o.Resize(inner)
		o.Move(pos)
	}
}

func (a *aspectLayout) MinSize([]fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(80, 45)
}

func newAspectContainer(children ...fyne.CanvasObject) *fyne.Container {
	return container.New(&aspectLayout{ratio: slideRatio}, children...)
}
