package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.New()
	w := a.NewWindow("Slydes")
	w.Resize(fyne.NewSize(600, 330))

	s := newSlides()
	g := newGUI(s, w)
	w.SetContent(g.makeUI())
	w.Canvas().Focus(g.content)
	w.ShowAndRun()
}
