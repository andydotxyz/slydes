package main

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
)

// backgroundImage returns the image inside a slide's backdrop, or nil.
func backgroundImage(s *slide) *canvas.Image {
	c, ok := s.bg.(*fyne.Container)
	if !ok {
		return nil
	}
	for _, o := range c.Objects {
		if img, ok := o.(*canvas.Image); ok {
			return img
		}
	}
	return nil
}

func TestSlideFullPageImage(t *testing.T) {
	test.NewTempApp(t)

	s := newSlide("![](Icon.png)\n", 0, newSlides())
	s.Resize(fyne.NewSize(1600, 900))

	if s.variant != imageSlide {
		t.Errorf("expected an image slide, got variant %d", s.variant)
	}
	img := backgroundImage(s)
	if img == nil {
		t.Fatal("expected a background image")
	}
	if img.Size() != fyne.NewSize(1600, 900) {
		t.Errorf("expected the image to fill the slide, got %v", img.Size())
	}
	if s.footer.Visible() {
		t.Error("expected the footer to be hidden on a full page image")
	}
}

func TestSlideImageBehindContent(t *testing.T) {
	test.NewTempApp(t)

	s := newSlide("![](Icon.png)\n\n# Head\n\n- a\n- b\n", 0, newSlides())
	s.Resize(fyne.NewSize(1600, 900))

	img := backgroundImage(s)
	if img == nil {
		t.Fatal("expected a background image")
	}
	if img.Size() != fyne.NewSize(1600, 900) {
		t.Errorf("expected the image to fill the slide, got %v", img.Size())
	}
	if s.variant == imageSlide {
		t.Error("expected the content to pick the variant, not the image")
	}
	if s.heading == nil {
		t.Error("expected the heading to survive alongside the image")
	}
	bullets := 0
	for _, o := range s.content.Objects {
		if _, ok := o.(*bullet); ok {
			bullets++
		}
	}
	if bullets != 2 {
		t.Errorf("expected 2 bullets over the image, got %d", bullets)
	}
	if !s.footer.Visible() {
		t.Error("expected the footer to show when the image is a backdrop")
	}
}

func TestSlideImageAfterHeadingIsContent(t *testing.T) {
	c := newSlides().parseMarkdown("# Head\n\n![](Icon.png)\n")

	if c.bgpath != "" {
		t.Errorf("expected no background image, got %q", c.bgpath)
	}
	if len(c.content) != 1 {
		t.Fatalf("expected the image as content, got %d objects", len(c.content))
	}
	if _, ok := c.content[0].(*canvas.Image); !ok {
		t.Errorf("expected an image, got %T", c.content[0])
	}
}
