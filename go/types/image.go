package types

import (
	"image"
	"image/color"

	"golang.org/x/exp/constraints"
)

// Image wraps around the standard go image library and provides an encoding
// function that it can be used to build buckets.
type Image struct {
	image.Image
	ignoreAlphaThreshold *uint8
}

// IgnoreSpuriousAlpha sets the threshold for the alpha cleaning during the serialisation.
// This means that if the image only contains `alpha >= threshold`, all alpha values will be set to 255.
func (i *Image) IgnoreSpuriousAlpha(threshold uint8) {
	i.ignoreAlphaThreshold = &threshold
}

func bmpPixels(im image.Image, bounds image.Rectangle) []uint8 {
	var pix []uint8
	// Reversing row order to match the BMP format standard
	for y := bounds.Max.Y - 1; y >= bounds.Min.Y; y-- {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Undoing the alpha pre-multiplication
			c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)

			// Reversing the color order to match the BMP format standard
			pix = append(pix, c.A, c.B, c.G, c.R)
		}
	}

	return pix
}

// Encode encodes the image as a data blob compatible with the raw BMP row and
// channel ordering + additional alpha and frame rectangle metadata.
//
//	| hasAlpha (byte) | frameRectangle (4 bytes) | blob [(A)BGR (4 or 3 bytes) | ...] |
//
// The blob contains only data for the subframe of the original image with
// non-zero alpha values to save storage space. Further, it does not have an
// alpha channel if all alpha values are 255 anyway.
// Depending on the Image.CleanSpuriousAlpha flag, we map all alpha values of
// 254->255 before encoding if alpha=254 is the lowest value in the image.
func (i Image) Encode() ([]byte, error) {
	r := shrinkBounds(i.Image)
	pix := bmpPixels(i.Image, r)

	if i.ignoreAlphaThreshold != nil {
		cleanSpuriousAlphaAbove(pix, *i.ignoreAlphaThreshold)
	}

	pix, hasAlpha := removeAlphaIfPossible(pix)

	var buf []byte

	// Header bytes
	if hasAlpha {
		buf = append(buf, 1)
	} else {
		buf = append(buf, 0)
	}

	b := i.Image.Bounds()
	// computing relative bounds. NB that the y-bounds reflect the reversing
	// of the rows (see also bmpPixels).
	buf = append(buf,
		uint8(r.Min.X-b.Min.X),
		uint8(b.Max.Y-r.Max.Y),
		uint8(r.Max.X-b.Min.X),
		uint8(b.Max.Y-r.Min.Y),
	)

	// data
	buf = append(buf, pix...)
	return buf, nil
}

func shrinkBounds(i image.Image) image.Rectangle {
	b := i.Bounds()
	r := image.Rectangle{
		Min: b.Max,
		Max: b.Min,
	}

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if _, _, _, a := i.At(x, y).RGBA(); a > 0 {
				r.Min.X = min(r.Min.X, x)
				r.Min.Y = min(r.Min.Y, y)
				r.Max.X = max(r.Max.X, x+1)
				r.Max.Y = max(r.Max.Y, y+1)
			}
		}
	}

	return r
}

func cleanSpuriousAlphaAbove(abgr []uint8, alphaThreshold uint8) {
	var minAlpha uint8 = 255

	for i := 0; i < len(abgr); i += 4 {
		minAlpha = min(minAlpha, abgr[i])
	}

	if minAlpha >= alphaThreshold {
		for i := 0; i < len(abgr); i += 4 {
			abgr[i] = 255
		}
	}
}

func removeAlphaIfPossible(abgr []uint8) ([]uint8, bool) {
	for i := 0; i < len(abgr); i += 4 {
		if abgr[i] < 255 {
			return abgr, true
		}
	}

	j := 0
	for i := 0; i < len(abgr); i += 4 {
		// Move BRG up by one index
		for k := 0; k < 3; k++ {
			abgr[j+k] = abgr[i+1+k]
		}
		j += 3
	}

	return abgr[:j], false
}

func min[T constraints.Integer](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func max[T constraints.Integer](a, b T) T {
	if a > b {
		return a
	}
	return b
}
