package main

import (
	"flag"
	"fmt"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

func main() {
	a := app.New()
	w := a.NewWindow("Slydes")
	w.Resize(fyne.NewSize(600, 330))

	s := newSlides()
	g := newGUI(s, w)
	w.SetMaster()
	w.SetContent(g.makeUI())
	w.Canvas().Focus(g.content)

	flag.Parse()
	if len(flag.Args()) > 0 && len(flag.Args()[0]) > 0 {
		path := flag.Args()[0]

		f, _ := os.Open(path)
		data, err := io.ReadAll(f)
		_ = f.Close()

		if err != nil {
			dialog.ShowError(err, g.win)
		} else {
			absPath, _ := filepath.Abs(path)
			g.s.uri = storage.NewFileURI(absPath)
			g.content.SetText(string(data))
		}
	}

	w.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Open", g.openFile)),
		slideshowMenu(g),
		transitionMenu(),
	))
	w.ShowAndRun()
}

// slideshowMenu builds the "Slideshow" menu, holding the options that apply the
// next time a presentation is played.
func slideshowMenu(g *gui) *fyne.Menu {
	menu := fyne.NewMenu("Slideshow",
		fyne.NewMenuItem("Play", g.showPresentWindow),
		fyne.NewMenuItemSeparator(),
	)
	loop := fyne.NewMenuItem("Loop", nil)
	loop.Checked = loopSlideshow
	loop.Action = func() {
		loopSlideshow = !loopSlideshow
		loop.Checked = loopSlideshow
		menu.Refresh()
	}
	menu.Items = append(menu.Items, loop, autoProgressItem(menu))
	return menu
}

func autoProgressItem(parent *fyne.Menu) *fyne.MenuItem {
	auto := fyne.NewMenuItem("Auto-progress", nil)
	auto.Checked = autoInterval != 0
	auto.ChildMenu = fyne.NewMenu("")

	intervals := []time.Duration{0, 5 * time.Second, 15 * time.Second, 30 * time.Second}
	for _, d := range intervals {
		title := "Off"
		if d != 0 {
			title = fmt.Sprintf("%d seconds", int(d.Seconds()))
		}

		item := fyne.NewMenuItem(title, nil)
		item.Checked = d == autoInterval
		item.Action = func() {
			autoInterval = d
			for i, other := range auto.ChildMenu.Items {
				other.Checked = intervals[i] == autoInterval
			}
			auto.Checked = autoInterval != 0
			auto.ChildMenu.Refresh()
			parent.Refresh()
		}
		auto.ChildMenu.Items = append(auto.ChildMenu.Items, item)
	}
	return auto
}

// transitionMenu builds the "Transitions" menu: one item per movement, ticked
// when it is the one in use.
func transitionMenu() *fyne.Menu {
	menu := fyne.NewMenu("Transitions")
	for _, t := range slideTransitions {
		item := fyne.NewMenuItem(t.title, nil)
		item.Checked = t == currentTransition
		item.Action = func() {
			currentTransition = t
			for i, other := range menu.Items {
				other.Checked = slideTransitions[i] == currentTransition
			}
			menu.Refresh()
		}
		menu.Items = append(menu.Items, item)
	}
	return menu
}

func nextSlide() {
	if currentPresenting == nil {
		return
	}

	p := currentPresenting
	if p.id >= len(p.items)-1 {
		if !p.loop || len(p.items) <= 1 {
			return
		}

		p.wrapping = true
		changeSlide(p, 0)
		p.wrapping = false
		return
	}

	changeSlide(p, p.id+1)
}

func prevSlide() {
	if currentPresenting == nil {
		return
	}

	p := currentPresenting
	if p.id <= 0 {
		return
	}

	changeSlide(p, p.id-1)
}

func exitPresent() {
	if currentPresenting == nil {
		return
	}

	close(currentPresenting.done)
	currentPresenting.live.Close()
	if currentPresenting.control != nil {
		currentPresenting.control.Close()
	}
	currentPresenting = nil
}

func togglePresent() {
	if currentPresenting == nil {
		return
	}

	preview := currentPresenting.control.Content()
	view := currentPresenting.live.Content()
	currentPresenting.control.SetContent(canvas.NewRectangle(color.Transparent))
	currentPresenting.live.SetContent(canvas.NewRectangle(color.Transparent))

	currentPresenting.flipped = !currentPresenting.flipped
	currentPresenting.control.SetContent(view)
	currentPresenting.live.SetContent(preview)

	currentPresenting.updateProgress()
	// in case of aspect ratio change
	go precaptureSlides(currentPresenting)
	view.Refresh()
	preview.Refresh()
}
