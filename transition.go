package main

import (
	"image"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/software"
)

// transitionDuration is how long the galaxy shuffle takes end to end.
const transitionDuration = 1300 * time.Millisecond

// precaptureSlides determines the capture resolution from the live window and
// renders the slides adjacent to the starting position using ensureNeighborsCaptured.
func precaptureSlides(p *presenting) {
	win := p.live
	if currentPresenting != nil && currentPresenting.flipped {
		win = p.control
	}

	// Wait for the window to settle into its final (fullscreen) size before capturing.
	var size fyne.Size
	var pixScale float32
	for i := 0; i < 20; i++ {
		var s fyne.Size
		var px float32
		fyne.DoAndWait(func() {
			s = win.Canvas().Size()
			if s.Width > 0 {
				pw, _ := win.Canvas().PixelCoordinateForPosition(fyne.NewPos(s.Width, 0))
				px = float32(pw) / s.Width
			}
		})
		if s.Width > 0 && s == size && px == pixScale {
			break
		}
		size = s
		pixScale = px
		time.Sleep(50 * time.Millisecond)
	}
	if size.Width <= 0 || size.Height <= 0 {
		return
	}
	if pixScale <= 0 {
		pixScale = 1
	}

	p.captureSize = size
	p.capturePixScale = pixScale
	ensureNeighborsCaptured(p)
}

// ensureNeighborsCaptured renders any of the current slide and its immediate
// neighbours that have not been captured yet. It is safe to call from any
// goroutine.
func ensureNeighborsCaptured(p *presenting) {
	p.captureMu.Lock()
	defer p.captureMu.Unlock()

	if p.captureSize.Width <= 0 {
		return // precaptureSlides has not finished establishing the size yet
	}

	id := p.id
	for _, idx := range [3]int{id, id + 1, id - 1} {
		if idx < 0 || idx >= len(p.items) || p.captures[idx] != nil {
			continue
		}
		captureSlide(p, idx, p.captureSize, p.capturePixScale)
	}
}

// captureSlide renders a single slide off-screen and stores the bitmap in
// p.captures[idx].
func captureSlide(p *presenting, idx int, size fyne.Size, pixScale float32) {
	data := p.items[idx]
	fyne.DoAndWait(func() {
		sl := newSlide(data, idx, p.deck)
		content := container.NewStack(
			canvas.NewRectangle(color.Black),
			newAspectContainer(sl))
		c := software.NewCanvas()
		c.SetPadded(false)
		c.SetScale(pixScale)
		c.SetContent(content)
		c.Resize(size)
		p.captures[idx] = c.Capture()
	})
}

// changeSlide moves the presentation to slide `to`, animating the move with the
// galaxy shuffle shader when the captures are ready (otherwise swapping
// instantly). dir is +1 when advancing and -1 when going back, picking the side
// the incoming card swings out to.
func changeSlide(p *presenting, to int) {
	if p.animating {
		return // ignore navigation while a transition is playing
	}

	from := p.id
	p.id = to
	updatePreviews(p)
	p.updateProgress()

	// Render the new neighbour off-screen so the next transition has its texture
	// ready. Runs in a goroutine so it does not block the current navigation.
	go ensureNeighborsCaptured(p)

	if p.body != nil && from >= 0 && from < len(p.captures) &&
		to >= 0 && to < len(p.captures) && p.captures[from] != nil && p.captures[to] != nil {
		startSlideTransition(p, from, to)
		return
	}

	applyLiveSlide(p)
}

// startSlideTransition overlays the shuffle shader on the live window and drives
// its progress uniform from 0 to 1, then drops back to the real slide.
func startSlideTransition(p *presenting, from, to int) {
	shader := canvas.NewShader("slideShuffle", shuffleShaderSource, shuffleShaderSourceES)
	shader.Textures = map[string]image.Image{
		"current": p.captures[from],
		"next":    p.captures[to],
	}

	dir := 1
	if to < from {
		dir = -1
	}

	shader.Uniforms = map[string]float32{
		"progress":  0,
		"direction": float32(dir),
		"time":      0,
	}

	p.animating = true
	win := p.live
	if currentPresenting != nil && currentPresenting.flipped {
		win = p.control
	}
	win.SetContent(
		container.NewStack(canvas.NewRectangle(color.Black),
			p.body,
			shader,
			container.NewBorder(nil, container.NewStack(p.progressBox,
				container.NewWithoutLayout(p.progressFill)), nil, nil)))

	seconds := float32(transitionDuration.Seconds())
	finished := false
	anim := &fyne.Animation{
		Duration: transitionDuration,
		Curve:    fyne.AnimationLinear,
		Tick: func(done float32) {
			shader.Uniforms["progress"] = done
			shader.Uniforms["time"] = done * seconds
			shader.Refresh()

			if done >= 1 && !finished {
				finished = true
				finishSlideTransition(p)
			}
		},
	}
	anim.Start()
}

// finishSlideTransition swaps the live window back to the real (now current)
// slide and clears the animating flag. The final shader frame already shows the
// incoming slide filling the frame, so the hand off is seamless.
func finishSlideTransition(p *presenting) {
	applyLiveSlide(p)
	p.animating = true
	win := p.live
	if currentPresenting != nil && currentPresenting.flipped {
		win = p.control
	}

	win.SetContent(
		container.NewStack(canvas.NewRectangle(color.Black),
			p.body,
			container.NewBorder(nil, container.NewStack(p.progressBox,
				container.NewWithoutLayout(p.progressFill)), nil, nil)))

	p.animating = false
}

// applyLiveSlide points the live slide widget at the current item.
func applyLiveSlide(p *presenting) {
	p.slide.setSource(p.items[p.id], p.id)
}

// updatePreviews keeps the presenter window's current/next previews in sync.
func updatePreviews(p *presenting) {
	if p.preview == nil {
		return
	}

	p.preview.setSource(p.items[p.id], p.id)
	if p.id < len(p.items)-1 {
		p.next.setSource(p.items[p.id+1], p.id+1)
	} else {
		p.next.setSource("", p.id+1)
	}
	p.updateNotes()
}
