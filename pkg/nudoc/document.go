package nudoc

import (
	"fmt"
	"html"
	"html/template"
)

type Document struct {
	Header *Header
	Body   *Body
}

func ParseDocument(r *Reader) (*Document, error) {
	header, err := ParseHeader(r)
	if err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}
	body, err := ParseBody(r)
	if err != nil {
		return nil, fmt.Errorf("parse body: %w", err)
	}
	return &Document{header, body}, nil
}

type Node interface {
	NuDoc() string
	Markdown() string
	HTML() template.HTML
}

type Link struct {
	URL   string
	Label string
}

func (n Link) NuDoc() string    { return string(SequenceLink) + " " + n.URL + " " + n.Label + "\n" }
func (n Link) Markdown() string { return "- [" + n.URL + "](" + n.Label + ")\n" }
func (n Link) HTML() template.HTML {
	return template.HTML("<a href=\"" + n.URL + "\">" + n.Label + "</a>")
}

type List struct {
	Title string
	Items []string
}

func (n List) NuDoc() (v string) {
	v += string(SequenceListTitle) + " " + n.Title + "\n"
	for _, item := range n.Items {
		v += string(SequenceListItem) + " " + item + "\n"
	}
	v += "\n"
	return v
}

func (n List) Markdown() (v string) {
	v += n.Title + "\n"
	for _, item := range n.Items {
		v += "- " + item + "\n"
	}
	v += "\n"
	return v
}

func (n List) HTML() template.HTML {
	v := "<div>"
	v += "<p>" + n.Title + "</p>\n"
	v += "<ul>\n"
	for _, item := range n.Items {
		v += "<li>" + item + "</li>\n"
	}
	v += "</ul>\n"
	v += "</div>\n"
	return template.HTML(v)
}

type PreformattedTextBlock struct {
	Type    string // Content type for client-side content hilighting (and a11y).
	Content string // Actual preformatted content.
	Legend  string // For screen-readers.
}

func (n PreformattedTextBlock) NuDoc() string {
	return string(SequencePreformatToggle) + " " + n.Legend + "\n" +
		n.Content + "\n" +
		string(SequencePreformatToggle) + "\n"
}

func (n PreformattedTextBlock) Markdown() string { return "```\n" + n.Content + "\n```\n" }

func (n PreformattedTextBlock) HTML() template.HTML {
	return template.HTML("<div class=\"pre-block\">\n" +
		"<pre aria-label=\"" + html.EscapeString(n.Legend) + "\">\n" + n.Content + "</pre>\n" +
		"<div class=\"meta\"><legend>" + n.Legend + "</legend><button>Copy</button></div>\n" +
		"</div>")
}

type PreformattedTextLine struct {
	Content string
}

func (n PreformattedTextLine) NuDoc() string {
	return string(SequencePreformatLine) + n.Content + "\n"
}

func (n PreformattedTextLine) Markdown() string { return "```\n" + n.Content + "\n```\n" }

func (n PreformattedTextLine) HTML() template.HTML {
	return template.HTML("<div class=\"pre-line\">\n" +
		"<pre>\n" + n.Content + "</pre>\n" +
		"</div>")
}

type Paragraph string

func (n Paragraph) NuDoc() string    { return string(n) + "\n" }
func (n Paragraph) Markdown() string { return string(n) + "\n" }
func (n Paragraph) HTML() template.HTML {
	return template.HTML("<p>" + html.EscapeString(string(n)) + "</p>")
}

type Topic string

func (n Topic) NuDoc() string       { return string(SequenceTopic) + " " + string(n) + "\n" }
func (n Topic) Markdown() string    { return "## " + string(n) + "\n" }
func (n Topic) HTML() template.HTML { return template.HTML("<h2>" + string(n) + "</h2>\n") }
