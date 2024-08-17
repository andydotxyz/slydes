package main

import (
	"bytes"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-pdf/fpdf"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
)

var imgID int // each image needs a unique name

func export(s *slides, w io.Writer) error {
	doc := fpdf.NewCustom(&fpdf.InitType{
		Size:    fpdf.SizeType{Wd: 1600, Ht: 900},
		UnitStr: fpdf.UnitPoint,
	})
	pageWidth, totalHeight := doc.GetPageSize()
	pageHeight := pageWidth * (9.0 / 16.0)
	for _, item := range s.items {
		doc.AddPage()

		s := newSlide(item, s)
		s.Resize(fyne.NewSize(float32(pageWidth), float32(pageHeight)))
		err := renderObjectsToPDF(doc, s.content, fyne.Position{Y: float32(totalHeight-pageHeight) / 2})
		if err != nil {
			fyne.LogError("Failed to encode the PDF", err)
		}
	}
	return doc.Output(w)
}

func renderObjectsToPDF(doc *fpdf.Fpdf, o fyne.CanvasObject, off fyne.Position) (err error) {
	switch c := o.(type) {
	case *fyne.Container:
		for _, o := range c.Objects {
			err2 := renderObjectsToPDF(doc, o, off.Add(c.Position()))
			if err == nil && err2 != nil { // propagate first error
				err = err2
			}
		}
	case fyne.Widget:
		for _, o := range test.WidgetRenderer(c).Objects() {
			err2 := renderObjectsToPDF(doc, o, off.Add(c.Position()))
			if err == nil && err2 != nil { // propagate first error
				err = err2
			}
		}
	case *canvas.Rectangle:
		x, y := c.Position().Add(off).Components()
		w, h := c.Size().Components()
		style := ""
		if c.FillColor != nil && c.FillColor != color.Transparent {
			style += "F"
			r, g, b, _ := c.FillColor.RGBA()
			doc.SetFillColor(int(r>>8), int(g>>8), int(b>>8))
		}
		if c.StrokeWidth > 0 && c.StrokeColor != nil && c.StrokeColor != color.Transparent {
			style += "D"
			r, g, b, _ := c.StrokeColor.RGBA()
			doc.SetDrawColor(int(r), int(g), int(b))
			doc.SetLineWidth(float64(c.StrokeWidth))
		}
		doc.Rect(float64(x), float64(y), float64(w), float64(h), style)
	case *canvas.Text:
		r, g, b, _ := c.Color.RGBA()
		doc.SetTextColor(int(r>>8), int(g>>8), int(b>>8))

		x, y := c.Position().Add(off).Components()
		size, base := fyne.CurrentApp().Driver().RenderedTextSize(c.Text, c.TextSize, c.TextStyle, c.FontSource)
		style := ""
		if c.TextStyle.Bold {
			style += "B"
		}
		if c.TextStyle.Italic {
			style += "I"
		}

		if c.TextStyle.Monospace {
			doc.SetFont("Courier", style, float64(c.TextSize))
		} else {
			doc.SetFont("Helvetica", style, float64(c.TextSize))
		}

		w := c.Size().Width
		switch c.Alignment {
		case fyne.TextAlignCenter:
			x += (w - size.Width) / 2
		case fyne.TextAlignTrailing:
			x += w - size.Width
		}
		topPad := (c.Size().Height - size.Height) / 2
		if topPad < 0 { // if size was accidentally too small!
			topPad = 0
		}

		doc.Text(float64(x), float64(y+base+topPad), c.Text)
	case *canvas.Circle:
		x, y := c.Position().Add(off).Components()
		w, h := c.Size().Components()
		style := ""
		if c.FillColor != nil && c.FillColor != color.Transparent {
			style += "F"
			r, g, b, _ := c.FillColor.RGBA()
			doc.SetFillColor(int(r>>8), int(g>>8), int(b>>8))
		}
		if c.StrokeWidth > 0 && c.StrokeColor != nil && c.StrokeColor != color.Transparent {
			style += "D"
			r, g, b, _ := c.StrokeColor.RGBA()
			doc.SetDrawColor(int(r), int(g), int(b))
			doc.SetLineWidth(float64(c.StrokeWidth))
		}
		r := w / 2
		if h < w {
			r = h / 2
		}
		doc.Circle(float64(x+r), float64(y+r), float64(r), style)
	case *canvas.Image:
		ext := ""
		if c.File != "" {
			ext = strings.ToLower(filepath.Ext(c.File))
		} else if c.Resource != nil {
			ext = strings.ToLower(filepath.Ext(c.Resource.Name()))
		}
		imgType := "PNG"
		if ext != "" && (ext == ".jpg" || ext == ".jpeg") {
			imgType = "JPEG"
		}
		size := c.Size()
		x, y := c.Position().Add(off).Components()
		w, h := size.Components()

		if c.FillMode == canvas.ImageFillContain {
			imageAspect := c.Aspect()
			viewAspect := size.Width / size.Height

			if viewAspect > imageAspect {
				w = size.Height * imageAspect
				x += (size.Width - w) / 2
			} else if viewAspect < imageAspect {
				h = size.Width / imageAspect
				y += (size.Height - h) / 2
			}
		}

		imgID++
		name := strconv.Itoa(imgID) + ".png" // a unique name in case any filename collides
		if imgType == "JPEG" {
			name = name[:len(name)-3] + "jpeg"
		}
		var r io.Reader
		b := &bytes.Buffer{}
		if c.Image != nil {
			if imgType == "JPEG" {
				err = jpeg.Encode(b, c.Image, nil)
			} else {
				err = png.Encode(b, c.Image)
			}
			r = bytes.NewReader(b.Bytes())
		} else if c.File != "" {
			r, err = os.Open(c.File)
			defer r.(io.ReadCloser).Close()
		} else if c.Resource != nil {
			r = bytes.NewReader(c.Resource.Content())
		}

		// TODO image fill mode
		opts := fpdf.ImageOptions{ImageType: imgType}
		doc.RegisterImageOptionsReader(name, opts, r)
		doc.ImageOptions(name, float64(x), float64(y), float64(w), float64(h), false, opts, 0, "")
	default:
		log.Println("Missing handler for", c)
	}

	return nil
}
