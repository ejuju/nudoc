package nudoc

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	KeyName = "Name"
	KeyDesc = "Description"
	KeySlug = "Slug"
	KeyDate = "Date"
	KeyTags = "Tags"
)

var ReservedKeys = [...]string{
	KeyName,
	KeyDesc,
	KeySlug,
	KeyDate,
	KeyTags,
}

// One line for each key, plus the end-of-header line ("---").
const MaxHeaderLines = len(ReservedKeys) + 1

var (
	ErrHeaderTooManyLines      = errors.New("too many header lines")
	ErrHeaderSeparatorNotFound = errors.New("header separator not found")
	ErrHeaderUnknownKey        = errors.New("unknown key in header")
	ErrHeaderInvalidDate       = errors.New("invalid date value in header")
	ErrHeaderInvalidTags       = errors.New("invalid tags value in header")
	ErrHeaderInvalidTag        = errors.New("invalid tag in header")
)

type Header struct {
	Name string
	Desc string
	Slug string
	Date time.Time
	Tags []string
}

func (h *Header) PrettyString() (v string) {
	v += fmt.Sprintf("%s: %q\n", KeyName, h.Name)
	v += fmt.Sprintf("%s: %q\n", KeyDesc, h.Desc)
	v += fmt.Sprintf("%s: %q\n", KeySlug, h.Slug)
	v += fmt.Sprintf("%s: %q\n", KeyDate, h.Date)
	v += fmt.Sprintf("%s: %q\n", KeyTags, h.Tags)
	return v
}

func ParseHeader(r *Reader) (*Header, error) {
	header := &Header{}
	for {
		if r.Line() > MaxHeaderLines {
			return nil, r.WrapErr(ErrHeaderTooManyLines)
		}
		line, err := r.ReadLine()
		if err != nil {
			return nil, r.WrapErr(err)
		} else if line == "---" {
			break
		}
		key, value, ok := strings.Cut(line, ": ")
		if !ok {
			return nil, r.WrapErr(ErrHeaderSeparatorNotFound)
		}
		switch key {
		default:
			return nil, r.WrapErr(fmt.Errorf("%w: %q", ErrHeaderUnknownKey, key))
		case KeyName:
			header.Name = value
		case KeyDesc:
			header.Desc = value
		case KeySlug:
			header.Slug = value
		case KeyDate:
			header.Date, err = ParseHeaderDate(value)
			if err != nil {
				return nil, r.WrapErr(err)
			}
		case KeyTags:
			header.Tags, err = ParseHeaderTags(value)
			if err != nil {
				return nil, r.WrapErr(err)
			}
		}
	}

	return header, nil
}

func ParseHeaderDate(v string) (time.Time, error) {
	t, err := time.Parse(time.DateOnly, v)
	if err != nil {
		return t, fmt.Errorf("%w: %w", ErrHeaderInvalidDate, err)
	}
	return t, nil
}

func ParseHeaderTags(v string) (tags []string, err error) {
	tags = strings.Split(v, " ")
	if len(tags) == 0 {
		return nil, fmt.Errorf("%w: no tags found", ErrHeaderInvalidTag)
	}
	for i, tag := range tags {
		tag, err = ParseHeaderTag(tag)
		if err != nil {
			return nil, err
		}
		tags[i] = tag
	}
	return tags, nil
}

func ParseHeaderTag(v string) (string, error) {
	// Ensure that tag is not empty and starts with a hashtag ("#").
	if v == "" {
		return "", fmt.Errorf("%w: empty", ErrHeaderInvalidTag)
	} else if !strings.HasPrefix(v, "#") {
		return "", fmt.Errorf("%w: missing leading hashtag", ErrHeaderInvalidTag)
	}

	// Trim leading hashtag and check characters.
	v = v[1:]
	for i, c := range v {
		if !IsValidHeaderTagCharacter(c) {
			return "", fmt.Errorf("%w: forbidden character %q at column %d", ErrHeaderInvalidTag, c, i)
		}
	}

	return v, nil
}

func IsValidHeaderTagCharacter(c rune) bool {
	return (c >= 'a' && c <= 'z') || c == '-'
}
