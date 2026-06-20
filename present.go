package main

import (
	_ "embed"
	"image"
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var currentPresenting *presenting

//go:embed "swap.svg"
var resourceSwapSvg []byte

// progressHeight is the thickness, in points, of the presentation progress bar.
const progressHeight = float32(5)

type presenting struct {
	live, control        fyne.Window
	slide, preview, next *slide
	deck                 *slides
	body                 *fyne.Container // the live window's aspect container
	flipped              bool

	id    int
	items []string

	progressBox, progressFill *canvas.Rectangle
	progressFraction          float32
	notesLabel                *widget.Label

	captures        []image.Image // one rendered bitmap per slide, for transition textures
	captureSize     fyne.Size     // resolution captures are rendered at
	capturePixScale float32       // pixel scale (size × pixScale = framebuffer pixels)
	captureMu       sync.Mutex    // serialises capture goroutines
	animating       bool          // true while a transition shader is on screen
}

// updateNotes copies the current preview slide's notes into the presenter UI.
func (p *presenting) updateNotes() {
	if p.notesLabel == nil || p.preview == nil {
		return
	}
	p.notesLabel.SetText(p.preview.notes)
}

// fraction returns how far through the deck we are, from 0 (first slide) to 1 (last).
func (p *presenting) fraction() float32 {
	if len(p.items) <= 1 {
		return 1
	}
	return float32(p.id) / float32(len(p.items)-1)
}

// updateProgress recolours the progress bar and animates its width to match the
// current slide.
func (p *presenting) updateProgress() {
	if p.progressFill == nil {
		return
	}

	p.progressFill.FillColor = p.deck.theme.Color(colorNameHeaderBackground,
		fyne.CurrentApp().Settings().ThemeVariant())
	p.progressFill.Refresh()

	p.progressFraction = p.fraction()
	target := fyne.NewSize(p.body.Size().Width*p.progressFraction, progressHeight)
	canvas.NewSizeAnimation(p.progressFill.Size(), target, transitionDuration,
		func(s fyne.Size) {
			p.progressFill.Resize(s)
		}).Start()
}

func (g *gui) showPresentWindow() {
	w2 := fyne.CurrentApp().NewWindow("Play")

	items := g.s.items
	id, _ := g.s.current.Get()
	content := newSlide(items[id], id, g.s)
	w2.SetPadded(false)

	body := newAspectContainer(content)
	p := &presenting{
		live: w2, slide: content, deck: g.s, body: body, id: id, items: items,
		captures: make([]image.Image, len(items)),
	}
	p.progressBox = canvas.NewRectangle(color.Black)
	p.progressBox.SetMinSize(fyne.NewSquareSize(progressHeight))
	p.progressFill = canvas.NewRectangle(p.slide.footerColor())
	p.progressFill.Resize(fyne.NewSize(0, progressHeight))
	w2.SetContent(
		container.NewStack(canvas.NewRectangle(color.Black),
			body,
			container.NewBorder(nil, container.NewStack(p.progressBox,
				container.NewWithoutLayout(p.progressFill)), nil, nil)),
	)

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
		w3.SetPadded(false)
		p.control = w3

		preview := newSlide(items[id], id, g.s)
		p.preview = preview
		nextString := ""
		if len(items) > id+1 {
			nextString = items[id+1]
		}
		next := newSlide(nextString, id+1, g.s)
		p.next = next

		pres.controls.Items[0].(*widget.ToolbarAction).OnActivated = prevSlide
		pres.controls.Items[1].(*widget.ToolbarAction).OnActivated = nextSlide
		pres.controls.Items[3].(*widget.ToolbarAction).Icon = theme.NewThemedResource(
			fyne.NewStaticResource("swap.svg", resourceSwapSvg),
		)
		pres.controls.Items[3].(*widget.ToolbarAction).OnActivated = togglePresent
		pres.controls.Items[4].(*widget.ToolbarAction).OnActivated = exitPresent
		pres.controls.Refresh()

		pres.currentPreview.Objects = []fyne.CanvasObject{newAspectContainer(preview)}
		pres.nextPreview.Objects = []fyne.CanvasObject{newAspectContainer(next)}
		p.notesLabel = pres.notes
		pres.notes.SizeName = theme.SizeNameSubHeadingText
		p.updateNotes()

		addPresentationKeys(w3)
		w3.Show()
		w3.SetFullScreen(true)
	} else {
		w2.SetFullScreen(true)
	}

	currentPresenting = p
	w2.Show()

	// Render and cache a bitmap of every slide so transitions can pass the
	// outgoing and incoming slides to the shuffle shader as textures.
	go precaptureSlides(p)
}

func addPresentationKeys(w fyne.Window) {
	w.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		switch k.Name {
		case fyne.KeyEscape:
			exitPresent()
		case fyne.KeyT:
			togglePresent()
		case fyne.KeyLeft, fyne.KeyUp:
			prevSlide()
		case fyne.KeyRight, fyne.KeyDown, fyne.KeySpace, fyne.KeyEnter, fyne.KeyReturn:
			nextSlide()
		}
	})
}
