package nudoc

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	KeyName = "Name"
	KeyDesc = "Desc"
	KeySlug = "Slug"
	KeyDate = "Date"
	KeyTags = "Tags"
)

var Keys = [...]string{
	KeyName,
	KeyDesc,
	KeySlug,
	KeyDate,
	KeyTags,
}

const MaxHeaderLines = len(Keys) + 1 // All keys + end of header line ("---").

var (
	ErrInvalidDate = errors.New("invalid date")
	ErrInvalidTag  = errors.New("invalid tag")
)

type Header struct {
	Name string
	Desc string
	Slug string
	Date time.Time
	Tags []string
}

func ParseHeader(r *bufio.Reader) (*Header, error) {
	header := &Header{}
	for i := 0; i <= MaxHeaderLines; i++ {
		if i == MaxHeaderLines {
			return nil, fmt.Errorf("too many header lines (#%d)", i)
		}
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("read header line (#%d): %w", i, err)
		}
		line = strings.TrimSpace(line)
		if line == "---" {
			break
		}
		key, value, ok := strings.Cut(line, ": ")
		if !ok {
			return nil, fmt.Errorf("invalid header line (#%d): separator-space not found", i)
		}
		switch key {
		default:
			return nil, fmt.Errorf("invalid header line (#%d): no separator found", i)
		case KeyName:
			header.Name = value
		case KeyDesc:
			header.Desc = value
		case KeySlug:
			header.Slug = value
		case KeyDate:
			header.Date, err = ParseDate(value)
			if err != nil {
				return nil, fmt.Errorf("invalid header date at line index (#%d): %w", i, err)
			}
		case KeyTags:
			header.Tags, err = ParseTags(value)
			if err != nil {
				return nil, fmt.Errorf("invalid header tags at line index (#%d): %w", i, err)
			}
		}
	}

	return header, nil
}

func ParseDate(v string) (time.Time, error) {
	t, err := time.Parse(time.DateOnly, v)
	if err != nil {
		return t, fmt.Errorf("%w: %w", ErrInvalidDate, err)
	}
	return t, nil
}

func ParseTags(v string) (tags []string, err error) {
	tags = strings.Split(v, " ")
	if len(tags) == 0 {
		return nil, fmt.Errorf("%w: no tags found", ErrInvalidTag)
	}
	for i, tag := range tags {
		tag, err = ParseTag(tag)
		if err != nil {
			return nil, err
		}
		tags[i] = tag
	}
	return tags, nil
}

func ParseTag(v string) (string, error) {
	if v == "" {
		return "", fmt.Errorf("%w: empty", ErrInvalidTag)
	} else if !strings.HasPrefix(v, "#") {
		return "", fmt.Errorf("%w: missing leading hashtag", ErrInvalidTag)
	}
	v = v[1:]
	for i, c := range v {
		if !IsValidTagCharacter(c) {
			return "", fmt.Errorf("%w: forbidden character %q at index %d", ErrInvalidTag, c, i)
		}
	}
	return v, nil
}

func IsValidTagCharacter(c rune) bool {
	return (c >= 'a' && c <= 'z') || c == '-'
}
