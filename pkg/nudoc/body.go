package nudoc

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	// 	LinePrefixLink            = "> "
	// 	LinePrefixListTitle       = "| "
	// 	LinePrefixListItem        = "- "
	// 	LinePrefixPreformatToggle = "= "
	// 	LinePrefixTopic           = "> "

	SequenceTopic                  = "# "
	SequenceLink                   = "> "
	SequenceListTitle              = "| "
	SequenceListItem               = "- "
	SequencePreformatLine          = "' "
	SequenceLineComment            = "* "
	SequencePreformatToggle        = "```"
	SequenceMultilineCommentToggle = "***"
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
		line, err := r.ReadLine()
		// TODO: Ensure EOF was reached with an empty line,
		// otherwise the line is silently ignored.
		if errors.Is(err, io.EOF) {
			return body, nil
		} else if err != nil {
			return nil, r.WrapErr(err)
		} else if line == "" {
			continue // Ignore empty line.
		}

		switch {
		default:
			content := line + "\n"
			for {
				line, err = r.ReadLine()
				if errors.Is(err, io.EOF) {
					if line != "" {
						content += line + "\n"
					}
					break
				} else if err != nil {
					return nil, r.WrapErr(fmt.Errorf("on normal line: %w", err))
				} else if line == "" {
					break // Reached empty line.
				} else if strings.HasPrefix(line, SequenceLineComment) {
					continue
				}
				content += line + "\n"
			}
			body.Nodes = append(body.Nodes, Paragraph(content))
		case strings.HasPrefix(line, SequenceLink):
			// TODO: Extract to ParseLink function and check charset, etc.
			if len(line) < 5 {
				// Smallest possible line example "> / A".
				return nil, r.WrapErr(errors.New("link line too short to be valid"))
			}
			url, label, found := strings.Cut(line[2:], " ")
			if !found {
				return nil, r.WrapErr(ErrInvalidLink)
			}
			body.Nodes = append(body.Nodes, Link{url, label})
		case strings.HasPrefix(line, SequenceListTitle):
			if len(line) < len("| A") {
				return nil, r.WrapErr(errors.New("list title line too short to be valid"))
			}
			list := List{Title: line[2:]}
			for {
				line, err = r.ReadLine()
				reachedEOF := errors.Is(err, io.EOF)
				if reachedEOF && line == "" {
					break // Reached EOF with empty line.
				} else if reachedEOF && line != "" {
					return nil, r.WrapErr(ErrBodyMissingTrailingLF) // Reached EOF with non-empty line.
				} else if err != nil {
					return nil, r.WrapErr(err)
				} else if line == "" && !reachedEOF {
					break // Reached end of list.
				} else if !strings.HasPrefix(line, SequenceListItem) {
					return nil, r.WrapErr(fmt.Errorf("not a list item line"))
				} else if len(line) < len("- A") {
					return nil, r.WrapErr(errors.New("list line too short to be valid"))
				}
				list.Items = append(list.Items, line[2:])
			}
			body.Nodes = append(body.Nodes, list)
		case strings.HasPrefix(line, SequencePreformatLine):
			body.Nodes = append(body.Nodes, &PreformattedTextLine{Content: line[2:]})
		case strings.HasPrefix(line, SequencePreformatToggle):
			typ := ""
			typ = strings.TrimSpace(line[3:])
			content := ""
			legend := ""
			for {
				line, err = r.ReadLine()
				reachedEOF := errors.Is(err, io.EOF)
				if strings.HasPrefix(line, SequencePreformatToggle) {
					legend = strings.TrimSpace(line[3:])
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
			body.Nodes = append(body.Nodes, &PreformattedTextBlock{typ, content, legend})
		case strings.HasPrefix(line, SequenceTopic):
			body.Nodes = append(body.Nodes, Topic(line[2:]))
		case strings.HasPrefix(line, SequenceLineComment):
			continue
		case strings.HasPrefix(line, SequenceMultilineCommentToggle):
			for {
				line, err := r.ReadLine()
				if errors.Is(err, io.EOF) {
					return body, nil // Reached EOF.
				} else if err != nil {
					return nil, r.WrapErr(err)
				} else if strings.HasPrefix(line, SequenceMultilineCommentToggle) {
					break
				}
			}
			continue
		}
	}
}
