package main

// Transition layers move the outgoing and incoming slides. They draw with
// transparency over whichever background layer is beneath, so they only need to
// paint the slides themselves (and any shadow they cast) - everywhere else they
// leave alpha at 0 and the backdrop shows through.

// shuffleTransition is the "galaxy shuffle": the incoming slide swings out from
// behind the current one and lands squarely on top, like a card shuffled from
// the back of the deck to the front.
var shuffleTransition = newShaderLayer("slideShuffle", `
uniform float progress;
uniform float direction;

uniform sampler2D current;
uniform sampler2D next;

// drawCard composites a slide "card" (a texture in its own rotated/scaled box)
// into the premultiplied accumulator, including a soft directional drop shadow
// so the card reads as floating in front of the background layer.
vec4 drawCard(vec4 acc, vec2 frag, vec2 center, vec2 halfSize, float ang, sampler2D tex) {
	// Drop shadow: the card's box cast down and to the right. Darkening the
	// background by a factor of 0.12 at full strength is the same as laying black
	// over it at 0.88 alpha, so the shadow composites to the same result now that
	// the backdrop is a separate layer we cannot read.
	vec2 shOff = vec2(0.012, 0.02) * frame.y;
	vec2 sLocal = rot((frag - shOff) - center, -ang);
	float sDist = length(max(abs(sLocal) - halfSize, 0.0));
	float shadow = exp(-sDist / (0.05 * frame.y));
	acc = over(acc, vec4(0.0, 0.0, 0.0, clamp(shadow, 0.0, 1.0) * 0.6 * 0.88));

	// Card face.
	vec2 local = rot(frag - center, -ang);
	vec2 uv = clamp(local / halfSize * 0.5 + 0.5, 0.0, 1.0);
	vec3 tc = texture2D(tex, uv).rgb;

	// Antialiased coverage of the card rectangle (no derivative funcs needed).
	float e = 1.5;
	float covx = 1.0 - smoothstep(halfSize.x - e, halfSize.x + e, abs(local.x));
	float covy = 1.0 - smoothstep(halfSize.y - e, halfSize.y + e, abs(local.y));
	float cov = covx * covy;
	return over(acc, vec4(tc * cov, cov));
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
	float curTravel = 0.12 * frame.x;
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

	// Z order swaps at the midpoint, where the cards barely overlap so the
	// change of stacking is invisible.
	vec4 col = vec4(0.0);
	if (p < 0.5) {
		col = drawCard(col, frag, inCenter, inHalf, inAng, next);
		col = drawCard(col, frag, curCenter, curHalf, curAng, current);
	} else {
		col = drawCard(col, frag, curCenter, curHalf, curAng, current);
		col = drawCard(col, frag, inCenter, inHalf, inAng, next);
	}

	gl_FragColor = unpremul(col);
}
`)

// slideTransitions lists the movements a deck can be presented with, and
// currentTransition is the one in use.
var slideTransitions = []*shaderLayer{shuffleTransition}

var currentTransition = shuffleTransition
