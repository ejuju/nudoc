package nudoc

import (
	"bufio"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"strings"
)

const (
	LinePrefixLink            byte = '@'
	LinePrefixListItem        byte = '-'
	LinePrefixPreformatToggle byte = '='
	LinePrefixText            byte = '|'
	LinePrefixTitle           byte = '#'
	LinePrefixTopic           byte = '>'
)

func Parse(r io.Reader) (doc *Document, err error) {
	br := bufio.NewReader(r)
	title, err := br.ReadString('\n')
	isEOF := errors.Is(err, io.EOF)
	if err != nil && !isEOF {
		return nil, fmt.Errorf("read title: %w", err)
	} else if !isEOF {
		title = strings.TrimRight(title, "\r\n")
	}
	title = strings.TrimPrefix(title, "# ")
	doc = &Document{Title: title}

Outer:
	for i := 0; true; i++ {
		line, err := br.ReadString('\n')
		if errors.Is(err, io.EOF) {
			break // Reached EOF.
		} else if err != nil {
			return nil, fmt.Errorf("read line %d: %w", i, err)
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue // Ignore empty line.
		}

		prefix := line[0]
		body := line[1:]
		if len(line) > 1 && line[1] == ' ' {
			body = body[1:]
		}

		switch prefix {
		default:
			continue // Ignore unknown line type.
		case LinePrefixLink:
			url, label, found := strings.Cut(body, " ")
			if !found {
				log.Printf("invalid link: %q", line)
				continue // Ignore invalid link.
			}
			doc.Nodes = append(doc.Nodes, Link{url, label})
		case LinePrefixListItem:
			var list List
			for {
				line, err := br.ReadString('\n')
				if errors.Is(err, io.EOF) {
					break Outer // Reached EOF (tolerated).
				} else if err != nil {
					return nil, fmt.Errorf("read line %d: %w", i, err)
				}
				line = strings.TrimRight(line, "\r\n")
				if line == "" {
					break // Reached end of list.
				}
				prefix := line[0]
				if prefix != LinePrefixListItem {
					log.Printf("invalid list item: %q", line)
				}
				body := line[1:]
				if len(line) > 1 && line[1] == ' ' {
					body = body[1:]
				}
				list = append(list, body)
			}
			doc.Nodes = append(doc.Nodes, list)
		case LinePrefixPreformatToggle:
			alt := body
			content := ""
			for {
				line, err := br.ReadString('\n')
				if errors.Is(err, io.EOF) {
					break Outer // Reached EOF (tolerated).
				} else if err != nil {
					return nil, fmt.Errorf("read line %d: %w", i, err)
				} else if line[0] == LinePrefixPreformatToggle {
					break // Reached end of preformatted block.
				}
				content += line
			}
			doc.Nodes = append(doc.Nodes, PreformattedTextBlock{alt, content})
		case LinePrefixText:
			doc.Nodes = append(doc.Nodes, Text(body))
		case LinePrefixTitle:
			doc.Nodes = append(doc.Nodes, Topic(body))
		case LinePrefixTopic:
			doc.Nodes = append(doc.Nodes, Topic(body))
		}
	}

	return doc, nil
}

//go:embed page.gohtml
var rawTmpl string

var tmpl = template.Must(template.New("").Parse(rawTmpl))

func WriteHTML(w io.Writer, doc *Document) (err error) {
	return tmpl.Execute(w, doc)
}
