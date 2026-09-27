package main

import "time"

// paperPlaneTransition folds the outgoing slide into a paper aeroplane where it
// lies, uncovering the incoming slide underneath as the paper is drawn in, then
// throws it off the side of the screen. It is the classic dart, nose pointing
// the way the deck is moving, so the slide makes an arrow of itself and then
// follows it:
//
//  1. the two leading corners fold in to the centre line, making the point
//  2. the sloping edges that leaves fold in to the centre line again
//  3. the body is pinched up along the centre line into a keel, which draws the
//     wings in either side of it
//  4. it lifts off the slide, banks, and leaves
//
// The sheet is only ever a handful of flat facets, each hinged to the next along
// a crease, so nothing here needs marching: for every facet we intersect the
// view ray with the plane it lies in, work out which scrap of the sheet that is,
// and keep whichever hit is nearest the camera.
//
// Both halves of the sheet fold the same way, so all the folding is described
// once for a half sheet with the centre line along its lower edge, in "half
// coordinates": x towards the nose, y away from the centre line. The top half of
// the screen and the bottom are that one description mirrored, and going back
// rather than advancing mirrors it left to right as well.
//
// Every facet is a hard edged polygon, so rather than soften each edge
// analytically - and have to get it right where two facets meet along a crease -
// the paper is supersampled. The slide underneath, and the body of the sheet
// while it is still lying flat, are looked up at the pixel centre all the same,
// so nothing is blurred that has not moved.
var paperPlaneTransition = newShaderLayer("slidePaperPlane", "Paper plane", `
uniform float progress;
uniform float direction;
uniform float slideRatio;

uniform sampler2D current;
uniform sampler2D next;

// Lengths throughout are in slide half heights, so the sheet is 2.0 tall and
// 2.0 * slideRatio wide whatever the resolution.

// camDist is the camera distance. Smaller values exaggerate how much a lifted
// corner, and the aeroplane as it climbs, loom towards the viewer.
const float camDist = 5.2;

// The timetable, as shares of the whole transition. Each corner takes foldTime
// to come over, the bottom half trails the top by stagger (0.0 folds both sides
// together), and rest is the beat between one fold landing and the next
// starting. The folds have to be over by the time the pinch begins, so the pinch
// is scheduled from them rather than given a time of its own.
const float foldTime = 0.19;
const float stagger = 0.04;
const float rest = 0.03;
const float fold2At = foldTime + rest;
const float pinchAt = fold2At + foldTime + stagger + rest;
const float pinchTime = 0.2;

// flightAt is when the aeroplane leaves the slide. Earlier than the end of the
// pinch has it still closing up as it goes.
const float flightAt = 0.58;

// keel is how much of each half goes into the body rather than the wing, and
// keelAngle how far the body closes up: at PI / 2 the two sides shut flat
// against each other and the wings meet, short of that a groove is left between
// them. dihedral is how far the wing tips are lifted above the wing roots.
const float keel = 0.3;
const float keelAngle = 1.4;
const float dihedral = 0.3;

// The flight. A sheet the size of the slide makes an aeroplane the size of the
// screen, which would be out of the frame nose first before it had been seen as
// one, so it flies away from the viewer as well as off to the side: farOff is
// how far away it has got by the time it leaves, and 0.0 keeps it skimming the
// slide at full size. swoop is how far up the screen its path bends, bank the
// most it rolls into that bend and pitchUp how far it lifts its nose to take
// off, both in radians.
const float farOff = 2.6;
const float swoop = 1.2;
const float bank = 0.9;
const float pitchUp = 0.35;

// The back of the sheet: paperTint, with showThrough of the printing on the
// other side visible through it.
const vec3 paperTint = vec3(0.86, 0.85, 0.82);
const float showThrough = 0.12;

// lightDir points at the light, which is over the viewer's left shoulder: x
// right, y down the screen, z out of it. ambient is how bright a face turned
// right away from the light still is, and paper lying flat is always exactly 1.0
// so that the ends match the real slides.
const vec3 lightDir = vec3(-0.35, -0.45, 0.82);
const float ambient = 0.55;

// shadowDepth is how dark the shadow under the paper is, shadowSoft how far its
// edge spreads where the paper is resting on the slide, and shadowSpread how
// much further it spreads for every unit the paper is lifted. shadowRise is how
// high over the slide the aeroplane's shadow has it climbing, and shadowGone how
// far through the flight the shadow has faded out altogether.
const float shadowDepth = 0.5;
const float shadowSoft = 0.035;
const float shadowSpread = 0.22;
const float shadowRise = 2.5;
const float shadowGone = 0.6;

// The two creases both run from the nose: the first at 45 degrees to the centre
// line and the second at 22.5. along is the direction of the crease and away the
// direction, at right angles to it, of the paper that is going to be folded.
const vec2 along1 = vec2(-0.70710678, 0.70710678);
const vec2 away1 = vec2(0.70710678, 0.70710678);
const vec2 along2 = vec2(-0.92387953, 0.38268343);
const vec2 away2 = vec2(0.38268343, 0.92387953);

// Worked out once in main.
float p = 0.0;             // progress
float halfW = 1.0;         // half the width of the sheet
float unit = 1.0;          // pixels to a slide half height
vec2 centre = vec2(0.0);   // the middle of the slide, in pixels
vec3 flatInk = vec3(0.0);  // the outgoing slide at this pixel, where it has not moved
vec3 at = vec3(0.0);       // where the aeroplane has got to
float pinch = 0.0;         // how far the keel has closed
float tips = 0.0;          // how far the wing tips have lifted
float yaw = 0.0;
float pitch = 0.0;
float roll = 0.0;

// front is the printed side of the sheet at a point on one half of it, and back
// the other side of the same scrap.
vec3 front(vec2 g, float side) {
	vec2 px = centre + vec2(direction * g.x, side * g.y) * unit;
	return texture2D(current, px / frame).rgb;
}

vec3 back(vec2 g, float side) {
	return mix(paperTint, front(g, side), showThrough);
}

// lit is how brightly a face is lit, n being the normal on the side we can see.
float lit(vec3 n, vec3 l) {
	return min(ambient + (1.0 - ambient) * max(dot(n, l), 0.0) / lightDir.z, 1.3);
}

// lightFor is the light as one half of the sheet sees it.
vec3 lightFor(float side) {
	return vec3(lightDir.x * direction, lightDir.y * side, lightDir.z);
}

// turned1 and turned2 are how far through its first and second fold a half is,
// 0 to 1. The top half leads.
float turned1(float side) {
	float from = side < 0.0 ? 0.0 : stagger;
	return smoothstep(from, from + foldTime, p);
}

float turned2(float side) {
	float from = fold2At + (side < 0.0 ? 0.0 : stagger);
	return smoothstep(from, from + foldTime, p);
}

// stageOf numbers where a half has got to:
//
//	0  untouched
//	1  first corner in the air
//	2  first corner down
//	3  second corner in the air
//	4  both down - the finished dart
int stageOf(float side) {
	float a = turned1(side);
	float b = turned2(side);
	if (a <= 0.0) {
		return 0;
	}
	if (a < 1.0) {
		return 1;
	}
	if (b <= 0.0) {
		return 2;
	}
	if (b < 1.0) {
		return 3;
	}
	return 4;
}

// across1 and across2 reflect a point in the first and second crease, which is
// what folding the paper flat does to it - and, reflection being its own
// inverse, how to get from where a folded scrap is lying back to where on the
// sheet it came from.
vec2 across1(vec2 g) {
	return vec2(halfW - g.y, halfW - g.x);
}

vec2 across2(vec2 g) {
	return g - 2.0 * dot(g - vec2(halfW, 0.0), away2) * away2;
}

// edgeShade darkens the sheet in the lee of a corner folded down on top of it,
// d being the distance from that corner's loose edge. It is all that separates
// one layer of white paper from another once both are lying flat.
float edgeShade(float d) {
	return 1.0 - 0.3 * exp(-max(d, 0.0) / 0.04);
}

// stack is whatever is uppermost of the paper lying flat at g, for a half that
// has reached the given stage. Corners in the air are not part of it. ink is the
// sheet's own printing at g. Alpha is 0 where there is no paper.
vec4 stack(vec2 g, float side, int stage, vec3 ink) {
	if (abs(g.x) > halfW || g.y < 0.0 || g.y > 1.0) {
		return vec4(0.0);
	}
	if (stage == 0) {
		return vec4(ink, 1.0);
	}

	vec2 r = g - vec2(halfW, 0.0);
	if (dot(r, away1) > 0.0) {
		return vec4(0.0); // the first fold took this away
	}
	if (stage == 1) {
		return vec4(ink, 1.0);
	}
	if (stage >= 3 && dot(r, away2) > 0.0) {
		return vec4(0.0); // and the second this
	}

	if (stage == 4) {
		// The second corner covers the first completely, so it is the only one
		// there is to see. Its loose edge is the sheet's top edge, brought round.
		vec2 q = across2(g);
		if (q.y <= 1.0) {
			return vec4(back(q, side), 1.0);
		}
		return vec4(ink * edgeShade((q.y - 1.0)), 1.0);
	}

	// The first corner, lying where it landed. Its loose edge is what was the
	// leading edge of the sheet.
	if (g.x >= halfW - 1.0) {
		return vec4(back(across1(g), side), 1.0);
	}
	return vec4(ink * edgeShade(halfW - 1.0 - g.x), 1.0);
}

// hinge finds where a ray meets the corner a half has in the air. along and away
// describe the crease it turns about and phi is how far it has turned. q is set
// to where that scrap lay before the corner lifted, which is beyond the crease
// only if the ray really has hit the corner, and facing is positive if it is
// the upper face - the one that was uppermost before it lifted - that the ray
// has come in through. Returns how far along the ray the hit is, negative for
// none.
float hinge(vec3 ro, vec3 rd, vec2 along, vec2 away, float phi, out vec2 q, out float facing) {
	float c = cos(phi);
	float s = sin(phi);
	vec3 upper = vec3(-away * s, c);
	vec3 nose = vec3(halfW, 0.0, 0.0);

	q = vec2(0.0);
	facing = -dot(rd, upper);
	if (abs(facing) < 0.00001) {
		return -1.0; // edge on
	}
	float t = dot(ro - nose, upper) / facing;
	vec3 x = ro + t * rd - nose;
	q = nose.xy + dot(x.xy, along) * along + dot(x, vec3(away * c, s)) * away;
	return t;
}

// inCorner is how far outside the corner a half has in the air a point of the
// sheet is, negative inside it.
float inCorner(vec2 q, int stage) {
	vec2 r = q - vec2(halfW, 0.0);
	if (stage == 1) {
		return max(max(-dot(r, away1), q.x - halfW), q.y - 1.0);
	}
	return max(max(max(-dot(r, away2), dot(r, away1)), q.y - 1.0), -halfW - q.x);
}

// folding is the paper at s while the folds are being made: the stack lying on
// the slide, under whichever corner this half has in the air. shade is the
// shadow those corners throw on the stack.
vec4 folding(vec2 s, float shade) {
	float side = s.y < 0.0 ? -1.0 : 1.0;
	int stage = stageOf(side);
	vec2 g = vec2(s.x, abs(s.y));

	vec4 col = stack(g, side, stage, flatInk);
	col.rgb *= shade;
	if (stage != 1 && stage != 3) {
		return col;
	}

	vec2 along = along1;
	vec2 away = away1;
	float phi = PI * turned1(side);
	if (stage == 3) {
		along = along2;
		away = away2;
		phi = PI * turned2(side);
	}

	vec2 q;
	float facing;
	float t = hinge(vec3(0.0, 0.0, camDist), vec3(g, -camDist), along, away, phi, q, facing);
	if (t <= 0.0 || inCorner(q, stage) > 0.0) {
		return col;
	}

	// The upper face is the printed one, except where the second corner is
	// carrying the first over with it.
	vec3 ink;
	if (facing < 0.0) {
		ink = back(q, side);
	} else if (stage == 3 && q.x >= halfW - 1.0) {
		ink = back(across1(q), side);
	} else {
		ink = front(q, side);
	}

	vec3 n = vec3(-away * sin(phi), cos(phi)) * sign(facing);
	return vec4(ink * lit(n, lightFor(side)), 1.0);
}

// cornerShadow is how much of the light reaching a point on the slide is kept
// off it by the corner one half has in the air.
float cornerShadow(vec2 s, float side) {
	int stage = stageOf(side);
	if (stage != 1 && stage != 3) {
		return 0.0;
	}

	vec2 along = along1;
	vec2 away = away1;
	float phi = PI * turned1(side);
	if (stage == 3) {
		along = along2;
		away = away2;
		phi = PI * turned2(side);
	}

	vec2 q;
	float facing;
	float t = hinge(vec3(s.x, side * s.y, 0.0), lightFor(side), along, away, phi, q, facing);
	if (t <= 0.0) {
		return 0.0;
	}
	float w = shadowSoft + shadowSpread * t;
	return shadowDepth * (1.0 - smoothstep(-w, w, inCorner(q, stage))) / (1.0 + 0.5 * t);
}

// outline is how far outside the paper lying flat a point is, negative inside,
// for a half at the given stage.
float outline(vec2 g, int stage) {
	float d = max(abs(g.x) - halfW, g.y - 1.0);
	vec2 r = g - vec2(halfW, 0.0);
	if (stage >= 3) {
		return max(d, dot(r, away2));
	}
	if (stage >= 1) {
		return max(d, dot(r, away1));
	}
	return d;
}

// toPlane turns a direction in the world into the aeroplane's own frame, where
// it is still lying as it was folded: nose along x, wings out along y.
vec3 toPlane(vec3 v) {
	v.xy = rot(v.xy, -yaw);
	v.xz = rot(v.xz, -pitch);
	v.yz = rot(v.yz, -roll);
	return v;
}

// panel tries a ray against one flat strip of the aeroplane: the paper that lay
// between lo and hi out from the centre line, which now runs from the point o
// in the direction across, o being where the paper at b0 has ended up. The ray
// is already in the half's own coordinates. best and col are the nearest hit so
// far, and are replaced if this one is nearer.
void panel(vec3 ro, vec3 rd, float side, vec3 l, vec3 o, vec3 across, float b0, float lo, float hi, inout float best, inout vec4 col) {
	vec3 upper = vec3(0.0, -across.z, across.y);
	float facing = -dot(rd, upper);
	if (abs(facing) < 0.00001) {
		return;
	}
	float t = dot(ro - o, upper) / facing;
	if (t <= 0.0 || t >= best) {
		return;
	}
	vec3 x = ro + t * rd;
	float b = b0 + dot(x - o, across);
	if (b < lo || b > hi) {
		return;
	}

	vec2 g = vec2(x.x, b);
	if (abs(g.x) > halfW || g.y > 1.0 || dot(g - vec2(halfW, 0.0), away2) > 0.0) {
		return;
	}

	// From above it is the finished dart; from below there is only ever the
	// underside of the sheet itself, as the corners were all folded on top.
	vec3 ink;
	if (facing > 0.0) {
		ink = stack(g, side, 4, front(g, side)).rgb;
	} else {
		ink = back(g, side);
	}

	// The inside of the keel is in its own shade, more so the deeper in and the
	// further it has closed.
	float groove = 1.0;
	if (facing > 0.0) {
		groove -= 0.5 * sin(pinch) * (1.0 - smoothstep(0.0, keel, b));
	}

	best = t;
	col = vec4(ink * lit(upper * sign(facing), l) * groove, 1.0);
}

// flying is the paper at s once the folds are made and it is an aeroplane.
vec4 flying(vec2 s) {
	vec3 ro = toPlane(vec3(0.0, 0.0, camDist) - at);
	vec3 rd = toPlane(vec3(s, -camDist));
	vec3 l = toPlane(vec3(lightDir.x * direction, lightDir.yz));

	// Each half is a strip of keel, hinged along the centre line at the bottom
	// of the body, and a wing hinged to the top of that.
	vec3 body = vec3(0.0, cos(pinch), sin(pinch));
	vec3 wing = vec3(0.0, cos(tips), sin(tips));
	vec3 low = vec3(0.0, 0.0, -keel * sin(pinch));
	vec3 root = vec3(0.0, keel * cos(pinch), 0.0);

	float best = 100000.0;
	vec4 col = vec4(0.0);
	vec3 flip = vec3(1.0, -1.0, 1.0);
	panel(ro, rd, 1.0, l, low, body, 0.0, 0.0, keel, best, col);
	panel(ro, rd, 1.0, l, root, wing, keel, keel, 1.0, best, col);
	panel(ro * flip, rd * flip, -1.0, l * flip, low, body, 0.0, 0.0, keel, best, col);
	panel(ro * flip, rd * flip, -1.0, l * flip, root, wing, keel, keel, 1.0, best, col);
	return col;
}

// planeShadow is how much of the light reaching a point on the slide is kept
// off it by the aeroplane, when that is flying at the given height. This is the
// one place the geometry is not taken at its word: an aeroplane flying away from
// us ought to have gone through the slide, where it could not throw a shadow at
// all. So the shadow is drawn for the aeroplane the eye believes in instead -
// one the size it appears, flat, and climbing off the slide as it goes.
float planeShadow(vec2 s, float height, float fade) {
	float size = camDist / (camDist - at.z);
	vec2 g = s + lightFor(1.0).xy / lightDir.z * height;
	g = rot(g - at.xy * size, -yaw) / size;

	// Pinching the body up has drawn both wings in towards the centre line.
	g.y = abs(g.y) + keel * (1.0 - cos(pinch));

	float w = shadowSoft + shadowSpread * height;
	return fade * shadowDepth * (1.0 - smoothstep(-w, w, outline(g, 4) * size)) / (1.0 + 0.5 * height);
}

void main() {
	vec2 frag = fragCoord();
	p = clamp(progress, 0.0, 1.0);

	// The captures are the whole window with the slide letterboxed inside them,
	// and it is the slide that is the sheet of paper, not the window. Whole
	// pixels, so its edges fall between samples and both ends of the transition
	// are an exact copy of the real slide.
	vec2 slide = vec2(frame.x, frame.x / slideRatio);
	if (frame.x > frame.y * slideRatio) {
		slide = vec2(frame.y * slideRatio, frame.y);
	}
	slide = floor(slide);
	centre = floor((frame - slide) * 0.5) + slide * 0.5;
	unit = slide.y * 0.5;
	halfW = slide.x / slide.y;

	flatInk = texture2D(current, frag / frame).rgb;
	vec3 under = texture2D(next, frag / frame).rgb;

	// The nose points along +x from here on, whichever way we are going.
	vec2 s = (frag - centre) / unit;
	s.x *= direction;

	// Four samples to the pixel, on a rotated grid so that near vertical and
	// near horizontal edges both get four distinct steps.
	vec2 s1 = s + vec2(0.125, 0.375) / unit;
	vec2 s2 = s + vec2(-0.375, 0.125) / unit;
	vec2 s3 = s + vec2(0.375, -0.125) / unit;
	vec2 s4 = s + vec2(-0.125, -0.375) / unit;

	vec4 paper;
	float shadow;
	if (p < pinchAt) {
		float side = s.y < 0.0 ? -1.0 : 1.0;
		float rim = 1.0 - smoothstep(-shadowSoft, shadowSoft, outline(vec2(s.x, abs(s.y)), stageOf(side)));
		float thrown = max(cornerShadow(s, -1.0), cornerShadow(s, 1.0));
		shadow = max(shadowDepth * rim, thrown);

		float shade = 1.0 - thrown;
		paper = 0.25 * (folding(s1, shade) + folding(s2, shade) + folding(s3, shade) + folding(s4, shade));
	} else {
		pinch = keelAngle * smoothstep(pinchAt, pinchAt + pinchTime, p);
		tips = dihedral * smoothstep(pinchAt, pinchAt + pinchTime, p);

		// It gathers speed all the way, off to the side, away from us and up the
		// screen, each a little later than the last so that the path bends. It has
		// left the frame when its tail has, and the frame is both wider than the
		// slide may be and wider the further away the aeroplane is.
		float f = clamp((p - flightAt) / (1.0 - flightAt), 0.0, 1.0);
		float reach = 0.5 * frame.x / unit * (camDist + farOff) / camDist + halfW + 0.5;
		vec3 path = vec3(reach * pow(f, 2.2), -swoop * pow(f, 3.0), -farOff * pow(f, 3.0));
		vec3 heading = vec3(2.2 * reach * pow(f, 1.2) + 0.0001, -3.0 * swoop * f * f, -3.0 * farOff * f * f);

		// The nose follows the path, once it has lifted to take off, and the wings
		// bank into the bend.
		yaw = atan(heading.y, heading.x);
		pitch = pitchUp * sin(PI * sqrt(f)) + atan(heading.z, length(heading.xy));
		roll = bank * sin(PI * smoothstep(0.0, 1.0, f));

		// The slide is a table top, and the paper cannot start out by going
		// through it: the keel closes by standing the wings up on it, not by
		// sinking, and the nose lifts with the tail still resting where it was.
		float lift = keel * sin(pinch) + halfW * sin(max(pitch, 0.0));
		at = path + vec3(0.0, 0.0, lift);

		shadow = planeShadow(s, lift + shadowRise * f, 1.0 - smoothstep(0.0, shadowGone, f));
		paper = 0.25 * (flying(s1) + flying(s2) + flying(s3) + flying(s4));
	}

	// paper is premultiplied by its coverage, and the slide underneath is solid,
	// so the backdrop layer never shows.
	gl_FragColor = vec4(paper.rgb + under * (1.0 - shadow) * (1.0 - paper.a), 1.0);
}
`).lasting(1800 * time.Millisecond)
