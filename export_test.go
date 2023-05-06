package main

import (
	"bytes"
	"os"
	"testing"
)

func TestExport(t *testing.T) {
	s := newSlides()
	s.parseSource(`# Slide 1
## Subhead

---
# Slide 2

Content!

* Bullet 1
* Bullet 2
* Bullet 3

---

# Icon

left side

![](Icon.png)
`)
	buf := &bytes.Buffer{}
	err := export(s, buf)
	if err != nil {
		t.Error(err)
		return
	}

	golden, _ := os.ReadFile("testdata/export.pdf")
	data := buf.Bytes()
	if len(golden) != len(data) {
		t.Error("Wrong number of bytes in output")
	}

	end := len(golden) - 1024 // end of PDF not consistent on re-export
	for i, b := range golden {
		if i > end {
			break
		}
		if data[i] != b {
			t.Error("Wrong data at index", i, b, "!=", data[i])
		}
	}
}
