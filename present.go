package main

import (
	"strings"

	"fyne.io/fyne/v2"
)

func (g *gui) showPresentWindow() {
	w2 := fyne.CurrentApp().NewWindow("Play")

	items := strings.Split(g.content.Text, "---")
	content := newSlide(items[0])
	w2.SetContent(newAspectContainer(content))

	addPresentationKeys(w2, content, items)
	w2.SetFullScreen(true)
	w2.Show()
}

func addPresentationKeys(w fyne.Window, content *slide, items []string) {
	id := 0
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
