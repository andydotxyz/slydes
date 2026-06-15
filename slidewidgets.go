package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type slideWidget interface {
	setScale(float32)
}

// textSegment is one styled run of text within a bullet, e.g. a bold word or
// an inline code span. Plain text has all flags false.
type textSegment struct {
	text   string
	bold   bool
	italic bool
	code   bool
	strike bool
}

type bullet struct {
	widget.BaseWidget
	theme fyne.Theme

	content  string // plain-text join of the segments, used for the empty check
	segments []textSegment
	indent   int
	scale    float32

	dot   *canvas.Circle
	texts []*canvas.Text      // one per segment, index-aligned with segments
	bgs   []*canvas.Rectangle // code-span backgrounds; nil entry for non-code segments
}

// segmentsText joins the plain text of every segment.
func segmentsText(segments []textSegment) string {
	s := ""
	for _, seg := range segments {
		s += seg.text
	}
	return s
}

func newBullet(segments []textSegment, indent int, th fyne.Theme) *bullet {
	return &bullet{theme: th, segments: segments, content: segmentsText(segments), indent: indent, scale: 1}
}

// segmentColor picks the text colour for a segment: inline code stays black to
// read on its grey background, everything else uses the bullet colour.
func (b *bullet) segmentColor(seg textSegment) color.Color {
	if seg.code {
		return color.Black
	}
	return b.theme.Color(colorNameBullet, theme.VariantLight)
}

func (b *bullet) CreateRenderer() fyne.WidgetRenderer {
	b.dot = canvas.NewCircle(b.theme.Color(colorNameBullet, theme.VariantLight))
	if b.content == "" {
		b.dot.FillColor = color.Transparent
	}

	objs := []fyne.CanvasObject{b.dot}
	b.texts = make([]*canvas.Text, len(b.segments))
	b.bgs = make([]*canvas.Rectangle, len(b.segments))
	for i, seg := range b.segments {
		t := canvas.NewText(seg.text, b.segmentColor(seg))
		t.TextStyle = fyne.TextStyle{Bold: seg.bold, Italic: seg.italic, Monospace: seg.code, Strikethrough: seg.strike}
		b.texts[i] = t
		if seg.code {
			bg := canvas.NewRectangle(color.Gray{Y: 0xcc})
			b.bgs[i] = bg
			objs = append(objs, bg) // behind its text
		}
		objs = append(objs, t)
	}
	return widget.NewSimpleRenderer(container.NewWithoutLayout(objs...))
}

func (b *bullet) Refresh() {
	if b.dot != nil {
		if b.content == "" {
			b.dot.FillColor = color.Transparent
		} else {
			b.dot.FillColor = b.theme.Color(colorNameBullet, theme.VariantLight)
		}
		b.dot.Refresh()
	}
	for i, t := range b.texts {
		t.Color = b.segmentColor(b.segments[i])
		t.Refresh()
	}
}

func (b *bullet) indentOffset() float32 {
	return float32(b.indent) * theme.Padding() * 4 * b.scale
}

func (b *bullet) Resize(size fyne.Size) {
	off := b.indentOffset()
	b.dot.Move(fyne.NewPos(off, (size.Height-b.dot.Size().Height)/2))

	x := off + b.dot.Size().Width + theme.Padding()*b.scale
	for i, t := range b.texts {
		min := t.MinSize()
		if bg := b.bgs[i]; bg != nil {
			bg.Move(fyne.NewPos(x, 0))
			bg.Resize(min)
		}
		// Height matches the bullet so the PDF exporter centres each run the
		// same way the single-text bullet used to; width is the run's own.
		t.Move(fyne.NewPos(x, 0))
		t.Resize(fyne.NewSize(min.Width, size.Height))
		x += min.Width
	}
}

func (b *bullet) MinSize() fyne.Size {
	if len(b.texts) == 0 {
		return fyne.NewSize(14, 4)
	}

	width := float32(0)
	height := float32(0)
	for _, t := range b.texts {
		min := t.MinSize()
		width += min.Width
		if min.Height > height {
			height = min.Height
		}
	}
	textMin := fyne.NewSize(width, height)
	return b.dot.Size().Add(textMin).AddWidthHeight(theme.Padding()*b.scale+b.indentOffset(), 0)
}

func (b *bullet) setScale(scale float32) {
	_ = test.WidgetRenderer(b)
	b.scale = scale

	b.dot.Resize(fyne.NewSize(5*scale, 5*scale))
	for _, t := range b.texts {
		t.TextSize = theme.TextSize() * scale
	}
}

// richLine renders a single line of styled text segments, used for slide
// headings and subheadings. Like bullet it draws one canvas.Text per segment
// (code spans get a grey background) but it has no dot and instead applies a
// whole-line colour, base style and horizontal alignment.
type richLine struct {
	widget.BaseWidget

	segments  []textSegment
	color     color.Color
	baseBold  bool
	textSize  float32
	alignment fyne.TextAlign

	texts []*canvas.Text
	bgs   []*canvas.Rectangle // code-span backgrounds; nil entry for non-code segments
}

func newRichLine(segments []textSegment, col color.Color, baseBold bool) *richLine {
	return &richLine{segments: segments, color: col, baseBold: baseBold, textSize: theme.TextSize()}
}

func (r *richLine) CreateRenderer() fyne.WidgetRenderer {
	objs := []fyne.CanvasObject{}
	r.texts = make([]*canvas.Text, len(r.segments))
	r.bgs = make([]*canvas.Rectangle, len(r.segments))
	for i, seg := range r.segments {
		t := canvas.NewText(seg.text, r.color)
		t.TextSize = r.textSize
		t.TextStyle = fyne.TextStyle{Bold: r.baseBold || seg.bold, Italic: seg.italic, Monospace: seg.code, Strikethrough: seg.strike}
		r.texts[i] = t
		if seg.code {
			bg := canvas.NewRectangle(color.Gray{Y: 0xcc})
			r.bgs[i] = bg
			objs = append(objs, bg) // behind its text
		}
		objs = append(objs, t)
	}
	return widget.NewSimpleRenderer(container.NewWithoutLayout(objs...))
}

func (r *richLine) Refresh() {
	for _, t := range r.texts {
		t.Color = r.color
		t.Refresh()
	}
	for _, bg := range r.bgs {
		if bg != nil {
			bg.Refresh()
		}
	}
}

func (r *richLine) MinSize() fyne.Size {
	width := float32(0)
	height := float32(0)
	for _, t := range r.texts {
		min := t.MinSize()
		width += min.Width
		if min.Height > height {
			height = min.Height
		}
	}
	return fyne.NewSize(width, height)
}

func (r *richLine) Resize(size fyne.Size) {
	total := r.MinSize().Width

	x := float32(0)
	switch r.alignment {
	case fyne.TextAlignCenter:
		x = (size.Width - total) / 2
	case fyne.TextAlignTrailing:
		x = size.Width - total
	}
	if x < 0 {
		x = 0
	}

	for i, t := range r.texts {
		min := t.MinSize()
		if bg := r.bgs[i]; bg != nil {
			bg.Move(fyne.NewPos(x, 0))
			bg.Resize(min)
		}
		t.Move(fyne.NewPos(x, 0))
		t.Resize(min)
		x += min.Width
	}
}

func (r *richLine) setTextSize(size float32) {
	_ = test.WidgetRenderer(r)
	r.textSize = size
	for _, t := range r.texts {
		t.TextSize = size
	}
}

// setScale lets a richLine be used as body content, where layoutContent scales
// it to the body text size. Headings instead set their size explicitly and are
// laid out separately, so setScale is never called on them.
func (r *richLine) setScale(scale float32) {
	r.setTextSize(theme.TextSize() * scale)
}

func (r *richLine) setColor(col color.Color) {
	_ = test.WidgetRenderer(r)
	r.color = col
	for _, t := range r.texts {
		t.Color = col
	}
}
