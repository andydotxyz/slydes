package main

// flipTransition prints the incoming slide on the back of the outgoing one and
// turns the pair through a half turn about the card's vertical axis. Advancing
// sends the right edge away from the viewer first, so the flip reads right to
// left; going back mirrors it.
//
// The card is a flat quad in 3D, so rather than projecting it forwards we invert
// the projection per pixel: given a point on screen, work out which point across
// the card lands there. For a rotation of ang about the vertical axis, a point u
// across the card sits at x = u*cos(ang), z = u*sin(ang), and a camera at
// distance d scales it by d / (d - z), so
//
//	x = u*cos(ang) * d / (d - u*sin(ang))
//
// which rearranges to the u = x*d / (d*cos(ang) + x*sin(ang)) used below.
var flipTransition = newShaderLayer("slideFlip", "Flip", `
uniform float progress;
uniform float direction;

uniform sampler2D current;
uniform sampler2D next;

// camDist is the camera distance in frame widths. Smaller values exaggerate the
// perspective; larger ones flatten it towards a plain orthographic flip.
const float camDist = 2.0;

void main() {
	vec2 frag = fragCoord();

	float p = clamp(progress, 0.0, 1.0);
	// A raised cosine bell: 0 at the ends, 1 at the midpoint, and with zero slope
	// at p == 0 and p == 1, so the flourishes ease in and out and have completely
	// vanished by the hand off to the real slide.
	float bell = 0.5 - 0.5 * cos(2.0 * PI * p);

	float ang = -direction * PI * p;
	float ca = cos(ang);
	float sa = sin(ang);

	// The card pulls back a little mid flip so the backdrop shows around it, and
	// carries a slight in plane tilt so it reads as thrown rather than hinged.
	float base = mix(1.0, 0.86, bell);
	float tilt = direction * bell * 0.05;

	// From here on we work in units of the frame width rather than pixels: the
	// projection multiplies coordinates together, and pixel sized numbers can
	// overflow mediump floats on lower end GL ES parts.
	vec2 s = rot(frag - frame * 0.5, -tilt) / frame.x;
	vec2 halfSize = vec2(0.5, 0.5 * frame.y / frame.x) * base;

	// Which point across the card projects onto this pixel?
	float den = camDist * ca + s.x * sa;
	if (abs(den) < 0.0001) {
		gl_FragColor = vec4(0.0); // edge on - the card has no width to draw
		return;
	}
	float u = s.x * camDist / den;

	float depth = camDist - u * sa; // distance from the camera plane
	if (depth < 0.0001) {
		gl_FragColor = vec4(0.0); // that part of the card has swung behind us
		return;
	}
	float scale = camDist / depth;
	float v = s.y / scale;

	// Antialiased coverage of the card, the edge softness carried back through the
	// projection so it stays about a pixel wide on screen however the card leans.
	float e = 1.5 / (frame.x * scale);
	float cov = (1.0 - smoothstep(halfSize.x - e, halfSize.x + e, abs(u)))
		* (1.0 - smoothstep(halfSize.y - e, halfSize.y + e, abs(v)));
	if (cov <= 0.0) {
		gl_FragColor = vec4(0.0);
		return;
	}

	// Once the card is more than a quarter turn round we are looking at its back,
	// which is the incoming slide. Sampling it mirrored across x lands it the
	// right way round for a viewer looking through the card.
	vec2 uv = clamp(vec2(0.5 + 0.5 * u / halfSize.x, 0.5 + 0.5 * v / halfSize.y), 0.0, 1.0);
	vec3 tc = texture2D(current, uv).rgb;
	if (ca < 0.0) {
		tc = texture2D(next, vec2(1.0 - uv.x, uv.y)).rgb;
	}

	// Depth shading: the receding half darkens, the half swinging towards us
	// lifts, and the whole face dims as it turns edge on. Every term is scaled by
	// sin(ang), so the first and last frames match the real slide exactly.
	float near = u * sa / halfSize.x;
	float light = (1.0 - 0.22 * abs(sa)) * (1.0 + 0.28 * near);

	// One opaque element, so straight alpha is exact here - no premultiplied
	// accumulator needed.
	gl_FragColor = vec4(tc * light, cov);
}
`)
