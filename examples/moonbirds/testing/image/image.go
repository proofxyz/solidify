// Pacakge image provides routines to load and compare images.
package image

import (
	"fmt"
	"image"
	"image/color"

	"github.com/proofxyz/solidify/go/types"
	"golang.org/x/image/draw"
)

// LoadReferenceImage loads a reference image from disc and scales it
// using nearest-neighbors by a given factor.
func LoadReferenceImage(path string, scaleup int) (image.Image, error) {
	ref, err := types.LoadImage(path)
	if err != nil {
		return nil, fmt.Errorf("types.LoadImage(%q): %w", path, err)
	}

	refScaled := image.NewRGBA(
		image.Rect(
			0,
			0,
			ref.Bounds().Max.X*scaleup,
			ref.Bounds().Max.Y*scaleup,
		),
	)
	draw.NearestNeighbor.Scale(refScaled, refScaled.Rect, ref, ref.Bounds(), draw.Over, nil)

	return refScaled, nil
}

// CompareImages compares two images and returns their difference.
func CompareImages(img1, img2 image.Image) (uint64, error) {
	if img1.Bounds() != img2.Bounds() {
		return 0, fmt.Errorf("image bounds mismatch: %+v != %+v", img1.Bounds(), img2.Bounds())
	}

	var diff uint64
	for y := img1.Bounds().Min.Y; y < img1.Bounds().Max.Y; y++ {
		for x := img1.Bounds().Min.X; x < img1.Bounds().Max.X; x++ {
			diff += compareColor(img1.At(x, y), img2.At(x, y))
		}
	}

	return diff, nil
}

func absDiff(x, y uint32) uint64 {
	d := int(x) - int(y)
	if d < 0 {
		return uint64(-d)
	}
	return uint64(d)
}

func compareColor(a, b color.Color) (diff uint64) {
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()

	diff += absDiff(r1, r2)
	diff += absDiff(g1, g2)
	diff += absDiff(b1, b2)
	diff += absDiff(a1, a2)
	return diff
}
