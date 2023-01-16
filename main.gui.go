package main

import (
	"fmt"
	"image/color"
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
}

func newGUI() *gui {
	g := &gui{count: binding.NewInt(), current: binding.NewInt()}
	_ = g.count.Set(1)
	return g
}

func (g *gui) makeUI() fyne.CanvasObject {
	entry := widget.NewMultiLineEntry()
	content := widget.NewRichText()
	render := container.NewMax(canvas.NewRectangle(color.White), content)

	previews := container.NewGridWithRows(1)
	refreshPreview := func() {
		count, _ := g.count.Get()
		items := make([]fyne.CanvasObject, count)
		for i := 0; i < count; i++ {
			bg := canvas.NewRectangle(color.White)
			bg.StrokeColor = theme.PrimaryColor()
			c, _ := g.current.Get()
			if c == i {
				bg.StrokeWidth = 3
			} else {
				bg.StrokeWidth = 0
			}

			t := fmt.Sprintf("Slide %d", i+1)
			title := canvas.NewText(t, theme.BackgroundColor())
			title.TextSize = 8
			slide := newAspectContainer(bg, container.NewPadded(container.NewVBox(title)))
			num := fmt.Sprintf("%d:", i+1)
			items[i] = container.NewPadded(
				container.NewHBox(container.NewVBox(canvas.NewText(num, theme.ForegroundColor())), slide))
		}
		previews.Objects = items
		previews.Refresh()
	}
	refreshPreview()
	g.count.AddListener(binding.NewDataListener(refreshPreview))
	g.current.AddListener(binding.NewDataListener(refreshPreview))

	entry.OnChanged = func(s string) {
		items := strings.Split(s, "---")
		_ = g.count.Set(len(items))
		content.ParseMarkdown(items[0])

		colorTexts(content.Segments)
		content.Refresh()
	}

	entry.SetText("# Slide 1")

	split := container.NewHSplit(entry, newAspectContainer(render))
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
					items := strings.Split(entry.Text, "---")
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
