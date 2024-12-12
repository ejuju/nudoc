package nudoc

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	SequenceTopic                  = "# "
	SequenceLink                   = "> "
	SequenceListTitle              = "| "
	SequenceListItem               = "- "
	SequencePreformatLine          = "' "
	SequenceLineComment            = "* "
	SequenceAlternative            = "~ "
	SequencePreformatToggle        = "``` "
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
			node, err := parseLink(line)
			if err != nil {
				return nil, r.WrapErr(err)
			}
			body.Nodes = append(body.Nodes, node)
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
				} else if strings.HasPrefix(line, "````") {
					line = line[1:] // Unescape "````" to "```".
				}
				content += line + "\n"
			}
			body.Nodes = append(body.Nodes, &PreformattedTextBlock{typ, content, legend})
		case strings.HasPrefix(line, SequenceTopic):
			node, err := parseTopic(line)
			if err != nil {
				return nil, r.WrapErr(err)
			}
			body.Nodes = append(body.Nodes, node)
		case strings.HasPrefix(line, SequenceAlternative):
			if len(line) < len("~ A") {
				return nil, r.WrapErr(errors.New("alternative line too short to be valid"))
			}
			content := line[2:]
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
				} else if !strings.HasPrefix(line, SequenceAlternative) {
					return nil, r.WrapErr(fmt.Errorf("not an alternative list item"))
				} else if len(line) < len("~ A") {
					return nil, r.WrapErr(errors.New("alternative line too short to be valid"))
				}
				content += line[2:]
			}
			body.Nodes = append(body.Nodes, Alternative(content))
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

func parseTopic(line string) (Topic, error) {
	if !strings.HasPrefix(line, SequenceTopic) {
		return "", errors.New("invalid topic start sequence")
	} else if len(line) < len(SequenceTopic)+1 {
		return "", errors.New("missing topic value")
	}
	return Topic(line[2:]), nil
}

func parseLink(line string) (*Link, error) {
	if !strings.HasPrefix(line, SequenceLink) {
		return nil, errors.New("invalid link start sequence")
	} else if len(line) < len(SequenceLink)+1+1+1 {
		return nil, errors.New("invalid link value")
	}
	url, label, found := strings.Cut(line[2:], " ")
	if !found {
		return &Link{URL: url, Label: url}, nil
	}
	// TODO: Check charset.
	return &Link{URL: url, Label: label}, nil
}
