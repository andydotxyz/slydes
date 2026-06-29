package main

// This file holds the GLSL fragment shaders that power the "galaxy shuffle"
// slide transition (see transition.go). Two source variants are kept so the
// effect renders on both desktop OpenGL (core profile, #version 110 - matching
// Fyne's built in vector shaders) and OpenGL ES / mobile / web (#version 100).
//
// Following the canvas.Shader contract, each shader is handed the standard
// uniforms:
//
//	uniform vec2  frame;   // output frame size in pixels
//	uniform vec4  bounds;  // this object's bounds (x1, y1, x2, y2) in pixels
//
// plus the parameters we drive from Go as "uniform float":
//
//	uniform float time;         // seconds since the transition began (galaxy motion)
//	uniform float progress;     // 0 -> 1 sweep of the transition
//	uniform float direction;    // +1 advancing, -1 going back (which side the card swings to)
//
// and the two slide captures as textures:
//
//	uniform sampler2D current;  // the slide we are leaving
//	uniform sampler2D next;     // the slide we are arriving at
//
// The vertex stage is Fyne's shared rectangle quad, so there is no texture
// varying: we derive every coordinate from gl_FragCoord and frame, exactly
// as the built in shapes do.

// shuffleShaderBody is the GLSL shared between the two targets. It is valid in
// both GLSL 1.10 and GLSL ES 1.00 (fixed loop bounds, texture2D, gl_FragColor,
// no derivative functions).
const shuffleShaderBody = `
uniform vec2 frame;
uniform vec4 rect_coords;
uniform float time;
uniform float progress;
uniform float direction;

uniform sampler2D current;
uniform sampler2D next;

const float PI = 3.14159265;

vec2 rot(vec2 v, float a) {
	float c = cos(a);
	float s = sin(a);
	return vec2(v.x * c - v.y * s, v.x * s + v.y * c);
}

float hash21(vec2 p) {
	p = fract(p * vec2(123.34, 456.21));
	p += dot(p, p + 45.32);
	return fract(p.x * p.y);
}

vec3 hash23(vec2 p) {
	return vec3(hash21(p), hash21(p + 1.7), hash21(p + 3.3));
}

float valueNoise(vec2 p) {
	vec2 i = floor(p);
	vec2 f = fract(p);
	f = f * f * (3.0 - 2.0 * f);
	float a = hash21(i);
	float b = hash21(i + vec2(1.0, 0.0));
	float c = hash21(i + vec2(0.0, 1.0));
	float d = hash21(i + vec2(1.0, 1.0));
	return mix(mix(a, b, f.x), mix(c, d, f.x), f.y);
}

float fbm(vec2 p) {
	float v = 0.0;
	float a = 0.5;
	for (int i = 0; i < 5; i++) {
		v += a * valueNoise(p);
		p *= 2.0;
		a *= 0.5;
	}
	return v;
}

// starLayer scatters one parallax sheet of twinkling stars across the plane.
float starLayer(vec2 uv, float t) {
	float total = 0.0;
	vec2 gv = fract(uv) - 0.5;
	vec2 id = floor(uv);
	for (int y = -1; y <= 1; y++) {
		for (int x = -1; x <= 1; x++) {
			vec2 offs = vec2(float(x), float(y));
			vec3 h = hash23(id + offs);
			vec2 pos = offs + vec2(h.x, h.y) - 0.5;
			float d = length(gv - pos);
			float bright = h.z * h.z;
			float twinkle = 0.6 + 0.4 * sin(t * 3.0 + h.z * 6.2831);
			total += smoothstep(0.07 * bright, 0.0, d) * bright * twinkle;
		}
	}
	return total;
}

// galaxy renders the deep space backdrop for normalized frame coord q (y down).
vec3 galaxy(vec2 q, float t) {
	float aspect = frame.x / frame.y;
	vec2 p = vec2(q.x * aspect, q.y);

	// Nebula clouds from domain warped fbm.
	vec2 np = p * 3.0 + vec2(t * 0.02, t * 0.015);
	float n = fbm(np + fbm(np * 0.5));
	vec3 col = vec3(0.02, 0.01, 0.05);
	col += vec3(0.28, 0.06, 0.5) * pow(n, 3.0) * 0.9;
	col += vec3(0.04, 0.13, 0.42) * pow(fbm(np * 1.7 + 5.0), 2.0) * 0.55;

	// A soft galactic core that drifts slowly.
	vec2 core = vec2(0.5 * aspect, 0.5) + 0.12 * vec2(sin(t * 0.1), cos(t * 0.08));
	col += vec3(0.32, 0.22, 0.38) * smoothstep(0.65, 0.0, length(p - core)) * 0.5;

	// Three parallax star sheets drifting at different speeds.
	float s = starLayer(p * 8.0 + vec2(t * 0.05, 0.0), t);
	s += starLayer(p * 16.0 + vec2(t * 0.10, 1.3), t) * 0.7;
	s += starLayer(p * 28.0 - vec2(t * 0.15, 2.1), t) * 0.5;
	col += vec3(s);

	return col;
}

// drawCard composites a slide "card" (a texture in its own rotated/scaled box)
// over the accumulated colour, including a soft directional drop shadow so the
// card reads as floating in front of the galaxy.
vec3 drawCard(vec3 bg, vec2 frag, vec2 center, vec2 halfSize, float ang, sampler2D tex) {
	// Drop shadow: the card's box cast down and to the right.
	vec2 shOff = vec2(0.012, 0.02) * frame.y;
	vec2 sLocal = rot((frag - shOff) - center, -ang);
	float sDist = length(max(abs(sLocal) - halfSize, 0.0));
	float shadow = exp(-sDist / (0.05 * frame.y));
	bg = mix(bg, bg * 0.12, clamp(shadow, 0.0, 1.0) * 0.6);

	// Card face.
	vec2 local = rot(frag - center, -ang);
	vec2 uv = clamp(local / halfSize * 0.5 + 0.5, 0.0, 1.0);
	vec3 tc = texture2D(tex, uv).rgb;

	// Antialiased coverage of the card rectangle (no derivative funcs needed).
	float e = 1.5;
	float covx = 1.0 - smoothstep(halfSize.x - e, halfSize.x + e, abs(local.x));
	float covy = 1.0 - smoothstep(halfSize.y - e, halfSize.y + e, abs(local.y));
	return mix(bg, tc, covx * covy);
}

void main() {
	// Work in pixels with a top-left origin so rotation stays rigid and texture
	// sampling matches Fyne's image orientation (v = 0 at the top).
	vec2 frag = vec2(gl_FragCoord.x, frame.y - gl_FragCoord.y);

	vec3 col = galaxy(frag / frame, time);

	float p = clamp(progress, 0.0, 1.0);
	// A raised cosine bell: 0 at the ends, 1 at the midpoint, and crucially with
	// zero slope at p == 0 and p == 1, so the cards ease into and out of motion
	// instead of lurching the instant the transition starts or stops.
	float bell = 0.5 - 0.5 * cos(2.0 * PI * p);

	// Both cards recede from the viewer mid transition (base 1.0 -> 0.8 -> 1.0)
	// so more of the galaxy shows around them; at p == 0 / p == 1 the active card
	// still exactly fills the frame for a seamless hand off to the real slide.
	float base = mix(1.0, 0.8, bell);
	vec2 center = frame * 0.5;

	// The current slide eases the opposite way to the incoming one - a small
	// drift, tilt and the shared recede - so both feel in motion rather than one
	// card sliding over a static backdrop. It is the top card until the incoming
	// slide swings over it.
	float curTravel = 0.12 * frame.x;
	vec2 curCenter = center + vec2(-direction * bell * curTravel, bell * 0.05 * frame.y);
	float curAng = -direction * bell * 0.05;
	vec2 curHalf = frame * 0.5 * base;

	// The incoming slide starts hidden directly behind the current one, swings
	// out to one side (revealing the galaxy behind), then slides back to land
	// squarely on top - a card shuffled from the back of the deck to the front.
	// travel must be large enough that at the midpoint (where the z order flips)
	// the incoming card has fully cleared the current card's bounds, otherwise
	// the two overlap at the crossover and the incoming appears to push straight
	// through the current slide instead of swinging around it.
	float travel = 0.82 * frame.x;
	vec2 inCenter = center + vec2(direction * bell * travel, -bell * 0.09 * frame.y);
	float inAng = direction * bell * 0.13;
	vec2 inHalf = frame * 0.5 * base * (1.0 + 0.06 * bell);

	// Z order swaps at the midpoint, where the cards barely overlap so the
	// change of stacking is invisible.
	if (p < 0.5) {
		col = drawCard(col, frag, inCenter, inHalf, inAng, next);
		col = drawCard(col, frag, curCenter, curHalf, curAng, current);
	} else {
		col = drawCard(col, frag, curCenter, curHalf, curAng, current);
		col = drawCard(col, frag, inCenter, inHalf, inAng, next);
	}

	gl_FragColor = vec4(col, 1.0);
}
`

// shuffleShaderSource is the desktop (OpenGL core profile) variant.
var shuffleShaderSource = []byte("#version 110\n" + shuffleShaderBody)

// shuffleShaderSourceES is the OpenGL ES / mobile / web variant.
var shuffleShaderSourceES = []byte(`#version 100

#ifdef GL_ES
# ifdef GL_FRAGMENT_PRECISION_HIGH
precision highp float;
# else
precision mediump float;
# endif
precision mediump int;
precision lowp sampler2D;
#endif
` + shuffleShaderBody)
