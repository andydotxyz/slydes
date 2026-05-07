package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

var currentPresenting *presenting

type presenting struct {
	live, control        fyne.Window
	slide, preview, next *slide

	id    int
	items []string
}

func (g *gui) showPresentWindow() {
	w2 := fyne.CurrentApp().NewWindow("Play")

	items := g.s.items
	id, _ := g.s.current.Get()
	content := newSlide(items[id], g.s)
	w2.SetPadded(false)
	w2.SetContent(newAspectContainer(content))

	p := &presenting{live: w2, slide: content, id: id, items: items}
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
