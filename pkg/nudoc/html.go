package nudoc

import (
	_ "embed"
	"html/template"
	"io"
)

//go:embed page.gohtml
var rawTmpl string

var tmpl = template.Must(template.New("").Parse(rawTmpl))

func WriteHTML(w io.Writer, doc *Document) (err error) {
	return tmpl.Execute(w, doc)
}
