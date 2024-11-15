package nudoc

import "html/template"

type Document struct {
	Title string
	Nodes []Node
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

func (n Link) NuDoc() string    { return string(LinePrefixLink) + " " + n.URL + " " + n.Label + "\n" }
func (n Link) Markdown() string { return "- [" + n.URL + "](" + n.Label + ")\n" }
func (n Link) HTML() template.HTML {
	return template.HTML("<a href=\"" + n.URL + "\">" + n.Label + "</a>")
}

type List []string

func (n List) NuDoc() (v string) {
	for _, item := range n {
		v += string(LinePrefixListItem) + " " + item + "\n"
	}
	v += "\n"
	return v
}

func (n List) Markdown() (v string) {
	for _, item := range n {
		v += "- " + item + "\n"
	}
	return v
}

func (n List) HTML() template.HTML {
	v := "<ul>\n"
	for _, item := range n {
		v += "<li>" + item + "</li>\n"
	}
	v += "</ul>\n"
	return template.HTML(v)
}

type PreformattedTextBlock struct {
	Alt string // For a11y.
	Pre string // Actual preformatted content.
}

func (n PreformattedTextBlock) NuDoc() string {
	return string(LinePrefixPreformatToggle) + " " + n.Alt + "\n" +
		n.Pre + "\n" +
		string(LinePrefixPreformatToggle) + "\n"
}

func (n PreformattedTextBlock) Markdown() string { return "```\n" + n.Pre + "\n```\n" }

func (n PreformattedTextBlock) HTML() template.HTML {
	return template.HTML("<div class=\"pre-block\">\n" +
		"<pre aria-label=\"" + n.Alt + "\">\n" + n.Pre + "</pre>\n" +
		"<div class=\"meta\"><legend>" + n.Alt + "</legend><button>Copy</button></div>\n" +
		"</div>")
}

type Text string

func (n Text) NuDoc() string       { return string(LinePrefixText) + " " + string(n) + "\n" }
func (n Text) Markdown() string    { return string(n) + "\n" }
func (n Text) HTML() template.HTML { return template.HTML("<p>" + string(n) + "</p>") }

type Title string

func (n Title) NuDoc() string       { return string(LinePrefixTitle) + " " + string(n) + "\n" }
func (n Title) Markdown() string    { return "# " + string(n) + "\n" }
func (n Title) HTML() template.HTML { return template.HTML("<h1>" + string(n) + "</h1>") }

type Topic string

func (n Topic) NuDoc() string       { return string(LinePrefixTopic) + " " + string(n) + "\n" }
func (n Topic) Markdown() string    { return "## " + string(n) + "\n" }
func (n Topic) HTML() template.HTML { return template.HTML("<h2>" + string(n) + "</h2>\n") }
