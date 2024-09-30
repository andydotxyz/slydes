package main

import (
	"image/color"
	"strings"
	"time"

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

	refresh bool
}

func newGUI(s *slides, w fyne.Window) *gui {
	return &gui{s: s, win: w}
}

func (g *gui) makeUI() fyne.CanvasObject {
	g.content = widget.NewMultiLineEntry()
	border := canvas.NewRectangle(color.Transparent)
	border.StrokeColor = theme.PrimaryColor()
	border.StrokeWidth = 2
	border.CornerRadius = theme.InputRadiusSize()

	grid := container.NewGridWithRows(1)
	cellSize := fyne.NewSize(0, 0)
	refreshPreviews := func() {
		count, _ := g.s.count.Get()
		items := make([]fyne.CanvasObject, count+1)
		for i := 0; i < count; i++ {
			slide := g.newSlideButton(i)
			items[i] = container.NewPadded(slide)
		}
		items[count] = widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
			rows := len(strings.Split(g.content.Text, "\n")) - 1

			g.content.CursorRow = rows + 2
			g.content.Append(`
---
# New Slide
`)
		})
		grid.Objects = items
		cellSize = grid.Objects[0].MinSize()
		height := cellSize.Height - 4
		border.Resize(fyne.NewSize(height*16.0/9.0-3, height))
		grid.Refresh()
	}

	var previewScroll *container.Scroll
	previews := container.NewStack(grid, container.NewWithoutLayout(border))
	go refreshPreviews()
	moveHighlight := func(anim bool) {
		i, _ := g.s.current.Get()
		dest := fyne.NewPos(cellSize.Width*float32(i)+(theme.Padding()*float32(i-1))+6, 2)

		if previewScroll != nil {
			if dest.X < previewScroll.Offset.X {
				previewScroll.Offset.X = dest.X
				previewScroll.Refresh()
			} else if dest.X+border.Size().Width > previewScroll.Offset.X+previewScroll.Size().Width {
				previewScroll.Offset.X = dest.X + border.Size().Width - previewScroll.Size().Width
				previewScroll.Refresh()
			}
		}

		if !anim {
			border.Move(dest)
			return
		}

		canvas.NewPositionAnimation(border.Position(), dest, canvas.DurationShort, func(p fyne.Position) {
			border.Move(p)
		}).Start()
	}
	moveHighlight(false)
	g.s.current.AddListener(binding.NewDataListener(func() {
		moveHighlight(true)
		g.refreshSlide()
	}))

	g.render = newSlide("", g.s)
	g.content.OnChanged = func(s string) {
		g.refresh = true
	}
	g.content.OnCursorChanged = g.slideForCursor
	g.content.SetText("# Slide 1\n")

	split := container.NewHSplit(g.content, newAspectContainer(g.render))
	split.Offset = 0.35
	play := &primaryAction{widget.NewToolbarAction(theme.MediaPlayIcon(), g.showPresentWindow)}

	go func() {
		for {
			time.Sleep(time.Second / 10)
			if !g.refresh {
				continue
			}
			g.refresh = false

			g.s.parseSource(g.content.Text)
			go refreshPreviews()
			g.slideForCursor()
			moveHighlight(true)
			g.refreshSlide()
		}
	}()

	previewScroll = container.NewHScroll(container.NewStack(
		canvas.NewRectangle(theme.MenuBackgroundColor()),
		container.NewHBox(previews)))

	return container.NewBorder(
		container.NewVBox(
			widget.NewToolbar(
				widget.NewToolbarAction(theme.FileIcon(), g.clearFile),
				widget.NewToolbarAction(theme.FolderOpenIcon(), g.openFile),
				widget.NewToolbarAction(theme.DocumentSaveIcon(), g.saveFile),
				widget.NewToolbarAction(theme.DocumentPrintIcon(), g.exportFile),
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
				play,
				widget.NewToolbarSpacer(),
				widget.NewToolbarAction(theme.HelpIcon(), func() {}),
			),
			previewScroll),
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
	g.content.Refresh()

	g.win.Canvas().Focus(g.content)
	_ = g.s.current.Set(id)
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

type primaryAction struct {
	*widget.ToolbarAction
}

func (t *primaryAction) ToolbarObject() fyne.CanvasObject {
	button := t.ToolbarAction.ToolbarObject().(*widget.Button)
	button.Importance = widget.HighImportance

	return button
}
