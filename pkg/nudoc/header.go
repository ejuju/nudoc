package nudoc

import (
	"errors"
	"fmt"
	"html/template"
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

func (h *Header) NuDoc() (v string) {
	v += KeyName + ": " + h.Name + "\n"
	v += KeyDesc + ": " + h.Desc + "\n"
	v += KeySlug + ": " + h.Slug + "\n"
	v += KeyDate + ": " + h.Date.Format(time.DateOnly) + "\n"
	v += KeyTags + ": " + strings.Join(h.Tags, ", ") + "\n"
	v += "---\n"
	return v
}

func (h *Header) Markdown() (v string) {
	v += "---\n"
	v += KeyName + ": " + h.Name + "\n"
	v += KeyDesc + ": " + h.Desc + "\n"
	v += KeySlug + ": " + h.Slug + "\n"
	v += KeyDate + ": " + h.Date.Format(time.DateOnly) + "\n"
	v += KeyTags + ": " + strings.Join(h.Tags, ", ") + "\n"
	v += "---\n"
	return v
}

func (h *Header) Text() (v string) {
	v += KeyName + ": " + h.Name + "\n"
	v += KeyDesc + ": " + h.Desc + "\n"
	v += KeySlug + ": " + h.Slug + "\n"
	v += KeyDate + ": " + h.Date.Format(time.DateOnly) + "\n"
	v += KeyTags + ": " + strings.Join(h.Tags, ", ") + "\n"
	v += "---\n"
	return v
}

func (h *Header) HTML() template.HTML {
	const tmpl = `
	<section id="top">
            <p id="tags">%s</p>
            <p id="date">%s</p>
            <h1>%s</h1>
            <p>%s</p>
    </section>
	`
	return template.HTML(fmt.Sprintf(tmpl, strings.Join(h.Tags, ", "), h.Date.Format(time.DateOnly), h.Name, h.Desc))
}

func ParseHeader(r *Reader) (*Header, error) {
	header := &Header{}
	for {
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
	tags = strings.Split(v, ", ")
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
	if v == "" {
		return "", fmt.Errorf("%w: empty", ErrHeaderInvalidTag)
	}
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
