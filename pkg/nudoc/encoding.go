package nudoc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	LinePrefixLink            byte = '@'
	LinePrefixPreformatToggle byte = '='
	LinePrefixText            byte = '|'
	LinePrefixTitle           byte = '#'
	LinePrefixTopic           byte = '>'
)

var linePrefixLabels = map[byte]string{
	LinePrefixLink:            "Link",
	LinePrefixPreformatToggle: "Preformat",
	LinePrefixText:            "Text",
	LinePrefixTitle:           "Title",
	LinePrefixTopic:           "Topic",
}

func LinePrefixLabel(prefix byte) (label string) { return linePrefixLabels[prefix] }

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
			body = line[2:]
		}

		switch prefix {
		default:
			continue // Ignore unknown line type.
		case LinePrefixLink:
			parts := strings.Split(body, " ")
			if len(parts) != 2 {
				continue // Ignore invalid link.
			}
			doc.Nodes = append(doc.Nodes, Link{parts[0], parts[1]})
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
