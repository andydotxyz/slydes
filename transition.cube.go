package main

// cubeTransition puts the outgoing slide on the front of a box and the incoming
// one on the side of it, and turns the box a quarter turn about its vertical
// axis to bring that side round to the front. Advancing turns the right hand
// side in, so the old slide leaves to the left; going back mirrors it.
//
// A slide is wider than it is tall, so the box is only a cube seen from above:
// as deep as it is wide, which is what lets one face take over exactly where the
// other left off.
//
// As with the flip the box is never projected forwards. Each pixel's view ray is
// turned into the box's own frame, where the two faces are simply the planes
// z = half a width and x = half a width, and met with each in turn. The box is
// convex and we only ever look at the outside of it, so a ray can meet at most
// one of the faces it is in front of, and there is no depth to sort.
var cubeTransition = newShaderLayer("slideCube", "Cube", `
uniform float progress;
uniform float direction;

uniform sampler2D current;
uniform sampler2D next;

// camDist is the camera distance in frame widths. Smaller values exaggerate the
// perspective, so the near edge of the box looms larger as it comes round.
const float camDist = 2.0;

// recede is how far the box backs away from the viewer at the middle of the
// turn, in frame widths. A box turned corner on is wider than the face it
// started with, so with none at all its corners swing out of the frame; the
// more there is the more of the backdrop shows around it.
const float recede = 0.6;

// edgeOn is how brightly a face is lit when it is side on to the viewer, against
// 1.0 for one that is square on.
const float edgeOn = 0.3;

// face is one side of the box where the ray meets it: across and up are how far
// over and how far up the face the ray has landed, t how far along the ray that
// is, and square how nearly square on to the viewer the face is, from 0 to 1.
// Returns the face's colour premultiplied by how much of the pixel it covers.
vec4 face(sampler2D tex, float across, float up, float t, float square, vec2 halfSize) {
	if (t <= 0.0) {
		return vec4(0.0); // behind the camera
	}

	// A pixel is wider on the face the further off the face is and, across it,
	// the more the face is turned away. The edge is a ramp one pixel wide centred
	// on the face's own, so a face that fills the frame covers its outermost
	// pixels completely and the ends are an exact copy of the slide.
	float px = t / frame.x;
	float cov = clamp((halfSize.x - abs(across)) / (px / max(square, 0.05)) + 0.5, 0.0, 1.0)
		* clamp((halfSize.y - abs(up)) / px + 0.5, 0.0, 1.0);
	if (cov <= 0.0) {
		return vec4(0.0);
	}

	vec2 uv = clamp(vec2(0.5 + 0.5 * across / halfSize.x, 0.5 + 0.5 * up / halfSize.y), 0.0, 1.0);
	return vec4(texture2D(tex, uv).rgb * mix(edgeOn, 1.0, square), 1.0) * cov;
}

void main() {
	vec2 frag = fragCoord();

	float p = clamp(progress, 0.0, 1.0);
	// A raised cosine bell: 0 at the ends with zero slope, 1 at the midpoint, so
	// the box eases away and back and is home again by the hand off.
	float bell = 0.5 - 0.5 * cos(2.0 * PI * p);

	// The turn eases in and out as well: a box has weight, and one that set off
	// and stopped dead at full speed would look like a card.
	float ang = direction * 0.5 * PI * p * p * (3.0 - 2.0 * p);

	// As in the flip we work in units of the frame width, not pixels.
	vec2 s = (frag - frame * 0.5) / frame.x;
	vec2 halfSize = vec2(0.5, 0.5 * frame.y / frame.x);

	// The box turns about its own centre, which is half a width behind the
	// screen so that the face at the front lies in the screen and fills it.
	vec3 ro = vec3(0.0, 0.0, camDist + halfSize.x + recede * bell);
	vec3 rd = vec3(s, -camDist);
	ro.xz = rot(ro.xz, -ang);
	rd.xz = rot(rd.xz, -ang);

	vec4 col = vec4(0.0);

	// The front, carrying the outgoing slide. It is only there to be seen while
	// the ray is coming at it from outside the box.
	if (rd.z < 0.0) {
		float t = (halfSize.x - ro.z) / rd.z;
		vec3 hit = ro + t * rd;
		col += face(current, hit.x, hit.y, t, cos(ang), halfSize);
	}

	// The side coming round, carrying the incoming one. The edge it shares with
	// the front is the edge of the slide that arrives first, so the slide runs
	// back along the side from there.
	if (direction * rd.x < 0.0) {
		float t = (direction * halfSize.x - ro.x) / rd.x;
		vec3 hit = ro + t * rd;
		col += face(next, -direction * hit.z, hit.y, t, abs(sin(ang)), halfSize);
	}

	// The two faces meet along an edge but never overlap, so their coverage
	// simply adds: a pixel on the edge is part one and part the other.
	col /= max(col.a, 1.0);
	gl_FragColor = unpremul(col);
}
`)
