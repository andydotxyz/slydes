package main

import (
	"flag"
	"io"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

func main() {
	a := app.New()
	w := a.NewWindow("Slydes")
	w.Resize(fyne.NewSize(600, 330))

	s := newSlides()
	g := newGUI(s, w)
	w.SetMaster()
	w.SetContent(g.makeUI())
	w.Canvas().Focus(g.content)

	flag.Parse()
	if len(flag.Args()) > 0 && len(flag.Args()[0]) > 0 {
		path := flag.Args()[0]

		f, _ := os.Open(path)
		data, err := io.ReadAll(f)
		_ = f.Close()

		if err != nil {
			dialog.ShowError(err, g.win)
		} else {
			absPath, _ := filepath.Abs(path)
			g.s.uri = storage.NewFileURI(absPath)
			g.content.SetText(string(data))
		}
	}

	w.ShowAndRun()
}

func nextSlide() {
	if currentPresenting == nil {
		return
	}

	p := currentPresenting
	if p.id >= len(p.items)-1 {
		return
	}

	changeSlide(p, p.id+1)
}

func prevSlide() {
	if currentPresenting == nil {
		return
	}

	p := currentPresenting
	if p.id <= 0 {
		return
	}

	changeSlide(p, p.id-1)
}

func exitPresent() {
	if currentPresenting == nil {
		return
	}

	currentPresenting.live.Close()
	if currentPresenting.control != nil {
		currentPresenting.control.Close()
	}
	currentPresenting = nil
}

func togglePresent() {
	if currentPresenting == nil {
		return
	}

	preview := currentPresenting.control.Content()
	view := currentPresenting.live.Content()

	currentPresenting.flipped = !currentPresenting.flipped
	currentPresenting.control.SetContent(view)
	currentPresenting.live.SetContent(preview)

	currentPresenting.updateProgress()
	// in case of aspect ratio change
	go precaptureSlides(currentPresenting)
}
