package main

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"sync"
	"time"

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
	g                    *gui

	id    int
	items []string

	progressBox, progressFill *canvas.Rectangle
	progressFraction          float32
	notesLabel                *widget.Label

	clockText, timerText *widget.RichText
	started              time.Time     // when the first slide transition happened
	done                 chan struct{} // closed when the presentation ends

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

func (p *presenting) startTimer() {
	if p.started.IsZero() {
		p.started = time.Now()
	}
}

// runClocks updates the presenter display's wall clock and elapsed timer once a
// second, until the presentation ends.
func (p *presenting) runClocks() {
	p.updateClocks()

	go func() {
		tick := time.NewTicker(time.Second)
		defer tick.Stop()

		for {
			select {
			case <-p.done:
				return
			case <-tick.C:
				fyne.Do(p.updateClocks)
			}
		}
	}()
}

func (p *presenting) updateClocks() {
	p.clockText.ParseMarkdown(time.Now().Format("# 3:04pm"))

	var elapsed time.Duration
	if !p.started.IsZero() {
		elapsed = time.Since(p.started)
	}
	p.timerText.ParseMarkdown(fmt.Sprintf("# %02d:%02d",
		int(elapsed.Minutes()), int(elapsed.Seconds())%60))
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
		captures: make([]image.Image, len(items)), g: g, done: make(chan struct{}),
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

		p.clockText = pres.clock
		p.timerText = pres.timer
		p.runClocks()

		addPresentationKeys(w3)
		w3.Show()
		w3.SetFullScreen(true)
	} else {
		w2.SetFullScreen(true)
	}

	currentPresenting = p
	w2.Show()
	changeSlide(p, id)

	go precaptureSlides(p)

	// TODO remove this workaround and move it to layout?
	// Caused by the window size not being set when initial slide is loaded.
	go func() {
		time.Sleep(time.Millisecond * 100)
		p.updateProgress()
	}()
}

func addPresentationKeys(w fyne.Window) {
	w.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		switch k.Name {
		case fyne.KeyEscape:
			exitPresent()
		case fyne.KeyT:
			togglePresent()
		case fyne.KeyLeft, fyne.KeyUp, fyne.KeyPageUp:
			prevSlide()
		case fyne.KeyRight, fyne.KeyDown, fyne.KeyPageDown, fyne.KeySpace, fyne.KeyEnter, fyne.KeyReturn:
			nextSlide()
		}
	})
}
