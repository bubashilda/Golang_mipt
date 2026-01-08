package reader

import (
	"bufio"
	"io"
	"strings"
)

type Wrapper struct {
	reader *bufio.Reader
}

type LineReader interface {
	ReadLine() (string, error)
}

func (w *Wrapper) ReadLine() (string, error) {
	s, err := w.reader.ReadString('\n')
	return strings.TrimSuffix(s, "\n"), err
}

func NewWrapper(r io.Reader) LineReader {
	return &Wrapper{bufio.NewReader(r)}
}
