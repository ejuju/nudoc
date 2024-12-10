package nudoc

import (
	_ "embed"
	"html/template"
	"io"
)

//go:embed page.gohtml
var rawTmpl string

var tmplDoc = template.Must(template.New("").Parse(rawTmpl))

func WriteHTML(w io.Writer, doc *Document) (err error) {
	return tmplDoc.Execute(w, doc)
}
