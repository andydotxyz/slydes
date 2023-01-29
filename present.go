package main

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

func (g *gui) showPresentWindow() {
	w2 := fyne.CurrentApp().NewWindow("Play")

	content := widget.NewRichText()
	items := strings.Split(g.content.Text, "---")
	content.ParseMarkdown(items[0])

	colorTexts(content.Segments)
	content.Refresh()
	w2.SetContent(newAspectContainer(canvas.NewRectangle(color.White), content))

	addPresentationKeys(w2, content, items)
	w2.SetFullScreen(true)
	w2.Show()
}

func addPresentationKeys(w fyne.Window, content *widget.RichText, items []string) {
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
			content.ParseMarkdown(items[id])
			colorTexts(content.Segments)
			content.Refresh()
		case fyne.KeyRight, fyne.KeyDown, fyne.KeySpace, fyne.KeyEnter, fyne.KeyReturn:
			if id >= len(items)-1 {
				return
			}

			id++
			content.ParseMarkdown(items[id])
			colorTexts(content.Segments)
			content.Refresh()
		}
	})
}
