package main

// twistTransition wrings the slide like a sweet wrapper. The slide is a sheet
// twisting about its own horizontal centre line, but the twist is not applied
// evenly: how far a column has turned depends on where it is, and that gradient
// migrates across the frame. Columns the wring has passed have turned their full
// half turn and show the incoming slide (printed on the back of the sheet),
// columns ahead of it are still the outgoing slide, and in between the sheet
// leans through edge on, letting the backdrop past it.
//
// A sheet of paper does not twist as a stack of flat slats though - it curls,
// most of all where it is closest to edge on. So each column's cross section is
// not a straight line but a cubic curve: the material rolls away from the plane
// of the sheet, hardest at the top and bottom edges. That is what makes the
// twist read as a surface rather than a rotating slab, and it means a single
// column can show its front near the middle and its back at the edges, with the
// two overlapping where the curl folds over itself.
//
// A curved cross section cannot be inverted in closed form the way a flat one
// can, so instead of solving for the point that lands on this pixel we walk
// along the sheet looking for it: march the material coordinate from one edge to
// the other, project each step, and keep the crossing nearest the camera. That
// also sorts the folds correctly for free - whichever layer of the curl is in
// front is the one that gets drawn.
var twistTransition = newShaderLayer("slideTwist", "Twist", `
uniform float progress;
uniform float direction;

uniform sampler2D current;
uniform sampler2D next;

// camDist is the camera distance in frame widths - the smaller it is the more
// the near side of the twist bulges towards the viewer.
const float camDist = 2.0;

// twistBand is how far the turn is spread out, as a fraction of the slide width.
// Around 0.5 and above the whole slide is in motion at once, wringing lazily;
// small values gather it into a tight travelling knot.
const float twistBand = 0.55;

// curlMax is how far the sheet's edges roll out of its own plane, in half
// heights, when a column is fully edge on.
const float curlMax = 0.6;

// steps is how finely the sheet is walked looking for the point under this
// pixel. It only has to resolve the folds in the curl, so it can stay low.
const int steps = 32;

void main() {
	vec2 frag = fragCoord();

	float p = clamp(progress, 0.0, 1.0);
	// A raised cosine bell: 0 at the ends with zero slope, 1 at the midpoint, so
	// the flourishes ease in and out and have gone by the hand off.
	float bell = 0.5 - 0.5 * cos(2.0 * PI * p);

	float base = mix(1.0, 0.94, bell);
	vec2 s = (frag - frame * 0.5) / frame.x;
	vec2 halfSize = vec2(0.5, 0.5 * frame.y / frame.x) * base;
	float h = halfSize.y;

	// The column this pixel sits in, measured so 1 is the edge the wring starts
	// from and 0 the edge it finishes at. Advancing wrings from the right (the
	// same way round as the flip transition), going back from the left.
	float u = s.x;
	float col = 0.5 + 0.5 * u / halfSize.x;
	if (direction < 0.0) {
		col = 1.0 - col;
	}

	// The turn migrates from just off the starting edge to just off the far one,
	// so at p == 0 no column has turned and at p == 1 every column has.
	float centre = mix(1.0 + twistBand, -twistBand, p);
	float ang = PI * smoothstep(centre - twistBand, centre + twistBand, col);
	float ca = cos(ang);
	float sa = sin(ang);

	// The curl peaks where the column is edge on and vanishes where it lies flat,
	// so the first and last frames are the plain slide.
	float curl = curlMax * sa;

	// Walk the sheet from its bottom edge to its top, projecting as we go, and
	// keep the crossing of this pixel's row that is nearest the camera.
	float hitV = 0.0;
	float hitZ = -1000.0;
	bool hit = false;
	float nearV = 0.0;      // closest approach, for the silhouette's soft edge
	float nearDist = 1000.0;
	float prevOff = 0.0;
	float prevV = 0.0;
	bool prevOK = false;
	for (int i = 0; i <= steps; i++) {
		float v = mix(-h, h, float(i) / float(steps));
		float m = v / h;
		float b = curl * h * m * m * m;   // roll out of the sheet's own plane
		float y = v * ca - b * sa;
		float z = v * sa + b * ca;
		float depth = camDist - z;
		if (depth < 0.0001) {
			prevOK = false; // this stretch has swung behind the camera
		} else {
			float off = y * camDist / depth - s.y;
			if (abs(off) < nearDist) {
				nearDist = abs(off);
				nearV = v;
			}
			if (prevOK && off * prevOff <= 0.0 && abs(off - prevOff) > 0.000001) {
				float t = prevOff / (prevOff - off);
				float vc = mix(prevV, v, t);
				float mc = vc / h;
				float zc = vc * sa + curl * h * mc * mc * mc * ca;
				if (zc > hitZ) {
					hitZ = zc;
					hitV = vc;
					hit = true;
				}
			}
			prevOK = true;
			prevOff = off;
			prevV = v;
		}
	}

	// Where the row misses the sheet entirely we are just past its silhouette, so
	// fade out over the last pixel rather than cutting hard.
	float aa = 1.5 / frame.x;
	float cov = (1.0 - smoothstep(halfSize.x - aa, halfSize.x + aa, abs(u)));
	if (!hit) {
		cov *= 1.0 - smoothstep(0.0, aa, nearDist);
	}
	if (cov <= 0.0) {
		gl_FragColor = vec4(0.0);
		return;
	}

	float v = hit ? hitV : nearV;
	float m = v / h;
	float b = curl * h * m * m * m;
	float slope = 3.0 * curl * m * m;   // how fast the roll is turning here
	float z = v * sa + b * ca;

	// The surface normal from the cross section's tangent. x is untouched by the
	// twist, so the tangent along the column is all we need.
	vec3 tangent = vec3(0.0, ca - slope * sa, sa + slope * ca);
	vec3 n = normalize(vec3(0.0, -tangent.z, tangent.y));

	// Whichever way this scrap of sheet faces decides which slide it carries: the
	// back is the incoming one, printed upside down so a half turn about the
	// horizontal axis lands it the right way up.
	vec2 uv = clamp(vec2(0.5 + 0.5 * u / halfSize.x, 0.5 + 0.5 * m), 0.0, 1.0);
	vec3 tc;
	if (n.z >= 0.0) {
		tc = texture2D(current, uv).rgb;
	} else {
		tc = texture2D(next, vec2(uv.x, 1.0 - uv.y)).rgb;
		n = -n; // shade by the face we can actually see
	}

	// Shading off that normal: the sheet dims as it rolls away from us, faces
	// tipped up catch more light than faces tipped down, material nearer the
	// camera lifts, and a sheen rides the steepest part of the curl. Flat, square
	// on material comes out at exactly 1.0, so the ends match the real slide.
	float tilt = 1.0 - abs(n.z);
	float light = (1.0 - 0.34 * tilt) * (1.0 - 0.3 * n.y) * (1.0 + 0.22 * z / h)
		+ 0.2 * pow(tilt, 5.0);

	// One opaque element, so straight alpha is exact here.
	gl_FragColor = vec4(tc * light, cov);
}
`)
