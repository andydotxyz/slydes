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
