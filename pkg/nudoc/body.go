package nudoc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
)

type Body struct {
	Nodes []Node
}

const (
	LinePrefixLink            byte = '@'
	LinePrefixListItem        byte = '-'
	LinePrefixPreformatToggle byte = '='
	LinePrefixText            byte = '|'
	LinePrefixTitle           byte = '#'
	LinePrefixTopic           byte = '>'
)

const MaxBodyLines = 100_000

func ParseBody(r *bufio.Reader) (*Body, error) {
	body := &Body{}
	for i := 0; i <= MaxBodyLines; i++ {
		if i == MaxBodyLines {
			return nil, fmt.Errorf("too many body lines (#%d)", i)
		}
		line, err := r.ReadString('\n')
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
		value := line[1:]
		if len(line) > 1 && line[1] == ' ' {
			value = value[1:]
		}

		switch prefix {
		default:
			continue // Ignore unknown line type.
		case LinePrefixLink:
			url, label, found := strings.Cut(value, " ")
			if !found {
				log.Printf("invalid link: %q", line)
				continue // Ignore invalid link.
			}
			body.Nodes = append(body.Nodes, Link{url, label})
		case LinePrefixListItem:
			var list List
			for {
				line, err := r.ReadString('\n')
				if errors.Is(err, io.EOF) {
					return nil, fmt.Errorf("read line %d: missing LF after list item: %w", i, err)
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
			body.Nodes = append(body.Nodes, list)
		case LinePrefixPreformatToggle:
			alt := value
			content := ""
			for {
				line, err := r.ReadString('\n')
				if errors.Is(err, io.EOF) {
					return nil, fmt.Errorf("read line %d: missing LF after preformatted block: %w", i, err)
				} else if err != nil {
					return nil, fmt.Errorf("read line %d: %w", i, err)
				} else if line[0] == LinePrefixPreformatToggle {
					break // Reached end of preformatted block.
				}
				content += line
			}
			body.Nodes = append(body.Nodes, PreformattedTextBlock{alt, content})
		case LinePrefixText:
			body.Nodes = append(body.Nodes, Text(value))
		case LinePrefixTitle:
			body.Nodes = append(body.Nodes, Topic(value))
		case LinePrefixTopic:
			body.Nodes = append(body.Nodes, Topic(value))
		}
	}

	return body, nil
}
