package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var currentPresenting *presenting

// progressHeight is the thickness, in points, of the presentation progress bar.
const progressHeight = float32(5)

type presenting struct {
	live, control        fyne.Window
	slide, preview, next *slide

	id    int
	items []string

	progressFill *canvas.Rectangle
	presentLay   *presentLayout
}

// fraction returns how far through the deck we are, from 0 (first slide) to 1 (last).
func (p *presenting) fraction() float32 {
	if len(p.items) <= 1 {
		return 1
	}
	return float32(p.id) / float32(len(p.items)-1)
}

// progressColor picks the bar colour from the current slide: header slides use the
// header background colour, all others use the standard background colour.
func (p *presenting) progressColor() color.Color {
	th := p.slide.parent.theme
	v := fyne.CurrentApp().Settings().ThemeVariant()
	if p.slide.variant == headingSlide {
		return th.Color(colorNameHeaderBackground, v)
	}
	return th.Color(theme.ColorNameBackground, v)
}

// updateProgress recolours the progress bar and animates its width to match the
// current slide.
func (p *presenting) updateProgress() {
	if p.progressFill == nil {
		return
	}

	p.progressFill.FillColor = p.progressColor()
	p.progressFill.Refresh()

	p.presentLay.fraction = p.fraction()
	target := fyne.NewSize(p.presentLay.slideSize.Width*p.presentLay.fraction, progressHeight)
	canvas.NewSizeAnimation(p.progressFill.Size(), target, canvas.DurationStandard,
		func(s fyne.Size) {
			p.progressFill.Resize(s)
		}).Start()
}

// presentLayout fills the window with the slide and pins a progress bar of width
// proportional to fraction along the bottom edge of the slide. The slide is
// letterboxed to slideRatio, so the bar tracks the slide rather than the window.
type presentLayout struct {
	fraction float32

	slidePos  fyne.Position
	slideSize fyne.Size
}

func (l *presentLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	objs[0].Resize(size)
	objs[0].Move(fyne.Position{})

	width, height := size.Width, size.Height
	if width > height*slideRatio {
		width = height * slideRatio
	} else {
		height = width / slideRatio
	}
	l.slidePos = fyne.NewPos((size.Width-width)/2, (size.Height-height)/2)
	l.slideSize = fyne.NewSize(width, height)

	fill := objs[1]
	fill.Resize(fyne.NewSize(width*l.fraction, progressHeight))
	fill.Move(fyne.NewPos(l.slidePos.X, l.slidePos.Y+height-progressHeight))
}

func (l *presentLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	return objs[0].MinSize()
}

func (g *gui) showPresentWindow() {
	w2 := fyne.CurrentApp().NewWindow("Play")

	items := g.s.items
	id, _ := g.s.current.Get()
	content := newSlide(items[id], g.s)
	w2.SetPadded(false)

	p := &presenting{live: w2, slide: content, id: id, items: items}
	p.progressFill = canvas.NewRectangle(p.progressColor())
	p.presentLay = &presentLayout{fraction: p.fraction()}
	w2.SetContent(container.New(p.presentLay, newAspectContainer(content), p.progressFill))
	addPresentationKeys(w2)

	a := fyne.CurrentApp()
	hasMonitor := false
	if deskDrive, ok := a.Driver().(desktop.Driver); ok {
		hasMonitor = deskDrive.HasSecondaryDisplay()
	}

	if deskWin, ok := w2.(desktop.Window); ok && hasMonitor {
		deskWin.RequestFullScreenSecondary()

		pres := newPresenterGUI()
		w3 := pres.makeWindow(a)
		p.control = w3

		preview := newSlide(items[id], g.s)
		p.preview = preview
		nextString := ""
		if len(items) > id+1 {
			nextString = items[id+1]
		}
		next := newSlide(nextString, g.s)
		p.next = next

		pres.controls.Items[0].(*widget.ToolbarAction).OnActivated = prevSlide
		pres.controls.Items[1].(*widget.ToolbarAction).OnActivated = nextSlide
		pres.controls.Items[3].(*widget.ToolbarAction).OnActivated = exitPresent
		pres.controls.Refresh()

		pres.currentPreview.Objects = []fyne.CanvasObject{newAspectContainer(preview)}
		pres.nextPreview.Objects = []fyne.CanvasObject{newAspectContainer(next)}

		addPresentationKeys(w3)
		w3.Show()
		w3.SetFullScreen(true)
	} else {
		w2.SetFullScreen(true)
	}

	currentPresenting = p
	w2.Show()
}

func addPresentationKeys(w fyne.Window) {
	w.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		switch k.Name {
		case fyne.KeyEscape:
			exitPresent()
		case fyne.KeyLeft, fyne.KeyUp:
			prevSlide()
		case fyne.KeyRight, fyne.KeyDown, fyne.KeySpace, fyne.KeyEnter, fyne.KeyReturn:
			nextSlide()
		}
	})
}
