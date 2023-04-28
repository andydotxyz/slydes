package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const (
	colorNameBullet           = "bullet"
	colorNameHeader           = "header"
	colorNameSubHeader        = "subHeader"
	colorNameHeaderBackground = "headerBackground"
)

type slideTheme struct {
	fyne.Theme
}

func (s *slideTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameForeground, colorNameBullet:
		return color.Black
	case theme.ColorNameBackground:
		return color.White
	case colorNameHeader:
		return color.White
	case colorNameSubHeader:
		return color.Gray{Y: 0x50}
	case colorNameHeaderBackground:
		return color.Gray{Y: 0xC0}
	default:
		return s.Theme.Color(n, v)
	}
}
