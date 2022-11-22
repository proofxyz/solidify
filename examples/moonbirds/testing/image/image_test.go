package image

import (
	"image"
	"image/color"
	"testing"
)

func TestCompareImages(t *testing.T) {
	tests := []struct {
		a, b     func() image.Image
		wantErr  bool
		wantDiff uint64
	}{
		{
			func() image.Image {
				im := image.NewRGBA(image.Rect(0, 0, 1, 1))
				im.SetRGBA(0, 0, color.RGBA{1, 2, 3, 4})
				return im
			},
			func() image.Image {
				im := image.NewRGBA(image.Rect(0, 0, 1, 1))
				im.SetRGBA(0, 0, color.RGBA{1, 2, 3, 4})
				return im
			},
			false,
			0,
		},
		{
			func() image.Image {
				im := image.NewRGBA(image.Rect(0, 0, 1, 1))
				im.SetRGBA(0, 0, color.RGBA{1, 2, 3, 4})
				return im
			},
			func() image.Image {
				im := image.NewRGBA(image.Rect(0, 0, 1, 1))
				im.SetRGBA(0, 0, color.RGBA{2, 2, 3, 0})
				return im
			},
			false,
			5,
		},
		{
			func() image.Image {
				im := image.NewRGBA(image.Rect(0, 0, 1, 2))
				return im
			},
			func() image.Image {
				im := image.NewRGBA(image.Rect(0, 0, 1, 1))
				return im
			},
			true,
			0,
		},
	}
	for _, tt := range tests {
		diff, err := CompareImages(tt.a(), tt.b())
		if tt.wantErr {
			if err == nil {
				t.Errorf("Want error but got none")
			}
			continue
		}

		if !tt.wantErr && err != nil {
			t.Errorf("Got unexpected error: %v", err)
		}

		// accounting for premultiplication
		diff >>= 8

		if diff != tt.wantDiff {
			t.Errorf("Unexpected difference, want %d, got %d", tt.wantDiff, diff)
		}
	}
}
