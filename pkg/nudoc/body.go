package nudoc

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	LinePrefixLink            byte = '@'
	LinePrefixListTitle       byte = '|'
	LinePrefixListItem        byte = '-'
	LinePrefixPreformatToggle byte = '='
	LinePrefixTopic           byte = '>'
)

type Body struct {
	Nodes []Node
}

const MaxBodyLines = 100_000

var (
	ErrBodyMissingTrailingLF         = errors.New("missing line break at end of file")
	ErrBodyMissingSpaceAfterLineType = errors.New("missing whitespace after line type")
	ErrInvalidLink                   = errors.New("invalid link")
)

func ParseBody(r *Reader) (*Body, error) {
	body := &Body{}
	for {
		line, typ, value, err := r.ReadBodyLine()
		// TODO: Check that EOF was not reached with an empty line,
		// otherwise line is silently ignored.
		if errors.Is(err, io.EOF) {
			break // Reached EOF.
		} else if err != nil {
			return nil, r.WrapErr(err)
		} else if line == "" {
			continue // Ignore empty line.
		}

		switch typ {
		default:
			content := line + "\n"
			for {
				line, err = r.ReadLine()
				if errors.Is(err, io.EOF) && line != "" {
					content += line + "\n"
					break // Reached tolerated EOF with non-empty line.
				} else if err != nil {
					return nil, r.WrapErr(err)
				} else if line == "" {
					break // Reached empty line.
				}
				content += line + "\n"
			}
			body.Nodes = append(body.Nodes, Paragraph(content))
		case LinePrefixLink:
			// TODO: Extract to ParseLink function and check charset, etc.
			url, label, found := strings.Cut(value[1:], " ")
			if !found {
				return nil, r.WrapErr(ErrInvalidLink)
			}
			body.Nodes = append(body.Nodes, Link{url, label})
		case LinePrefixListTitle:
			list := List{Title: value[1:]}
			for {
				line, typ, value, err = r.ReadBodyLine()
				reachedEOF := errors.Is(err, io.EOF)
				if reachedEOF && line == "" {
					break // Reached EOF with empty line.
				} else if reachedEOF && line != "" {
					return nil, r.WrapErr(ErrBodyMissingTrailingLF) // Reached EOF with non-empty line.
				} else if err != nil {
					return nil, r.WrapErr(err)
				} else if line == "" && !reachedEOF {
					break // Reached end of list.
				} else if typ != LinePrefixListItem {
					return nil, r.WrapErr(fmt.Errorf("unexpected list item type %q", typ))
				} else if len(line) >= 2 && line[1] != ' ' {
					return nil, r.WrapErr(ErrBodyMissingSpaceAfterLineType)
				}
				list.Items = append(list.Items, value[1:])
			}
			body.Nodes = append(body.Nodes, list)
		case LinePrefixPreformatToggle:
			contentType := value[1:]
			content := ""
			alt := ""
			for {
				line, typ, _, err = r.ReadBodyLine()
				reachedEOF := errors.Is(err, io.EOF)
				if typ == LinePrefixPreformatToggle {
					if len(line) > 2 {
						alt = line[2:]
					}
					break // Reached end of preformatted block.
				} else if reachedEOF && line == "" {
					break // Reached EOF (with empty line).
				} else if reachedEOF && line != "" {
					return nil, r.WrapErr(ErrBodyMissingTrailingLF) // Reached EOF with non-empty line.
				} else if err != nil {
					return nil, r.WrapErr(err)
				}
				content += line + "\n"
			}
			body.Nodes = append(body.Nodes, &PreformattedTextBlock{contentType, content, alt})
		case LinePrefixTopic:
			body.Nodes = append(body.Nodes, Topic(value[1:]))
		}
	}

	return body, nil
}
