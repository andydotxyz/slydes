package main

// Background layers are the opaque backdrop a transition plays in front of.
// They see the shared uniforms plus "time" (seconds since the transition began)
// and know nothing about the slides, so any of them can pair with any movement.

// starfieldBackground is deep space: domain warped nebula clouds, a drifting
// galactic core and three parallax sheets of twinkling stars.
var starfieldBackground = newShaderLayer("slideStarfield", "Starfield", `
uniform float time;

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

void main() {
	gl_FragColor = vec4(galaxy(fragCoord() / frame, time), 1.0);
}
`)

// currentBackground is the backdrop transitions play against. Nil draws no
// backdrop, leaving the transition layer over the window's black background.
var currentBackground = starfieldBackground
