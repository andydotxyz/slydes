package main

import (
	"flag"
	"io/ioutil"
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
	w.SetContent(g.makeUI())
	w.Canvas().Focus(g.content)

	flag.Parse()
	if len(flag.Args()) > 0 && len(flag.Args()[0]) > 0 {
		path := flag.Args()[0]

		f, _ := os.Open(path)
		data, err := ioutil.ReadAll(f)
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
