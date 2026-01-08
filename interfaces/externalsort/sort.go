//go:build !solution

package externalsort

import (
	"container/heap"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"
)

const (
	bufferCapacity = 32
)

type BufferedLineReader struct {
	reader       io.Reader
	buffer       []byte
	offsetBuffer int
}

type BufferedLineWriter struct {
	writer io.Writer
}

func (r *BufferedLineReader) ReadLine() (l string, err error) {
	var ans strings.Builder
	for {
		if r.offsetBuffer == len(r.buffer) {
			prevSize := len(r.buffer)
			r.buffer = r.buffer[:cap(r.buffer)]
			size, err := r.reader.Read(r.buffer)
			if err != nil && err != io.EOF {
				r.buffer = r.buffer[:prevSize]
				return ans.String(), err
			}
			if size == 0 && err == io.EOF {
				r.buffer = r.buffer[:prevSize]
				if ans.String() != "" {
					return ans.String(), nil
				} else {
					return ans.String(), io.EOF
				}
			}
			r.buffer = r.buffer[:size]
			r.offsetBuffer = 0
		}
		idx := slices.Index(r.buffer[r.offsetBuffer:], '\n')
		if idx == -1 {
			ans.Write(r.buffer[r.offsetBuffer:])
			r.offsetBuffer = len(r.buffer)
		} else {
			ans.Write(r.buffer[r.offsetBuffer : r.offsetBuffer+idx])
			r.offsetBuffer += idx + 1
			return ans.String(), nil
		}
	}
}

func (w *BufferedLineWriter) Write(l string) error {
	_, err := w.writer.Write([]byte(l + "\n"))
	return err
}

func NewReader(r io.Reader) LineReader {
	return &BufferedLineReader{reader: r, buffer: make([]byte, bufferCapacity), offsetBuffer: bufferCapacity}
}

func NewWriter(w io.Writer) LineWriter {
	return &BufferedLineWriter{writer: w}
}

type StringHeap []string

func (h *StringHeap) Len() int           { return len(*h) }
func (h *StringHeap) Less(i, j int) bool { return (*h)[i] < (*h)[j] }
func (h *StringHeap) Swap(i, j int)      { (*h)[i], (*h)[j] = (*h)[j], (*h)[i] }

func (h *StringHeap) Push(x any) {
	*h = append(*h, x.(string))
}

func (h *StringHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func Merge(w LineWriter, readers ...LineReader) error {
	h := &StringHeap{}
	heap.Init(h)

	bufferLastString := make([]string, len(readers))
	bufferInvalidated := make([]bool, len(readers))

	for idx, reader := range readers {
		line, err := reader.ReadLine()
		if err != nil && err != io.EOF {
			return fmt.Errorf("error while reading string: %w", err)
		}
		if err == nil {
			heap.Push(h, line)
			bufferLastString[idx] = line
		} else {
			bufferInvalidated[idx] = true
		}
	}

	for h.Len() > 0 {
		minLine := heap.Pop(h).(string)
		if err := w.Write(minLine); err != nil {
			return fmt.Errorf("error while writing string: %w", err)
		}

		for idx, reader := range readers {
			if minLine == bufferLastString[idx] && !bufferInvalidated[idx] {
				line, err := reader.ReadLine()
				if err != nil && err != io.EOF {
					return fmt.Errorf("error while reading string: %w", err)
				}
				if err == nil {
					heap.Push(h, line)
					bufferLastString[idx] = line
				} else {
					bufferInvalidated[idx] = true
				}
				break
			}
		}
	}

	return nil
}

func SortFile(file *os.File) error {
	reader := NewReader(file)

	var lines []string
	for {
		line, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error while reading string: %w", err)
		}
		lines = append(lines, line)
	}

	sort.Strings(lines)
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("error while seeking to start: %w", err)
	}

	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("error while truncating file: %w", err)
	}

	writer := NewWriter(file)

	for _, line := range lines {
		if err := writer.Write(line); err != nil {
			return fmt.Errorf("error while writing string: %w", err)
		}
	}

	return nil
}

func Sort(w io.Writer, in ...string) error {
	for _, filename := range in {
		file, err := os.OpenFile(filename, os.O_RDWR, 0644)
		if err != nil {
			return fmt.Errorf("error while opening file %s: %w", filename, err)
		}

		if err := SortFile(file); err != nil {
			if err := file.Close(); err != nil {
				return fmt.Errorf("error while closing file %s: %w", filename, err)
			}
			return fmt.Errorf("error while sorting file %s: %w", filename, err)
		}

		if err := file.Close(); err != nil {
			return fmt.Errorf("error while closing file %s: %w", filename, err)
		}
	}

	readers := make([]LineReader, 0, len(in))
	for _, filename := range in {
		file, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("error while opening file %s: %w", filename, err)
		}

		defer func(file *os.File) {
			if err := file.Close(); err != nil {
				panic(err)
			}
		}(file)

		reader := NewReader(file)
		readers = append(readers, reader)
	}

	if err := Merge(NewWriter(w), readers...); err != nil {
		return fmt.Errorf("error while merging sorted files: %w", err)
	}

	return nil
}
