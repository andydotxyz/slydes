package main

import (
	"strings"

	"fyne.io/fyne/v2"
)

func (g *gui) showPresentWindow() {
	w2 := fyne.CurrentApp().NewWindow("Play")

	items := strings.Split(g.content.Text, "---")
	id, _ := g.s.current.Get()
	content := newSlide(items[id], g.s)
	w2.SetContent(newAspectContainer(content))

	addPresentationKeys(w2, content, items, id)
	w2.SetFullScreen(true)
	w2.Show()
}

func addPresentationKeys(w fyne.Window, content *slide, items []string, id int) {
	w.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		switch k.Name {
		case fyne.KeyEscape:
			w.Close()
		case fyne.KeyLeft, fyne.KeyUp:
			if id <= 0 {
				return
			}

			id--
			content.setSource(items[id])
		case fyne.KeyRight, fyne.KeyDown, fyne.KeySpace, fyne.KeyEnter, fyne.KeyReturn:
			if id >= len(items)-1 {
				return
			}

			id++
			content.setSource(items[id])
		}
	})
}
