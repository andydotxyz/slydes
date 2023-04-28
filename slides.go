package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"

	"github.com/BurntSushi/toml"
)

type slides struct {
	count, current binding.Int
	uri            fyne.URI
	theme          fyne.Theme

	divideRows []int
	items      []string
	config     config
}

func newSlides() *slides {
	s := &slides{count: binding.NewInt(), current: binding.NewInt(),
		divideRows: make([]int, 0), items: []string{""},
		theme: &slideTheme{Theme: theme.DefaultTheme()}}
	_ = s.count.Set(1)
	return s
}

func (s *slides) parseSource(in string) {
	id := 0
	data := ""
	items := make([]string, 0)
	breaks := make([]int, 0)
	s.config = config{}

	scanner := bufio.NewScanner(strings.NewReader(in))
	row := 0
	frontMatter := false
	header := ""
	for scanner.Scan() {
		line := scanner.Text()
		trim := strings.TrimSpace(line)
		if trim == "+++" {
			row++
			if frontMatter {
				frontMatter = false
				s.config = s.parseHeader(strings.TrimSpace(header))
				continue
			} else {
				frontMatter = true
				continue
			}
		}
		if frontMatter {
			row++
			header += trim + "\n"
			continue
		}
		if strings.TrimSpace(line) == "---" {
			items = append(items, data)
			breaks = append(breaks, row)
			id++
			data = ""
		} else {
			data = data + "\n" + line
		}
		row++
	}
	items = append(items, data)

	_ = s.count.Set(len(items))
	id, _ = s.current.Get()
	if id >= len(items) {
		id = len(items) - 1
	}
	s.items = items
	s.divideRows = breaks
}

func (s *slides) currentSource() string {
	id, _ := s.current.Get()
	return s.items[id]
}

func (s *slides) parseHeader(blob string) (c config) {
	_, err := toml.Decode(blob, &c)
	if err != nil { // don't print as it will likely be partial content
		return c
	}
	s.theme = &slideTheme{Theme: theme.DefaultTheme()}

	if c.Theme != "" {
		path := filepath.Join(".", c.Theme+".json")

		f, err := os.Open(path)
		if err == nil {
			th, err := theme.FromJSONReader(f)
			f.Close()
			if err == nil {
				s.theme = th
			}
		}
	}
	return c
}
