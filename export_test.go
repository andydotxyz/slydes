package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"testing"
)

func TestExport(t *testing.T) {
	s := newSlides()
	s.parseSource(`+++
name = "Presenter"
footer = "Slydes"
+++

# Slide 1
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
	golden = normalisePDF(golden)
	data := normalisePDF(buf.Bytes())
	if len(golden) != len(data) {
		t.Error("Wrong number of bytes in output")
	}

	end := len(golden) - 1024 // end of PDF not consistent on re-export
	for i, b := range golden {
		if i > end || i >= len(data) {
			break
		}
		if data[i] != b {
			t.Error("export missmatch starting at index", i, b, "!=", data[i])
			t.Fail()
			return
		}
	}
}

var (
	pdfStream    = regexp.MustCompile(`/Length (\d+)>>\nstream\n`)
	pdfImageName = regexp.MustCompile(`/I[0-9a-f]{40}`)
)

// normalisePDF inflates the compressed streams of a PDF and renames its images,
// which are keyed by a hash of their compressed data. The zlib output differs
// between Go releases, so the raw bytes cannot be compared with a golden file.
func normalisePDF(data []byte) []byte {
	out := &bytes.Buffer{}
	pos := 0
	for {
		loc := pdfStream.FindSubmatchIndex(data[pos:])
		if loc == nil {
			break
		}
		start, body := pos+loc[0], pos+loc[1]
		length, _ := strconv.Atoi(string(data[pos+loc[2] : pos+loc[3]]))
		if body+length > len(data) {
			break
		}

		stream := data[body : body+length]
		dict := data[pos:start]
		if obj := bytes.LastIndex(dict, []byte(" obj\n")); obj >= 0 {
			dict = dict[obj:]
		}
		if bytes.Contains(dict, []byte("/FlateDecode")) {
			if r, err := zlib.NewReader(bytes.NewReader(stream)); err == nil {
				if inflated, err := io.ReadAll(r); err == nil {
					stream = inflated
				}
			}
		}

		out.Write(data[pos:start])
		fmt.Fprintf(out, "/Length %d>>\nstream\n", len(stream))
		out.Write(stream)
		pos = body + length
	}
	out.Write(data[pos:])

	names := map[string]string{}
	return pdfImageName.ReplaceAllFunc(out.Bytes(), func(name []byte) []byte {
		id, ok := names[string(name)]
		if !ok {
			id = "/I" + strconv.Itoa(len(names))
			names[string(name)] = id
		}
		return []byte(id)
	})
}
