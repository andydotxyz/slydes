package main

// pixelateTransition drops the outgoing slide's resolution a step at a time
// until it is a handful of coarse blocks, swaps the slides while the picture is
// held at its blockiest - so what the eye sees is the same grid of squares
// changing colour rather than one slide replacing another - then walks the
// incoming slide back up to full resolution the same way.
//
// The steps are deliberately discrete: the block size is quantised to a fixed
// number of levels rather than growing smoothly, so it reads as a picture losing
// resolution rather than as a zoom.
var pixelateTransition = newShaderLayer("slidePixelate", "Pixelate", `
uniform float progress;
uniform float direction;
uniform float slideRatio;

uniform sampler2D current;
uniform sampler2D next;

// levels is how many steps of pixelation the slide goes through on the way down,
// and again on the way back up.
const float levels = 10.0;

// coarsest is the largest block size, as a fraction of the frame height: 1/14
// leaves the slide as roughly fourteen rows of squares at its blockiest.
const float coarsest = 14.0;

// blockColour averages four taps across a block, which keeps text and rules from
// sparkling as the blocks grow and the single pixel under a block centre swings
// between ink and background. The taps close up completely at a block size of
// one, so the first and last frames sample texel centres exactly and hand over to
// the real slide without a hint of blur. lo and hi pen every tap inside the
// slide, so no block picks up the black border around it.
vec3 blockColour(sampler2D tex, vec2 cell, float block, vec2 lo, vec2 hi) {
	float o = (block - 1.0) * 0.25;
	vec3 c = texture2D(tex, clamp(cell + vec2(-o, -o), lo, hi) / frame).rgb;
	c += texture2D(tex, clamp(cell + vec2(o, -o), lo, hi) / frame).rgb;
	c += texture2D(tex, clamp(cell + vec2(-o, o), lo, hi) / frame).rgb;
	c += texture2D(tex, clamp(cell + vec2(o, o), lo, hi) / frame).rgb;
	return c * 0.25;
}

void main() {
	vec2 frag = fragCoord();

	// The captures are the whole window, with the slide letterboxed inside them at
	// slideRatio and black bars filling the rest. The blocks have to line up with
	// that rectangle and stop at its edge: a block straddling the boundary would
	// otherwise average slide and bar together and smear the slide out into the
	// border. Whole pixels, so that a block size of one still lands exactly on
	// texel centres and the ends stay an exact copy of the slide.
	vec2 slide = vec2(frame.x, frame.x / slideRatio);
	if (frame.x > frame.y * slideRatio) {
		slide = vec2(frame.y * slideRatio, frame.y);
	}
	slide = floor(slide);
	vec2 origin = floor((frame - slide) * 0.5);

	vec2 rel = frag - origin;
	if (rel.x < 0.0 || rel.y < 0.0 || rel.x >= slide.x || rel.y >= slide.y) {
		gl_FragColor = vec4(0.0, 0.0, 0.0, 1.0); // the border stays solid black
		return;
	}

	float p = clamp(progress, 0.0, 1.0);

	// Coarseness runs 0 -> 1 -> 0 across the transition, so both ends are the
	// slide at full resolution and the middle is the blockiest.
	float k = 1.0 - abs(1.0 - 2.0 * p);

	// Quantised to whole steps. Dividing by levels but stepping over levels + 1
	// gives the coarsest step a window of its own to sit in rather than a single
	// instant, so the slides have time to swap while the picture is held there.
	float level = min(floor(k * (levels + 1.0)), levels) / levels;

	// Block size grows geometrically - each step multiplies rather than adds -
	// which is what makes the drop in resolution look even. Whole pixels only, so
	// block edges land on the same boundary all over the frame.
	float block = max(floor(pow(slide.y / coarsest, level)), 1.0);

	// Blocks are laid out from the slide's own corner, not the frame's, so the
	// grid meets the border squarely however the window is letterboxed.
	vec2 cell = origin + (floor(rel / block) + 0.5) * block;

	// Keep the taps a pixel clear of the border once the blocks are big enough to
	// reach it, in case the capture's letterbox edge lands half a pixel off where
	// this works it out. At a block size of one the margin has to close up again,
	// or the outermost row of the slide would sample its neighbour.
	float margin = 1.5;
	if (block <= 1.0) {
		margin = 0.5;
	}
	vec2 lo = origin + margin;
	vec2 hi = origin + slide - margin;

	// The swap happens at the midpoint, buried inside the coarsest step.
	vec3 col;
	if (p < 0.5) {
		col = blockColour(current, cell, block, lo, hi);
	} else {
		col = blockColour(next, cell, block, lo, hi);
	}

	gl_FragColor = vec4(col, 1.0);
}
`)
