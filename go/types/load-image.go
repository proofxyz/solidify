package types

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"

	"golang.org/x/image/bmp"
)

// LoadImage loads an image from a given path.
// Currently supported types (BMP, PNG)
func LoadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("os.Open(%q): %w", path, err)
	}
	defer f.Close()

	return DecodeImage(f)
}

// DecodeImage decodes image data. Currently supported types (BMP, PNG)
func DecodeImage(r interface {
	io.Reader
	io.ReaderAt
}) (image.Image, error) {
	// DetectContentType states that it only requires up to 512 bytes.
	typeInfo, err := io.ReadAll(io.NewSectionReader(r, 0, 512))
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll(io.NewSectionReader(%T, 0, 512)): %v", r, err)
	}

	switch ct := http.DetectContentType(typeInfo); ct {
	case "image/png":
		return decodePNG(r)
	case "image/bmp":
		return decodeBMP(r)
	default:
		return nil, fmt.Errorf("unknown data type %v of type %s", r, ct)
	}
}

func decodePNG(r io.Reader) (image.Image, error) {
	im, err := png.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("png.Decode(%v): %v", r, err)
	}

	return im, nil
}

func decodeBMP(r io.Reader) (image.Image, error) {
	im, err := bmp.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("bmp.Decode(%v): %v", r, err)
	}

	return im, nil
}
