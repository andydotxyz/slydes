package main

import "testing"

func TestParseNestedBullets(t *testing.T) {
	s := newSlides()
	c := s.parseMarkdown("* level 1\n  * level 2\n")

	if len(c.content) != 2 {
		t.Fatalf("expected 2 bullets, got %d", len(c.content))
	}
	b1, ok := c.content[0].(*bullet)
	if !ok {
		t.Fatalf("expected first item to be *bullet, got %T", c.content[0])
	}
	if b1.content != "level 1" {
		t.Errorf("expected first bullet content %q, got %q", "level 1", b1.content)
	}
	if b1.indent != 0 {
		t.Errorf("expected first bullet indent 0, got %d", b1.indent)
	}
	b2, ok := c.content[1].(*bullet)
	if !ok {
		t.Fatalf("expected second item to be *bullet, got %T", c.content[1])
	}
	if b2.content != "level 2" {
		t.Errorf("expected second bullet content %q, got %q", "level 2", b2.content)
	}
	if b2.indent != 1 {
		t.Errorf("expected second bullet indent 1, got %d", b2.indent)
	}
}

func TestParseBulletStyledSegments(t *testing.T) {
	s := newSlides()
	c := s.parseMarkdown("* foo `bar` **baz** *qux*\n")

	if len(c.content) != 1 {
		t.Fatalf("expected 1 bullet, got %d", len(c.content))
	}
	b, ok := c.content[0].(*bullet)
	if !ok {
		t.Fatalf("expected *bullet, got %T", c.content[0])
	}
	if b.content != "foo bar baz qux" {
		t.Errorf("expected joined content %q, got %q", "foo bar baz qux", b.content)
	}

	want := []textSegment{
		{text: "foo "},
		{text: "bar", code: true},
		{text: " "},
		{text: "baz", bold: true},
		{text: " "},
		{text: "qux", italic: true},
	}
	if len(b.segments) != len(want) {
		t.Fatalf("expected %d segments, got %d: %#v", len(want), len(b.segments), b.segments)
	}
	for i, w := range want {
		if b.segments[i] != w {
			t.Errorf("segment %d: expected %#v, got %#v", i, w, b.segments[i])
		}
	}
}

func TestParseHeadingStyledSegments(t *testing.T) {
	s := newSlides()
	c := s.parseMarkdown("# Using `context` **today**\n")

	// Inline code in a heading must stay in the heading, not leak into the body.
	if len(c.content) != 0 {
		t.Fatalf("expected no body content, got %d objects", len(c.content))
	}
	want := []textSegment{
		{text: "Using "},
		{text: "context", code: true},
		{text: " "},
		{text: "today", bold: true},
	}
	if len(c.heading) != len(want) {
		t.Fatalf("expected %d heading segments, got %d: %#v", len(want), len(c.heading), c.heading)
	}
	for i, w := range want {
		if c.heading[i] != w {
			t.Errorf("segment %d: expected %#v, got %#v", i, w, c.heading[i])
		}
	}
}

func TestParseBodyStyledSegments(t *testing.T) {
	s := newSlides()
	c := s.parseMarkdown("Here is `code` and **bold** text.\n")

	if len(c.content) != 1 {
		t.Fatalf("expected 1 body line, got %d objects", len(c.content))
	}
	line, ok := c.content[0].(*richLine)
	if !ok {
		t.Fatalf("expected body to be *richLine, got %T", c.content[0])
	}
	want := []textSegment{
		{text: "Here is "},
		{text: "code", code: true},
		{text: " and "},
		{text: "bold", bold: true},
		{text: " text."},
	}
	if len(line.segments) != len(want) {
		t.Fatalf("expected %d segments, got %d: %#v", len(want), len(line.segments), line.segments)
	}
	for i, w := range want {
		if line.segments[i] != w {
			t.Errorf("segment %d: expected %#v, got %#v", i, w, line.segments[i])
		}
	}
}

func TestParseStrikethroughSegments(t *testing.T) {
	s := newSlides()

	// Bullet, heading and body paragraph should all carry the strike flag.
	bc := s.parseMarkdown("* keep ~~drop~~\n")
	b, ok := bc.content[0].(*bullet)
	if !ok {
		t.Fatalf("expected *bullet, got %T", bc.content[0])
	}
	wantBullet := []textSegment{
		{text: "keep "},
		{text: "drop", strike: true},
	}
	if len(b.segments) != len(wantBullet) {
		t.Fatalf("bullet: expected %d segments, got %d: %#v", len(wantBullet), len(b.segments), b.segments)
	}
	for i, w := range wantBullet {
		if b.segments[i] != w {
			t.Errorf("bullet segment %d: expected %#v, got %#v", i, w, b.segments[i])
		}
	}

	hc := s.parseMarkdown("# Title ~~old~~\n")
	wantHeading := []textSegment{
		{text: "Title "},
		{text: "old", strike: true},
	}
	if len(hc.heading) != len(wantHeading) {
		t.Fatalf("heading: expected %d segments, got %d: %#v", len(wantHeading), len(hc.heading), hc.heading)
	}
	for i, w := range wantHeading {
		if hc.heading[i] != w {
			t.Errorf("heading segment %d: expected %#v, got %#v", i, w, hc.heading[i])
		}
	}

	pc := s.parseMarkdown("Some ~~removed~~ text.\n")
	line, ok := pc.content[0].(*richLine)
	if !ok {
		t.Fatalf("expected body *richLine, got %T", pc.content[0])
	}
	wantBody := []textSegment{
		{text: "Some "},
		{text: "removed", strike: true},
		{text: " text."},
	}
	if len(line.segments) != len(wantBody) {
		t.Fatalf("body: expected %d segments, got %d: %#v", len(wantBody), len(line.segments), line.segments)
	}
	for i, w := range wantBody {
		if line.segments[i] != w {
			t.Errorf("body segment %d: expected %#v, got %#v", i, w, line.segments[i])
		}
	}
}

func TestParseHTMLCommentsAsNotes(t *testing.T) {
	s := newSlides()

	c := s.parseMarkdown("# Heading\n\n<!-- block note -->\n\nSome body <!-- inline note --> text.\n")
	want := "block note\ninline note"
	if c.notes != want {
		t.Errorf("expected notes %q, got %q", want, c.notes)
	}

	c2 := s.parseMarkdown("<!--   leading and trailing whitespace   -->")
	if c2.notes != "leading and trailing whitespace" {
		t.Errorf("expected trimmed note, got %q", c2.notes)
	}

	c3 := s.parseMarkdown("# Heading\n\njust text\n")
	if c3.notes != "" {
		t.Errorf("expected empty notes, got %q", c3.notes)
	}
}
