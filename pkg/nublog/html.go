package nublog

import (
	_ "embed"
	"html/template"
	"io"

	"github.com/ejuju/nublog/pkg/nudoc"
)

//go:embed page.gohtml
var rawTmpl string

var tmpl = template.Must(template.New("").Parse(rawTmpl))

func WriteHTML(w io.Writer, doc *nudoc.Document) (err error) {
	return tmpl.Execute(w, doc)
}
