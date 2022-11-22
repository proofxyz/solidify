package image

import (
	"bytes"
	"fmt"
	"image"

	"github.com/vincent-petithory/dataurl"
	"golang.org/x/image/bmp"
)

// ParseBMPURI parses a dataURI encoded BMP.
func ParseBMPURI(u string) (image.Image, error) {
	uri, err := dataurl.DecodeString(u)
	if err != nil {
		return nil, fmt.Errorf("dataurl.DecodeString([data]): %v", err)
	}

	r := bytes.NewReader(uri.Data)
	im, err := bmp.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("bmp.Decode(%T): %v", r, err)
	}

	return im, nil
}
