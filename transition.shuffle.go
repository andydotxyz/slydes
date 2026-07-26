package main

// Transition layers move the outgoing and incoming slides. They draw with
// transparency over whichever background layer is beneath, so they only need to
// paint the slides themselves (and any shadow they cast) - everywhere else they
// leave alpha at 0 and the backdrop shows through.

// shuffleTransition is the "galaxy shuffle": the incoming slide swings out from
// behind the current one and lands squarely on top, like a card shuffled from
// the back of the deck to the front. Whichever card is on top peels its trailing
// edge back as it moves, the way a book page lifts, so it reads as a sheet rather
// than a flat cutout - and the peel opens onto whatever is behind it.
var shuffleTransition = newShaderLayer("slideShuffle", "Shuffle", `
uniform float progress;
uniform float direction;

uniform sampler2D current;
uniform sampler2D next;

// uvFor maps a point on the card (centre origin, pixels) to texture coords.
vec2 uvFor(vec2 point, vec2 halfSize) {
	return clamp(point / halfSize * 0.5 + 0.5, 0.0, 1.0);
}

// rectCov is the antialiased coverage of the card's rectangle at a point on it,
// so material rolled off the edge of the page simply stops existing.
float rectCov(vec2 point, vec2 halfSize) {
	float e = 1.5;
	return (1.0 - smoothstep(halfSize.x - e, halfSize.x + e, abs(point.x)))
		* (1.0 - smoothstep(halfSize.y - e, halfSize.y + e, abs(point.y)));
}

// paper is the back of a slide: page white with a little of the printing showing
// through from the other side.
vec3 paper(vec3 ink) {
	return mix(vec3(0.66, 0.65, 0.62), ink, 0.18);
}

// backLit shades the underside of the page ang radians round the roll: brightest
// where it has come all the way over and square on to the viewer, falling away
// towards the roll's silhouette where it turns edge on. The flap is the same
// surface carried past a half turn, so it has to be lit by this same curve at
// ang == PI - lighting it independently leaves a step along the fold, which
// shows up as a faint line drawn across the curl.
float backLit(float ang) {
	return 0.32 - 0.58 * cos(ang);
}

// outsideBox is how far a point lies beyond the card's edge, 0 anywhere on it.
float outsideBox(vec2 point, vec2 halfSize) {
	return length(max(abs(point) - halfSize, 0.0));
}

// creaseShade is how far into shadow the tightest part of the curl goes. The
// page beside the fold and the page climbing out of it both use it, so the
// shadow runs unbroken from one into the other.
const float creaseShade = 0.35;

// curlSlant is how much of the pull is vertical: 1.0 lifts the corner along its
// 45 degree diagonal, lower values tip the fold line upright so the page is
// drawn in from the side rather than from underneath.
const float curlSlant = 0.38;

// drawCard composites a slide "card" into the premultiplied accumulator: a soft
// directional drop shadow, then the page itself with one corner curled back.
//
// The curl is the standard page turn: a fold line runs at right angles to the
// pull direction out towards the lifting corner, and everything past it is
// wrapped onto a cylinder of radius r lying along that fold. Measuring s in the
// pull direction, a
// scrap of page u beyond the fold sits at s = fold + r*sin(u/r) once it is on the
// cylinder, so a pixel at s can show two scraps - one climbing away from the
// viewer at u = r*asin((s-fold)/r) and one coming back over the top at its
// supplement - plus, past a half turn, the flap lying flat back on the page. Each
// is composited in depth order, and past the roll's silhouette nothing covers the
// page at all, so the backdrop shows through where the corner used to be.
//
// curl is 0 for a flat card and 1 for one rolled up past its own centre; corner
// picks which corner lifts, as (+-1, +-1).
vec4 drawCard(vec4 acc, vec2 frag, vec2 center, vec2 halfSize, float ang, vec2 corner, float curl, sampler2D tex) {
	// Drop shadow: the card's box cast down and to the right. Darkening the
	// background by a factor of 0.12 at full strength is the same as laying black
	// over it at 0.88 alpha, so the shadow composites to the same result now that
	// the backdrop is a separate layer we cannot read.
	vec2 shOff = vec2(0.012, 0.02) * frame.y;
	vec2 sLocal = rot((frag - shOff) - center, -ang);
	float sDist = length(max(abs(sLocal) - halfSize, 0.0));
	float shadow = exp(-sDist / (0.05 * frame.y));
	acc = over(acc, vec4(0.0, 0.0, 0.0, clamp(shadow, 0.0, 1.0) * 0.6 * 0.88));

	vec2 local = rot(frag - center, -ang);

	// The fold sweeps in from the chosen corner as curl grows, along a line tipped
	// upright by curlSlant so the page peels in from the side.
	vec2 n = normalize(corner * vec2(1.0, curlSlant));
	float far = dot(corner * halfSize, n);
	float fold = mix(far, -0.2 * far, curl);
	float radius = max(0.22 * (far - fold), 0.25);
	float s = dot(local, n);
	vec2 perp = local - s * n;
	float e = 1.5;

	vec4 card = vec4(0.0); // premultiplied, built up from the page outwards

	if (s < fold) {
		// The flap: the part that has come all the way over and lies back down on
		// the page, its own edge resting on the slide.
		vec2 laid = perp + (fold + PI * radius + (fold - s)) * n;

		// The part of the page still lying flat, in the shade of everything the
		// curl has lifted off it. That has two boundaries and the shadow has to
		// follow both: the fold, along most of the peel where the roll is what is
		// standing up, and the flap's own edge nearer the corner where the page has
		// come back down over itself. Measuring from the flap alone leaves the
		// shadow existing only where the flap does and stopping dead at its edge.
		//
		// It has to reach zero a finite distance out, though: spread far enough
		// and it stops reading as the curl's shadow and becomes a stripe laid
		// across the slide - a band of shading with no edge to explain it.
		float cov = rectCov(local, halfSize);
		if (cov > 0.0) {
			float raised = min(fold - s, outsideBox(laid, halfSize));
			float lit = 1.0 - creaseShade * smoothstep(0.0, 0.04, curl)
				* (1.0 - smoothstep(0.0, 0.55 * radius, raised));
			card = vec4(texture2D(tex, uvFor(local, halfSize)).rgb * lit * cov, cov);
		}

		float cov3 = rectCov(laid, halfSize);
		if (cov3 > 0.0) {
			vec3 tc = paper(texture2D(tex, uvFor(laid, halfSize)).rgb) * backLit(PI);
			card = over(card, vec4(tc * cov3, cov3));
		}
	} else {
		// Past the fold only rolled page can cover this pixel, and only as far as
		// the roll's silhouette.
		float tip = 1.0 - smoothstep(fold + radius - e, fold + radius + e, s);
		float t = clamp((s - fold) / radius, 0.0, 1.0);

		// Climbing away from the viewer: still the printed side, shading off as it
		// turns, and level with the flat page where it leaves the fold.
		//
		// Its own crease is the most boxed in part of the whole curl, so it starts
		// as deep in shadow as the page it is rising out of and opens up as it
		// climbs. Without that the strip of page climbing out of the fold is lit at
		// full brightness right where the page beside it is at its darkest, leaving
		// a lit gap between the shadow and the curl it belongs to.
		float rise = asin(t);
		vec2 out1 = perp + (fold + radius * rise) * n;
		float cov1 = rectCov(out1, halfSize) * tip;
		if (cov1 > 0.0) {
			float crease = 1.0 - creaseShade * (1.0 - smoothstep(0.0, 0.9, rise));
			vec3 tc = texture2D(tex, uvFor(out1, halfSize)).rgb
				* (1.0 - 0.55 * (1.0 - cos(rise))) * crease;
			card = over(card, vec4(tc * cov1, cov1));
		}

		// Coming back over the top: the underside of the page, so page white, and
		// nearer the viewer than the climbing side - it goes on last.
		float back = PI - rise;
		vec2 out2 = perp + (fold + radius * back) * n;
		float cov2 = rectCov(out2, halfSize) * tip;
		if (cov2 > 0.0) {
			vec3 tc = paper(texture2D(tex, uvFor(out2, halfSize)).rgb) * backLit(back);
			card = over(card, vec4(tc * cov2, cov2));
		}
	}

	return over(acc, card);
}

void main() {
	vec2 frag = fragCoord();

	float p = clamp(progress, 0.0, 1.0);
	// A raised cosine bell: 0 at the ends, 1 at the midpoint, and crucially with
	// zero slope at p == 0 and p == 1, so the cards ease into and out of motion
	// instead of lurching the instant the transition starts or stops.
	float bell = 0.5 - 0.5 * cos(2.0 * PI * p);

	// Both cards recede from the viewer mid transition (base 1.0 -> 0.8 -> 1.0)
	// so more of the background shows around them; at p == 0 / p == 1 the active
	// card still exactly fills the frame for a seamless hand off to the real slide.
	float base = mix(1.0, 0.8, bell);
	vec2 center = frame * 0.5;

	// The current slide eases the opposite way to the incoming one - a small
	// drift, tilt and the shared recede - so both feel in motion rather than one
	// card sliding over a static backdrop. It is the top card until the incoming
	// slide swings over it.
	float curTravel = 0.025 * frame.x;
	vec2 curCenter = center + vec2(-direction * bell * curTravel, bell * 0.05 * frame.y);
	float curAng = -direction * bell * 0.05;
	vec2 curHalf = frame * 0.5 * base;

	// The incoming slide starts hidden directly behind the current one, swings
	// out to one side (revealing the background behind), then slides back to land
	// squarely on top. travel must be large enough that at the midpoint (where the
	// z order flips) the incoming card has fully cleared the current card's
	// bounds, otherwise the two overlap at the crossover and the incoming appears
	// to push straight through the current slide instead of swinging around it.
	float travel = 0.82 * frame.x;
	vec2 inCenter = center + vec2(direction * bell * travel, -bell * 0.09 * frame.y);
	float inAng = direction * bell * 0.13;
	vec2 inHalf = frame * 0.5 * base * (1.0 + 0.06 * bell);

	// Only the outgoing slide peels - the slide being dealt in arrives as a clean
	// sheet - and it takes the whole transition to do it: sin(PI*p) opens the peel
	// over the first half and lays it back down over the second, half the rate of a
	// curl that has to open and close within one half, and still flat at both ends
	// where the cards have to match the real slides.
	float curl = 0.5 * sin(PI * p);
	vec2 corner = vec2(direction, 1.0);

	// Z order swaps at the midpoint, where the cards barely overlap so the
	// change of stacking is invisible.
	vec4 col = vec4(0.0);
	if (p < 0.5) {
		col = drawCard(col, frag, inCenter, inHalf, inAng, corner, 0.0, next);
		col = drawCard(col, frag, curCenter, curHalf, curAng, corner, curl, current);
	} else {
		col = drawCard(col, frag, curCenter, curHalf, curAng, corner, curl, current);
		col = drawCard(col, frag, inCenter, inHalf, inAng, corner, 0.0, next);
	}

	gl_FragColor = unpremul(col);
}
`)
