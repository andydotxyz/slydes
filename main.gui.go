package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type gui struct {
	content *widget.Entry
	render  *slide

	win fyne.Window
	s   *slides
	uri fyne.URI
}

func newGUI(s *slides, w fyne.Window) *gui {
	return &gui{s: s, win: w}
}

func (g *gui) makeUI() fyne.CanvasObject {
	g.content = widget.NewMultiLineEntry()

	previews := container.NewGridWithRows(1)
	refreshPreviews := func() {
		count, _ := g.s.count.Get()
		items := make([]fyne.CanvasObject, count)
		for i := 0; i < count; i++ {
			slide := g.newSlideButton(i)
			items[i] = container.NewPadded(slide)
		}
		previews.Objects = items
		previews.Refresh()
	}
	refreshPreviews()
	g.s.count.AddListener(binding.NewDataListener(refreshPreviews))
	g.s.current.AddListener(binding.NewDataListener(func() {
		refreshPreviews()
		g.refreshSlide()
	}))

	g.render = newSlide("")
	g.content.OnChanged = func(s string) {
		g.s.parseSource(s)
		g.slideForCursor()
		refreshPreviews()
		g.refreshSlide()
	}
	g.content.OnCursorChanged = g.slideForCursor
	g.content.SetText("# Slide 1")

	split := container.NewHSplit(g.content, newAspectContainer(g.render))
	split.Offset = 0.35
	return container.NewBorder(
		container.NewVBox(
			widget.NewToolbar(
				widget.NewToolbarAction(theme.FileIcon(), g.clearFile),
				widget.NewToolbarAction(theme.FolderOpenIcon(), g.openFile),
				widget.NewToolbarAction(theme.DocumentSaveIcon(), g.saveFile),
				widget.NewToolbarSeparator(),
				widget.NewToolbarAction(theme.NavigateBackIcon(), func() {
					i, _ := g.s.current.Get()
					if i > 0 {
						g.moveToSlide(i - 1)
					}
				}),
				widget.NewToolbarAction(theme.NavigateNextIcon(), func() {
					i, _ := g.s.current.Get()
					c, _ := g.s.count.Get()
					if i < c-1 {
						g.moveToSlide(i + 1)
					}
				}),
				widget.NewToolbarAction(theme.MediaPlayIcon(), g.showPresentWindow),
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

func (g *gui) moveToSlide(id int) {
	g.content.CursorColumn = 0
	if len(g.s.divideRows) == 0 || id == 0 {
		g.content.CursorRow = 0
	} else {
		div := g.s.divideRows[id-1]
		g.content.CursorRow = div + 1
	}

	g.win.Canvas().Focus(g.content)
}

func (g *gui) slideForCursor() {
	id := 0
	for _, row := range g.s.divideRows {
		if g.content.CursorRow < row {
			break
		} else if g.content.CursorRow == row && g.content.CursorColumn < 3 {
			break // if it's a divide line, but not on the end
		}
		id++
	}
	g.s.current.Set(id)
}

func (g *gui) refreshSlide() {
	g.render.setSource(g.s.currentSource())
}
