package main

import (
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func colorTexts(list []widget.RichTextSegment) {
	for _, s := range list {
		switch seg := s.(type) {
		case *widget.TextSegment:
			seg.Style.ColorName = theme.ColorNameBackground
		case *widget.ListSegment:
			colorTexts(seg.Items)
		case *widget.ParagraphSegment:
			colorTexts(seg.Texts)
		}
	}
}
