package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.New()
	w := a.NewWindow("Slydes")
	w.Resize(fyne.NewSize(600, 320))

	g := newGUI()
	w.SetContent(g.makeUI())
	w.ShowAndRun()
}
