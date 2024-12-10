package nudoc

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	LinePrefixLink            byte = '@'
	LinePrefixListItem        byte = '-'
	LinePrefixPreformatToggle byte = '='
	LinePrefixText            byte = '|'
	LinePrefixTopic           byte = '>'
)

type Body struct {
	Nodes []Node
}

const MaxBodyLines = 100_000

var (
	ErrBodyTooManyLines              = errors.New("too many body lines")
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
		} else if len(line) >= 2 && line[1] != ' ' && typ != LinePrefixPreformatToggle {
			return nil, r.WrapErr(ErrBodyMissingSpaceAfterLineType)
		} else if len(line) >= 2 && line[1] == ' ' && typ != LinePrefixPreformatToggle {
			value = value[1:]
		}

		switch typ {
		default:
			continue // Ignore unknown line type.
		case LinePrefixLink:
			// TODO: Extract to ParseLink function and check charset, etc.
			url, label, found := strings.Cut(value, " ")
			if !found {
				return nil, r.WrapErr(ErrInvalidLink)
			}
			body.Nodes = append(body.Nodes, Link{url, label})
		case LinePrefixListItem:
			var list List
			list = append(list, value)
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
				value = value[1:]
				list = append(list, value)
			}
			body.Nodes = append(body.Nodes, list)
		case LinePrefixPreformatToggle:
			alt := value
			content := ""
			for {
				line, typ, _, err = r.ReadBodyLine()
				reachedEOF := errors.Is(err, io.EOF)
				if typ == LinePrefixPreformatToggle {
					break // Reached end of preformatted block.
				} else if reachedEOF && line == "" {
					break // Reached EOF with empty line.
				} else if reachedEOF && line != "" {
					return nil, r.WrapErr(ErrBodyMissingTrailingLF) // Reached EOF with non-empty line.
				} else if err != nil {
					return nil, r.WrapErr(err)
				}
				content += line + "\n"
			}
			body.Nodes = append(body.Nodes, PreformattedTextBlock{alt, content})
		case LinePrefixText:
			body.Nodes = append(body.Nodes, Text(value))
		case LinePrefixTopic:
			body.Nodes = append(body.Nodes, Topic(value))
		}
	}

	return body, nil
}
