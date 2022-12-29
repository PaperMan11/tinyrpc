package compressor

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/golang/snappy"
)

func init() {
	Compressors[Snappy] = NewSnappyCompressor()
}

type SnappyCompressor struct{}

func NewSnappyCompressor() Compressor {
	return &SnappyCompressor{}
}

// Zip .
func (*SnappyCompressor) Zip(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w := snappy.NewBufferedWriter(buf)
	defer func() {
		w.Close()
	}()
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Flush()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

// Unzip .
func (*SnappyCompressor) Unzip(data []byte) ([]byte, error) {
	r := snappy.NewReader(bytes.NewBuffer(data))
	data, err := ioutil.ReadAll(r)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return data, nil
}
