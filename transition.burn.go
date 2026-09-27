package main

// burnTransition sets light to the outgoing slide along one edge and lets it
// burn across to the other, leaving the incoming slide where it has been. The
// paper browns and blackens ahead of the fire, the edge of what is left glows,
// and the glow lingers for a moment on the slide behind.
//
// Every point of the slide is given the moment it catches, part from how far
// across it is - so the fire does travel, from the right when advancing - and
// part from a field of noise, which is what makes the edge ragged and lets the
// fire run ahead in tongues and leave islands behind. Progress is then just a
// threshold moving through those moments, and how far a point is from the
// threshold says what state it is in: untouched, scorched, alight or gone.
var burnTransition = newShaderLayer("slideBurn", "Burn", `
uniform float progress;
uniform float direction;
uniform float time;
uniform float slideRatio;

uniform sampler2D current;
uniform sampler2D next;

// sweep is how much of the order things catch in comes from position rather
// than noise. At 1.0 the fire is a straight line crossing the slide, at 0.0 it
// breaks out everywhere at once.
const float sweep = 0.7;

// grain is the size of the largest tongues and islands, in the number of them
// that fit up the slide.
const float grain = 2.6;

// scorch is how far ahead of the fire the paper is already browning, and ember
// how long the glow lingers behind it, both as shares of the whole transition.
// glow is how much of the scorched band is alight rather than just blackened.
const float scorch = 0.16;
const float ember = 0.12;
const float glow = 0.45;

// flicker is how much the fire's brightness wavers, 0.0 for a steady glow.
const float flicker = 0.5;

// The fire, from its coolest to its hottest.
const vec3 fireRed = vec3(0.85, 0.12, 0.0);
const vec3 fireOrange = vec3(1.0, 0.5, 0.05);
const vec3 fireWhite = vec3(1.0, 0.92, 0.65);

// fire is the colour of something burning with the given heat, 0 to 1.
vec3 fire(float heat) {
	vec3 c = mix(fireRed, fireOrange, smoothstep(0.0, 0.6, heat));
	return mix(c, fireWhite, smoothstep(0.6, 1.0, heat)) * heat;
}

void main() {
	vec2 frag = fragCoord();

	// The captures are the whole window with the slide letterboxed inside them,
	// and it is only the slide that is paper: the border stays solid black.
	vec2 slide = vec2(frame.x, frame.x / slideRatio);
	if (frame.x > frame.y * slideRatio) {
		slide = vec2(frame.y * slideRatio, frame.y);
	}
	slide = floor(slide);
	vec2 rel = frag - floor((frame - slide) * 0.5);
	if (rel.x < 0.0 || rel.y < 0.0 || rel.x >= slide.x || rel.y >= slide.y) {
		gl_FragColor = vec4(0.0, 0.0, 0.0, 1.0);
		return;
	}

	float p = clamp(progress, 0.0, 1.0);

	// When this point catches, 0 to 1. The noise is laid out in slide heights
	// both ways, so the tongues are as wide as they are tall.
	float across = rel.x / slide.x;
	if (direction >= 0.0) {
		across = 1.0 - across;
	}
	vec2 q = rel / slide.y;

	// The noise keeps to the middle of its range, and left like that the fire
	// would spend the start and the end of the transition with nothing to burn.
	float noise = clamp((fbm(q * grain + 7.3) - 0.18) / 0.62, 0.0, 1.0);
	float catches = mix(noise, across, sweep);

	// The threshold sets off far enough back that nothing is even scorched, and
	// finishes far enough on that the last ember has gone out, so both ends are
	// an exact copy of the slide.
	float d = catches - mix(-scorch, 1.0 + ember, p);

	vec3 paper = texture2D(current, frag / frame).rgb;
	vec3 under = texture2D(next, frag / frame).rgb;

	// Ahead of the fire the paper browns, then blackens as the fire reaches it.
	float heat = 1.0 - smoothstep(0.0, scorch, d);
	paper *= mix(vec3(1.0), vec3(0.72, 0.5, 0.3), heat);
	paper = mix(paper, vec3(0.04, 0.025, 0.02), heat * heat);

	// The fire itself wavers, in patches that drift up the slide as flames do.
	float waver = 1.0 - flicker + 2.0 * flicker * valueNoise(q * 26.0 + vec2(0.0, 5.0 * time));

	vec3 col;
	if (d > 0.0) {
		// Only the part nearest the edge is alight.
		float alight = 1.0 - smoothstep(0.0, glow * scorch, d);
		col = paper + fire(alight * alight) * waver;
	} else {
		float linger = 1.0 - smoothstep(0.0, ember, -d);
		col = under + fire(0.75 * linger * linger * linger) * waver;
	}

	gl_FragColor = vec4(min(col, 1.0), 1.0);
}
`)
