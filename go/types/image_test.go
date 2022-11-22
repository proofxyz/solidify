package types

import (
	"image"
	"image/color"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestShrinkBounds(t *testing.T) {
	tests := []struct {
		name string
		im   image.Image
		want image.Rectangle
	}{
		{
			name: "Single point",
			im: (func() image.Image {
				im := image.NewRGBA(image.Rect(0, 0, 4, 4))
				im.SetRGBA(2, 2, color.RGBA{0, 0, 0, 1})
				return im
			})(),
			want: image.Rectangle{
				Min: image.Point{2, 2},
				Max: image.Point{3, 3},
			},
		},
		{
			name: "2x2 block",
			im: (func() image.Image {
				im := image.NewRGBA(image.Rect(0, 0, 4, 4))
				im.SetRGBA(2, 1, color.RGBA{0, 0, 0, 1})
				im.SetRGBA(3, 2, color.RGBA{0, 0, 0, 1})
				return im
			})(),
			want: image.Rectangle{
				Min: image.Point{2, 1},
				Max: image.Point{4, 3},
			},
		},
		{
			name: "Unshrinkable",
			im: (func() image.Image {
				im := image.NewRGBA(image.Rect(0, 0, 4, 4))
				im.SetRGBA(0, 3, color.RGBA{0, 0, 0, 1})
				im.SetRGBA(3, 0, color.RGBA{0, 0, 0, 1})
				return im
			})(),
			want: image.Rectangle{
				Min: image.Point{0, 0},
				Max: image.Point{4, 4},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shrinkBounds(tt.im)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("shrinkBounds([im]) diff (-want +got):\n%v", diff)
			}
		})
	}
}

func TestCleanAlpha(t *testing.T) {
	tests := []struct {
		name      string
		threshold uint8
		abgrPix   []uint8
		wantPix   []uint8
	}{
		{
			name:      "With spurious alpha",
			threshold: 254,
			abgrPix:   []uint8{255, 0, 0, 0, 254, 0, 0, 0},
			wantPix:   []uint8{255, 0, 0, 0, 255, 0, 0, 0},
		},
		{
			name:      "Without spurious alpha",
			threshold: 254,
			abgrPix:   []uint8{253, 0, 0, 0, 254, 0, 0, 0},
			wantPix:   []uint8{253, 0, 0, 0, 254, 0, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanSpuriousAlphaAbove(tt.abgrPix, tt.threshold)

			if diff := cmp.Diff(tt.wantPix, tt.abgrPix); diff != "" {
				t.Errorf("cleanSpuriousAlphaAbove([abgrPix], [threshold]) diff (-want +got):\n%v", diff)
			}
		})
	}
}

func TestRemoveAlpha(t *testing.T) {
	tests := []struct {
		name      string
		abgrPix   []uint8
		wantPix   []uint8
		wantAlpha bool
	}{
		{
			name:      "Not reducible",
			abgrPix:   []uint8{255, 0, 0, 0, 254, 0, 0, 0},
			wantPix:   []uint8{255, 0, 0, 0, 254, 0, 0, 0},
			wantAlpha: true,
		},
		{
			name:      "Reducible",
			abgrPix:   []uint8{255, 1, 2, 3, 255, 4, 5, 6},
			wantPix:   []uint8{1, 2, 3, 4, 5, 6},
			wantAlpha: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPix, gotAlpha := removeAlphaIfPossible(tt.abgrPix)

			if diff := cmp.Diff(gotPix, tt.wantPix); diff != "" {
				t.Errorf("removeAlphaIfPossible([abgrPix]) pixel diff (-want +got):\n%v", diff)
			}

			if gotAlpha != tt.wantAlpha {
				t.Errorf("removeAlphaIfPossible([abgrPix]) hasAlpha = %v, want %v", gotAlpha, tt.wantAlpha)
			}
		})
	}
}

func TestImageEncode(t *testing.T) {
	tests := []struct {
		name                 string
		im                   image.Image
		ignoreAlphaThreshold uint8
		wantBlob             []byte
	}{
		{
			name: "Horizontal only",
			im: (func() image.Image {
				im := image.NewRGBA(image.Rect(0, 0, 2, 1))
				im.SetRGBA(0, 0, color.RGBA{1, 2, 3, 255})
				im.SetRGBA(1, 0, color.RGBA{4, 5, 6, 255})
				return im
			})(),
			wantBlob: []uint8{
				0, 0, 0, 2, 1,
				3, 2, 1, 6, 5, 4,
			},
		},
		{
			name: "Vertical only, with inversion",
			im: (func() image.Image {
				im := image.NewRGBA(image.Rect(0, 0, 1, 2))
				im.SetRGBA(0, 0, color.RGBA{1, 2, 3, 255})
				im.SetRGBA(0, 1, color.RGBA{4, 5, 6, 255})
				return im
			})(),
			wantBlob: []uint8{
				0, 0, 0, 1, 2,
				6, 5, 4, 3, 2, 1,
			},
		},
		{
			name: "With spurious alpha",
			im: (func() image.Image {
				im := image.NewRGBA(image.Rect(0, 0, 1, 1))
				im.SetRGBA(0, 0, color.RGBA{1, 2, 3, 254})
				return im
			})(),
			ignoreAlphaThreshold: 255,
			wantBlob: []uint8{
				1, 0, 0, 1, 1,
				254, 3, 2, 1,
			},
		},
		{
			name: "With spurious alpha and cleaning",
			im: (func() image.Image {
				im := image.NewRGBA(image.Rect(0, 0, 1, 1))
				im.SetRGBA(0, 0, color.RGBA{1, 2, 3, 254})
				return im
			})(),
			ignoreAlphaThreshold: 254,
			wantBlob: []uint8{
				0, 0, 0, 1, 1,
				3, 2, 1,
			},
		},
		{
			name: "With regular alpha",
			im: (func() image.Image {
				im := image.NewRGBA(image.Rect(0, 0, 1, 1))
				im.SetRGBA(0, 0, color.RGBA{1, 2, 3, 253})
				return im
			})(),
			wantBlob: []uint8{
				1, 0, 0, 1, 1,
				253, 3, 2, 1,
			},
		},
		{
			name: "With regular alpha and cleaning",
			im: (func() image.Image {
				im := image.NewRGBA(image.Rect(0, 0, 1, 1))
				im.SetRGBA(0, 0, color.RGBA{1, 2, 3, 253})
				return im
			})(),
			ignoreAlphaThreshold: 254,
			wantBlob: []uint8{
				1, 0, 0, 1, 1,
				253, 3, 2, 1,
			},
		},
		{
			name: "With reduced frame",
			im: (func() image.Image {
				im := image.NewRGBA(image.Rect(0, 0, 3, 3))
				im.SetRGBA(0, 0, color.RGBA{1, 2, 3, 255})
				return im
			})(),
			wantBlob: []uint8{
				0, 0, 2, 1, 3,
				3, 2, 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			im := Image{tt.im, nil}
			if tt.ignoreAlphaThreshold > 0 {
				im.IgnoreSpuriousAlpha(tt.ignoreAlphaThreshold)
			}
			blob, err := im.Encode()
			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			if diff := cmp.Diff(blob, tt.wantBlob); diff != "" {
				t.Errorf("Wrong blob: %s", diff)
			}
		})
	}
}
