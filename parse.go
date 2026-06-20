package main

import (
	"image/color"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/watzon/goshot/pkg/chrome"
	"github.com/watzon/goshot/pkg/content/code"
	"github.com/watzon/goshot/pkg/render"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/storage"
)

type content struct {
	heading, subheading []textSegment
	bgpath              string
	notes               string

	content []fyne.CanvasObject
}

func (s *slides) parseMarkdown(data string) content {
	c := content{}
	if data == "" {
		return c
	}

	r := &parser{c: &c, parent: s}
	md := goldmark.New(goldmark.WithRenderer(r), goldmark.WithExtensions(extension.Strikethrough))
	err := md.Convert([]byte(data), nil)
	if err != nil {
		fyne.LogError("Failed to parse markdown", err)
	}
	return c
}

type parser struct {
	blockquote, heading, list, code bool
	bold, italic, strike            bool
	listDepth                       int
	parent                          *slides

	segments []textSegment // styled runs accumulated for the current bullet
	c        *content
}

func (p *parser) AddOptions(...renderer.Option) {}

func (p *parser) Render(_ io.Writer, source []byte, n ast.Node) error {
	tmpText := ""
	// flush moves the pending text run into a styled segment. Inline text always
	// lives inside a block (paragraph, heading or list item) that emits its
	// accumulated segments at its closing node, so flushing is unconditional.
	flush := func() {
		if tmpText == "" {
			return
		}
		p.segments = append(p.segments, textSegment{text: tmpText, bold: p.bold, italic: p.italic, strike: p.strike})
		tmpText = ""
	}
	err := ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return p.closeElement(n, flush)
		}

		switch n.Kind().String() {
		case "List":
			if p.list { // a nested list: flush the parent item's text first
				flush()
				p.renderBullet()
			}
			p.list = true
			p.listDepth++
		case "ListItem":
			tmpText = ""
		case "Emphasis":
			flush() // start a new styled run
			if em, ok := n.(*ast.Emphasis); ok {
				if em.Level >= 2 {
					p.bold = true
				} else {
					p.italic = true
				}
			}
		case "Strikethrough":
			flush() // start a new styled run
			p.strike = true
		case "Heading":
			p.heading = true
			tmpText = ""
		case "HorizontalRule", "ThematicBreak": // we won't get this as we're splitting slides
		case "Paragraph":
			tmpText = ""
		case "Text":
			if !p.code {
				ret := addTextToSegment(string(n.Text(source)), &tmpText, n)
				if ret != 0 {
					return ret, nil
				}
			}
		case "Blockquote":
			p.blockquote = true
		case "Image":
			p.renderImage(n)
		case "CodeSpan":
			// Inline code becomes a styled segment of its line (bullet, heading or
			// body paragraph) so it renders in place.
			p.code = true
			flush()
			p.segments = append(p.segments, textSegment{text: string(n.Text(source)), code: true})
		case "HTMLBlock":
			lines := n.Lines()
			var sb strings.Builder
			for i := 0; i < lines.Len(); i++ {
				seg := lines.At(i)
				sb.Write(source[seg.Start:seg.Stop])
			}
			if hb, ok := n.(*ast.HTMLBlock); ok && hb.HasClosure() {
				sb.Write(hb.ClosureLine.Value(source))
			}
			p.appendComments(sb.String())
		case "RawHTML":
			rh := n.(*ast.RawHTML)
			var sb strings.Builder
			for i := 0; i < rh.Segments.Len(); i++ {
				seg := rh.Segments.At(i)
				sb.Write(source[seg.Start:seg.Stop])
			}
			p.appendComments(sb.String())
		case "FencedCodeBlock", "CodeBlock":
			p.renderCodeBlock(n, source)
		}

		return ast.WalkContinue, nil
	})
	return err
}

func (p *parser) closeElement(n ast.Node, flush func()) (ast.WalkStatus, error) {
	switch n.Kind().String() {
	case "Heading":
		p.renderHeading(n, flush)
	case "Paragraph":
		// if p.blockquote // TODO
		// In a list the segments belong to the bullet, emitted at ListItem.
		if !p.list {
			flush()
			if len(p.segments) > 0 {
				// color.Black is a placeholder; addContent recolours body
				// lines to the theme foreground.
				p.c.content = append(p.c.content, newRichLine(p.segments, color.Black, false))
				p.segments = nil
			}
		}
	case "ListItem":
		flush()
		p.renderBullet()
	case "Emphasis":
		flush() // close the styled run before clearing the style
		if em, ok := n.(*ast.Emphasis); ok {
			if em.Level >= 2 {
				p.bold = false
			} else {
				p.italic = false
			}
		}
	case "Strikethrough":
		flush() // close the styled run before clearing the style
		p.strike = false
	case "CodeSpan":
		p.code = false
	}
	return ast.WalkContinue, p.handleExitNode(n)
}

// renderBullet builds a bullet from the accumulated segments, if any.
func (p *parser) renderBullet() {
	if len(p.segments) > 0 {
		p.c.content = append(p.c.content, newBullet(p.segments, p.listDepth-1, p.parent.theme))
		p.segments = nil
	}
}

func (p *parser) renderCodeBlock(n ast.Node, source []byte) {
	language := ""
	if c, ok := n.(*ast.FencedCodeBlock); ok {
		language = string(c.Language(source))

		if language != "" {
			lex := lexers.Get(language)
			if lex == nil {
				log.Println("Failed to find lexer for language", language)
				language = ""
			}
		}
	}

	lines := n.Lines()
	raw := ""
	if lines.Len() > 0 {
		raw = string(source[lines.At(0).Start:lines.At(lines.Len()-1).Stop])
	}

	codeContent := code.DefaultRenderer(raw).
		WithTheme("catppuccin-mocha"). // or "-latte" for light
		WithLanguage(language).
		WithLineNumbers(true).
		WithTabWidth(4).
		WithFontSize(42).
		WithMinWidth(600).
		WithMaxWidth(1900)

	draw := render.NewCanvas().
		WithChrome(chrome.NewBlankChrome()).
		WithContent(codeContent)

	img, err := draw.RenderToImage()
	if err != nil {
		fyne.LogError("Failed to render code", err)
	} else {
		rendered := canvas.NewImageFromImage(img)
		rendered.FillMode = canvas.ImageFillContain
		p.c.content = append(p.c.content, rendered)
	}
}

func (p *parser) renderHeading(n ast.Node, flush func()) {
	flush()
	switch n.(*ast.Heading).Level {
	case 1:
		p.c.heading = p.segments
	case 2:
		if len(p.c.subheading) == 0 {
			p.c.subheading = p.segments
		} else {
			t := canvas.NewText(segmentsText(p.segments)+"\r", color.Black)
			t.TextStyle.Bold = true
			p.c.content = append(p.c.content, t)
		}
	default:
		t := canvas.NewText(segmentsText(p.segments)+"\r", color.Black)
		t.TextStyle.Bold = true
		p.c.content = append(p.c.content, t)
	}
	p.segments = nil
	p.heading = false
}

func (p *parser) renderImage(n ast.Node) {
	name := string(n.(*ast.Image).Destination)
	path := filepath.Join(p.root(), name)
	if len(p.c.heading) == 0 {
		p.c.bgpath = path
	} else {
		img := canvas.NewImageFromFile(path)
		img.FillMode = canvas.ImageFillContain
		p.c.content = append(p.c.content, img)
	}
}

// appendComments scans raw HTML for <!-- ... --> comments and appends their
// trimmed contents to the slide's presenter notes, one comment per line.
func (p *parser) appendComments(raw string) {
	s := raw
	for {
		start := strings.Index(s, "<!--")
		if start < 0 {
			return
		}
		rest := s[start+4:]
		end := strings.Index(rest, "-->")
		if end < 0 {
			return
		}
		text := strings.TrimSpace(rest[:end])
		if text != "" {
			if p.c.notes != "" {
				p.c.notes += "\n"
			}
			p.c.notes += text
		}
		s = rest[end+3:]
	}
}

func (p *parser) handleExitNode(n ast.Node) error {
	switch n.Kind().String() {
	case "Blockquote":
		p.blockquote = false
	case "List":
		p.listDepth--
		if p.listDepth == 0 {
			p.list = false
		}
	}
	return nil
}

func addTextToSegment(text string, s *string, node ast.Node) ast.WalkStatus {
	trimmed := strings.ReplaceAll(text, "\n", " ") // newline inside paragraph is not newline
	if trimmed == "" {
		return ast.WalkContinue
	}
	next := node.(*ast.Text).NextSibling()
	if next != nil {
		if nextText, ok := next.(*ast.Text); ok {
			if nextText.Segment.Start > node.(*ast.Text).Segment.Stop { // detect presence of a trailing newline
				trimmed = trimmed + " "
			}
		}
	}

	*s = *s + trimmed
	return 0
}

func (p *parser) root() string {
	if p.parent.uri == nil {
		return ""
	}

	dir, _ := storage.Parent(p.parent.uri)
	return dir.Path()
}
