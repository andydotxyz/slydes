package main

import (
	"io/ioutil"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

func (g *gui) clearFile() {
	dialog.ShowConfirm("Clear content", "Are you sure you want to reset your slide content", func(ok bool) {
		if ok {
			g.s.uri = nil
			g.content.SetText("# Slide 1")
		}
	}, g.win)
}

func (g *gui) openFile() {
	dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, g.win)
			return
		}
		if r == nil {
			return
		}

		data, err := ioutil.ReadAll(r)
		_ = r.Close()

		if err != nil {
			dialog.ShowError(err, g.win)
		} else {
			g.s.uri = r.URI()
			g.content.SetText(string(data))
		}
	}, g.win)
}

func (g *gui) saveFile() {
	if g.s.uri != nil {
		w, err := storage.Writer(g.s.uri)
		if err != nil {
			dialog.ShowError(err, g.win)
			return
		}

		_, err = w.Write([]byte(g.content.Text))
		if err != nil {
			dialog.ShowError(err, g.win)
		}
		return
	}

	dialog.ShowFileSave(func(w fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, g.win)
			return
		}
		if w == nil {
			return
		}

		_, err = w.Write([]byte(g.content.Text))
		if err != nil {
			dialog.ShowError(err, g.win)
		}
		g.s.uri = w.URI()
	}, g.win)
}
