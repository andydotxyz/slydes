package main

import (
	"bufio"
	"strings"

	"fyne.io/fyne/v2/data/binding"
)

type slides struct {
	count, current binding.Int

	divideRows []int
	items      []string
}

func newSlides() *slides {
	s := &slides{count: binding.NewInt(), current: binding.NewInt(),
		divideRows: make([]int, 0), items: []string{""}}
	_ = s.count.Set(1)
	return s
}

func (s *slides) parseSource(in string) {
	id := 0
	data := ""
	items := make([]string, 0)
	breaks := make([]int, 0)
	scanner := bufio.NewScanner(strings.NewReader(in))
	row := 0
	for scanner.Scan() {
		line := scanner.Text()
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
