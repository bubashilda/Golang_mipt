//go:build !solution

package otp

import (
	"io"
)

const (
	bufferCapacity int = 32
)

type BufferedReader struct {
	reader          io.Reader
	encryptor       io.Reader
	bufferReader    []byte
	bufferEncryptor []byte
	offsetReader    int
	offsetEncryptor int
}

func (r *BufferedReader) Read(arr []byte) (int, error) {
	if len(r.bufferEncryptor) == r.offsetEncryptor {
		_, _ = r.encryptor.Read(r.bufferEncryptor)
		r.offsetEncryptor = 0
	}
	if len(r.bufferReader) == r.offsetReader {
		r.bufferReader = r.bufferReader[:cap(r.bufferReader)]
		size, err := r.reader.Read(r.bufferReader)
		if err != nil && err != io.EOF {
			return 0, err
		}
		if size == 0 && err == io.EOF {
			return 0, io.EOF
		}
		r.bufferReader = r.bufferReader[:size]
		r.offsetReader = 0
	}

	toRead := min(len(r.bufferReader)-r.offsetReader, len(r.bufferEncryptor)-r.offsetEncryptor, len(arr))
	copy(arr, r.bufferReader[r.offsetReader:r.offsetReader+toRead])

	for i := 0; i < toRead; i++ {
		arr[i] ^= r.bufferEncryptor[r.offsetEncryptor+i]
	}

	r.offsetReader += toRead
	r.offsetEncryptor += toRead
	return toRead, nil
}

func NewReader(r io.Reader, prng io.Reader) io.Reader {
	return &BufferedReader{
		reader: r, encryptor: prng,
		bufferReader: make([]byte, bufferCapacity), bufferEncryptor: make([]byte, bufferCapacity),
		offsetReader: bufferCapacity, offsetEncryptor: bufferCapacity}
}

type BufferedWriter struct {
	writer          io.Writer
	encryptor       io.Reader
	bufferEncryptor []byte
	offsetEncryptor int
}

func (w *BufferedWriter) Write(arr []byte) (int, error) {
	indexArr := 0
	for indexArr < len(arr) {
		if len(w.bufferEncryptor) == w.offsetEncryptor {
			_, _ = w.encryptor.Read(w.bufferEncryptor)
			w.offsetEncryptor = 0
		}
		toWrite := min(len(w.bufferEncryptor)-w.offsetEncryptor, len(arr)-indexArr)
		ans := make([]byte, toWrite)
		for i := 0; i < toWrite; i++ {
			ans[i] = w.bufferEncryptor[w.offsetEncryptor+i] ^ arr[indexArr+i]
		}
		size, err := w.writer.Write(ans)
		indexArr += toWrite
		w.offsetEncryptor += toWrite
		if size == 0 || err != nil {
			return indexArr, err
		}
	}
	return len(arr), nil
}

func NewWriter(w io.Writer, prng io.Reader) io.Writer {
	return &BufferedWriter{
		writer: w, encryptor: prng,
		bufferEncryptor: make([]byte, bufferCapacity), offsetEncryptor: bufferCapacity}
}
