package main

import (
	"image"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// transitionDuration is how long the galaxy shuffle takes end to end.
const transitionDuration = 1300 * time.Millisecond

// precaptureSlides renders every slide onto the live canvas in turn and grabs a
// bitmap of each with Canvas().Capture(), caching them so a transition can hand
// the outgoing and incoming slides to the shader as textures. It runs in its own
// goroutine (Capture reads the painted front buffer, so each slide needs a paint
// before we grab it) and marks the presentation ready once every slide is held.
func precaptureSlides(p *presenting) {
	win := p.live
	if currentPresenting != nil && currentPresenting.flipped {
		win = p.control
	}

	p.progressBox.Hide()
	caps := make([]image.Image, len(p.items))
	for i := range p.items {
		idx := i
		fyne.DoAndWait(func() {
			p.slide.setSource(p.items[idx], idx)
		})
		// Allow a few frames for the slide to reach the front buffer.
		time.Sleep(60 * time.Millisecond)
		fyne.DoAndWait(func() {
			caps[idx] = win.Canvas().Capture()
		})
	}
	fyne.DoAndWait(func() {
		p.slide.setSource(p.items[p.id], p.id)
	})

	p.captures = caps
	p.progressBox.Show()
	p.ready = true
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

	if p.ready && p.body != nil && from >= 0 && from < len(p.captures) &&
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
