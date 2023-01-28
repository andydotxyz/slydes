package main

import (
	"image/color"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type gui struct {
	count, current binding.Int

	content *widget.Entry
	preview *widget.RichText
}

func newGUI() *gui {
	g := &gui{count: binding.NewInt(), current: binding.NewInt()}
	_ = g.count.Set(1)
	return g
}

func (g *gui) makeUI() fyne.CanvasObject {
	g.content = widget.NewMultiLineEntry()
	g.preview = widget.NewRichText()
	render := container.NewMax(canvas.NewRectangle(color.White), g.preview)

	previews := container.NewGridWithRows(1)
	refreshPreviews := func() {
		count, _ := g.count.Get()
		items := make([]fyne.CanvasObject, count)
		for i := 0; i < count; i++ {
			slide := g.newSlideButton(i)
			items[i] = container.NewPadded(slide)
		}
		previews.Objects = items
		previews.Refresh()
	}
	refreshPreviews()
	g.count.AddListener(binding.NewDataListener(refreshPreviews))
	g.current.AddListener(binding.NewDataListener(func() {
		refreshPreviews()
		g.refreshSlide()
	}))

	g.content.OnChanged = func(s string) {
		g.refreshSlide()
	}
	g.content.SetText("# Slide 1")

	split := container.NewHSplit(g.content, newAspectContainer(render))
	split.Offset = 0.35
	return container.NewBorder(
		container.NewVBox(
			widget.NewToolbar(
				widget.NewToolbarAction(theme.FileIcon(), func() {}),
				widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {}),
				widget.NewToolbarSeparator(),
				widget.NewToolbarAction(theme.NavigateBackIcon(), func() {
					i, _ := g.current.Get()
					if i > 0 {
						_ = g.current.Set(i - 1)
					}
				}),
				widget.NewToolbarAction(theme.NavigateNextIcon(), func() {
					i, _ := g.current.Get()
					c, _ := g.count.Get()
					if i < c-1 {
						_ = g.current.Set(i + 1)
					}
				}),
				widget.NewToolbarAction(theme.MediaPlayIcon(), func() {
					w2 := fyne.CurrentApp().NewWindow("Play")

					content := widget.NewRichText()
					items := strings.Split(g.content.Text, "---")
					content.ParseMarkdown(items[0])

					colorTexts(content.Segments)
					content.Refresh()
					w2.SetContent(newAspectContainer(canvas.NewRectangle(color.White), content))

					w2.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
						if k.Name == fyne.KeyEscape {
							w2.Close()
						}
					})
					w2.SetFullScreen(true)
					w2.Show()
				}),
				widget.NewToolbarSpacer(),
				widget.NewToolbarAction(theme.HelpIcon(), func() {}),
			),
			container.NewHScroll(container.NewMax(
				canvas.NewRectangle(theme.MenuBackgroundColor()),
				container.NewHBox(previews)))),
		nil,
		nil,
		nil,
		split)
}

func (g *gui) refreshSlide() {
	items := strings.Split(g.content.Text, "---")
	_ = g.count.Set(len(items))
	id, _ := g.current.Get()
	if id >= len(items) {
		log.Println("Cannot set slide beyond length")
		id = len(items) - 1
	}
	g.preview.ParseMarkdown(items[id])

	colorTexts(g.preview.Segments)
	g.preview.Refresh()
}
