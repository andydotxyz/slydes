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
	"github.com/yuin/goldmark/renderer"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
)

type content struct {
	heading, subheading string
	bgpath              string

	content []fyne.CanvasObject
}

func (s *slides) parseMarkdown(data string) content {
	c := content{}
	if data == "" {
		return c
	}

	r := &parser{c: &c, parent: s}
	md := goldmark.New(goldmark.WithRenderer(r))
	err := md.Convert([]byte(data), nil)
	if err != nil {
		fyne.LogError("Failed to parse markdown", err)
	}
	return c
}

type parser struct {
	blockquote, heading, list, code bool
	parent                          *slides

	c *content
}

func (p *parser) AddOptions(...renderer.Option) {}

func (p *parser) Render(_ io.Writer, source []byte, n ast.Node) error {
	tmpText := ""
	err := ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			switch n.Kind().String() {
			case "Heading":
				switch n.(*ast.Heading).Level {
				case 1:
					p.c.heading = tmpText
				case 2:
					p.c.subheading = tmpText
				default:
					log.Println("unsupported heading level", n.(*ast.Heading).Level)
				}
			case "Paragraph":
				// if p.blockquote // TODO
				if !p.list && tmpText != "" {
					p.c.content = append(p.c.content, canvas.NewText(tmpText+"\000", color.Black))
				}
			case "ListItem":
				p.c.content = append(p.c.content, newBullet(tmpText, p.parent.theme))
			case "CodeSpan":
				p.code = false
			}
			return ast.WalkContinue, p.handleExitNode(n)
		}

		switch n.Kind().String() {
		case "List":
			p.list = true
		case "ListItem":
			tmpText = ""
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
			name := string(n.(*ast.Image).Destination)
			path := filepath.Join(p.root(), name)
			if p.c.heading == "" {
				p.c.bgpath = path
			} else {
				img := canvas.NewImageFromFile(path)
				img.FillMode = canvas.ImageFillContain
				p.c.content = append(p.c.content, img)
			}
		case "CodeSpan":
			if !p.list && tmpText != "" {
				p.c.content = append(p.c.content, canvas.NewText(tmpText, color.Black))
			}

			p.code = true
			inline := canvas.NewText(string(n.Text(source)), color.Black)
			bg := canvas.NewRectangle(color.Gray{Y: 0xcc})
			p.c.content = append(p.c.content, container.NewStack(bg, inline))
			tmpText = ""
		case "FencedCodeBlock", "CodeBlock":
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
			raw := string(source[lines.At(0).Start:lines.At(lines.Len()-1).Stop])

			codeContent := code.DefaultRenderer(raw).
				WithTheme("catppuccin-mocha"). // or "-latte" for light
				WithLanguage(language).
				WithLineNumbers(true).
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

		return ast.WalkContinue, nil
	})
	return err
}

func (p *parser) handleExitNode(n ast.Node) error {
	switch n.Kind().String() {
	case "Blockquote":
		p.blockquote = false
	case "List":
		p.list = false
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
