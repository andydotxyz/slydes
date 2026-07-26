package main

import (
	"fyne.io/fyne/v2/canvas"
)

// This file holds the plumbing shared by the GLSL layers that make up a slide
// transition. A transition is drawn as two stacked canvas.Shader objects:
//
//	background - an opaque backdrop (see transition.background.go)
//	transition - the slide movement itself, drawn with alpha over the backdrop
//	             (see transition.shuffle.go)
//
// Two source variants are built for every layer so effects render on both
// desktop OpenGL (core profile, #version 110 - matching Fyne's built in vector
// shaders) and OpenGL ES / mobile / web (#version 100).
//
// Following the canvas.Shader contract, each shader is handed the standard
// uniforms:
//
//	uniform vec2  frame;   // output frame size in pixels
//	uniform vec4  bounds;  // this object's bounds (x1, y1, x2, y2) in pixels
//
// plus the parameters we drive from Go as "uniform float":
//
//	uniform float time;         // seconds since the transition began
//	uniform float progress;     // 0 -> 1 sweep of the transition
//	uniform float direction;    // +1 advancing, -1 going back
//
// and, for transition layers, the two slide captures as textures:
//
//	uniform sampler2D current;  // the slide we are leaving
//	uniform sampler2D next;     // the slide we are arriving at
//
// The vertex stage is Fyne's shared rectangle quad, so there is no texture
// varying: we derive every coordinate from gl_FragCoord and frame, exactly
// as the built in shapes do.

// slideTransitions lists the movements a deck can be presented with, and
// currentTransition is the one in use.
var slideTransitions = []*shaderLayer{shuffleTransition, flipTransition, twistTransition}

var currentTransition = shuffleTransition

// shaderLayer is one compiled-in GLSL effect in the two flavours Fyne needs.
type shaderLayer struct {
	name             string // unique - Fyne caches the compiled program under it
	title            string // human readable, for the menu
	source, sourceES []byte
}

// newShaderLayer prepends the shared prelude to body and wraps the result in
// the desktop and ES headers. body must be valid in both GLSL 1.10 and GLSL ES
// 1.00 (fixed loop bounds, texture2D, gl_FragColor, no derivative functions).
func newShaderLayer(name, title, body string) *shaderLayer {
	src := shaderPrelude + body
	return &shaderLayer{
		name:   name,
		title:  title,
		source: []byte("#version 110\n" + src),
		sourceES: []byte(`#version 100

#ifdef GL_ES
# ifdef GL_FRAGMENT_PRECISION_HIGH
precision highp float;
# else
precision mediump float;
# endif
precision mediump int;
precision lowp sampler2D;
#endif
` + src),
	}
}

// newShader builds a canvas object for this layer. Uniforms and textures are
// filled in by the caller.
func (l *shaderLayer) newShader() *canvas.Shader {
	return canvas.NewShader(l.name, l.source, l.sourceES)
}

// shaderPrelude is the GLSL every layer starts with: the frame uniform plus the
// maths helpers the effects are built from. Helpers a layer does not call are
// dropped by the compiler.
const shaderPrelude = `
uniform vec2 frame;

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

// over composites src on top of dst, both premultiplied by their alpha. Layers
// that draw with transparency accumulate in premultiplied form and unpremul()
// once at the end, matching the src-alpha blend Fyne draws shaders with.
vec4 over(vec4 dst, vec4 src) {
	return src + dst * (1.0 - src.a);
}

vec4 unpremul(vec4 c) {
	if (c.a <= 0.0) {
		return vec4(0.0);
	}
	return vec4(c.rgb / c.a, c.a);
}

// fragCoord returns the pixel position with a top-left origin, so rotation stays
// rigid and texture sampling matches Fyne's image orientation (v = 0 at the top).
vec2 fragCoord() {
	return vec2(gl_FragCoord.x, frame.y - gl_FragCoord.y);
}
`
