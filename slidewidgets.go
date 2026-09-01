package main

import (
	"image/color"
	"strconv"

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

// dotSize is the diameter of an unordered bullet's dot, before scaling. A
// numbered bullet reserves the same height so both list styles line up.
const dotSize = float32(5)

type bullet struct {
	widget.BaseWidget
	theme fyne.Theme

	content  string // plain-text join of the segments, used for the empty check
	segments []textSegment
	indent   int
	numbered bool // an ordered list item, marked with its number instead of a dot
	number   int
	scale    float32

	dot   *canvas.Circle      // the marker of an unordered bullet
	label *canvas.Text        // the marker of a numbered bullet
	texts []*canvas.Text      // one per segment, index-aligned with segments
	bgs   []*canvas.Rectangle // code-span backgrounds; nil entry for non-code segments
}

func newBullet(segments []textSegment, indent int, th fyne.Theme) *bullet {
	return &bullet{theme: th, segments: segments, content: segmentsText(segments), indent: indent, scale: 1}
}

// newNumberedBullet builds the bullet for an item of an ordered list, marked
// with the given number rather than a dot.
func newNumberedBullet(segments []textSegment, indent, number int, th fyne.Theme) *bullet {
	b := newBullet(segments, indent, th)
	b.numbered = true
	b.number = number
	return b
}

// markerColor is the colour of the dot or number. An empty item has no marker.
func (b *bullet) markerColor() color.Color {
	if b.content == "" {
		return color.Transparent
	}
	return b.theme.Color(colorNameBullet, theme.VariantLight)
}

// markerSize is the space the marker occupies. A numbered list lays out its
// lines with the same spacing as an unordered one.
func (b *bullet) markerSize() fyne.Size {
	if b.numbered {
		width := float32(0)
		if b.label != nil {
			width = b.label.MinSize().Width
		}
		return fyne.NewSize(width, dotSize*b.scale)
	}
	return b.dot.Size()
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
	var objs []fyne.CanvasObject
	if b.numbered {
		// The number matches the body text size so it reads as part of the list.
		b.label = canvas.NewText(strconv.Itoa(b.number)+".", b.markerColor())
		b.label.TextSize = theme.TextSize() * b.scale
		objs = []fyne.CanvasObject{b.label}
	} else {
		b.dot = canvas.NewCircle(b.markerColor())
		b.dot.Resize(fyne.NewSize(dotSize*b.scale, dotSize*b.scale))
		objs = []fyne.CanvasObject{b.dot}
	}
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
		b.dot.FillColor = b.markerColor()
		b.dot.Refresh()
	}
	if b.label != nil {
		b.label.Color = b.markerColor()
		b.label.Refresh()
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
	marker := b.markerSize()
	if b.numbered {
		// Sized like the segment texts below so the exporter centres it the same way.
		b.label.Move(fyne.NewPos(off, 0))
		b.label.Resize(fyne.NewSize(marker.Width, size.Height))
	} else {
		b.dot.Move(fyne.NewPos(off, (size.Height-marker.Height)/2))
	}

	x := off + marker.Width + theme.Padding()*b.scale
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
	return b.markerSize().Add(textMin).AddWidthHeight(theme.Padding()*b.scale+b.indentOffset(), 0)
}

func (b *bullet) setScale(scale float32) {
	_ = test.WidgetRenderer(b)
	b.scale = scale

	if b.numbered {
		b.label.TextSize = theme.TextSize() * scale
	} else {
		b.dot.Resize(fyne.NewSize(dotSize*scale, dotSize*scale))
	}
	for _, t := range b.texts {
		t.TextSize = theme.TextSize() * scale
	}
}

type separator struct {
	widget.BaseWidget
	theme fyne.Theme

	scale float32

	line *canvas.Rectangle
}

func newSeparator(th fyne.Theme) *separator {
	return &separator{theme: th, scale: 1}
}

func (s *separator) CreateRenderer() fyne.WidgetRenderer {
	s.line = canvas.NewRectangle(s.theme.Color(theme.ColorNameSeparator, theme.VariantLight))

	objs := []fyne.CanvasObject{s.line}
	return widget.NewSimpleRenderer(container.NewWithoutLayout(objs...))
}

func (s *separator) Refresh() {
	if s.line != nil {
		s.line.FillColor = s.theme.Color(theme.ColorNameSeparator, theme.VariantLight)
		s.line.Refresh()
	}
}

func (s *separator) Resize(size fyne.Size) {
	h := s.theme.Size(theme.SizeNameSeparatorThickness) * s.scale
	pad := s.theme.Size(theme.SizeNamePadding) * s.scale
	s.line.Move(fyne.NewPos(0, pad))
	s.line.Resize(fyne.NewSize(size.Width, h))
}

func (s *separator) MinSize() fyne.Size {
	h := s.theme.Size(theme.SizeNameSeparatorThickness) * s.scale
	pad := s.theme.Size(theme.SizeNamePadding) * s.scale
	return fyne.NewSize(h, h+2*pad)
}

func (s *separator) setScale(scale float32) {
	_ = test.WidgetRenderer(s)
	s.scale = scale

	h := s.theme.Size(theme.SizeNameSeparatorThickness) * s.scale
	s.line.Resize(fyne.NewSize(s.line.Size().Width, h))
}

// segmentsText joins the plain text of every segment.
func segmentsText(segments []textSegment) string {
	s := ""
	for _, seg := range segments {
		s += seg.text
	}
	return s
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
