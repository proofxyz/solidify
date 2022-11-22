// Package deflate provides binary data compression via the DEFLATE algorithm.
package deflate

import (
	"bytes"
	"compress/flate"
	"fmt"
	"io"
)

// Compressed stores deflated data and the original size of the blob before
// compression.
type Compressed struct {
	Data             []byte
	UncompressedSize int
}

// Deflate deflates a blob of data
func Deflate(data io.Reader) (*Compressed, error) {
	var output bytes.Buffer

	c := flate.BestCompression
	zw, err := flate.NewWriter(&output, c)
	if err != nil {
		return nil, fmt.Errorf("flate.NewWriter(%T, %d): %v", &output, c, err)
	}

	inputSize, err := io.Copy(zw, data)
	if err != nil {
		return nil, fmt.Errorf("deflating data, io.Copy(%T, data): %v", zw, err)
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("%T.Close(): %v", zw, err)
	}

	return &Compressed{
		UncompressedSize: int(inputSize),
		Data:             output.Bytes(),
	}, nil
}

// Inflate inflates a blob of data
func Inflate(c *Compressed) ([]byte, error) {
	var res bytes.Buffer
	zr := flate.NewReader(bytes.NewReader(c.Data))

	if _, err := io.Copy(&res, zr); err != nil {
		return nil, fmt.Errorf("inflating data io.Copy(%T, %T): %v", &res, zr, err)
	}

	if err := zr.Close(); err != nil {
		return nil, fmt.Errorf("%T.Close(): %v", zr, err)
	}

	return res.Bytes(), nil
}
