package nudoc

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Reader struct {
	r    *bufio.Reader
	line int
}

func NewReader(r io.Reader) *Reader {
	return &Reader{r: bufio.NewReader(r), line: 1}
}

type ReaderError struct {
	Err  error
	Line int
}

func (rerr *ReaderError) Error() string {
	return fmt.Sprintf("%s (on line %d)", rerr.Err, rerr.Line)
}

func (r *Reader) WrapErr(err error) error {
	return &ReaderError{Err: err, Line: r.line}
}

func (r *Reader) Line() int { return r.line }

// Reads a line and trims trailing (CR and) LF characters.
func (r *Reader) ReadLine() (string, error) {
	line, err := r.r.ReadString('\n')
	if err != nil {
		return "", err
	}
	r.line++
	return strings.TrimRight(line, "\r\n"), nil
}
