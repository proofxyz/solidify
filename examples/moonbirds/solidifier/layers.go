package main

import (
	"fmt"
	"image"
	"image/color"
	"path/filepath"
	"sort"
	"strings"

	"github.com/proofxyz/solidify/go/aggregators"
	"github.com/proofxyz/solidify/go/storage"
	"github.com/proofxyz/solidify/go/types"
)

const (
	maxStorageSize       = 18000
	maxBucketsPerStorage = 100
	maxBucketSize        = 4000
)

type layerGroup struct {
	name string
	ims  []*types.Image
}

func (l layerGroup) Name() string {
	return l.name
}

func (l layerGroup) NumFields() int {
	return len(l.ims)
}

func getLayers(gs []types.FeatureGroup, assetsDir string) ([]layerGroup, error) {
	layerDir := filepath.Join(assetsDir, "moonbirds-assets", "traits")

	loadLayer := func(path string, l *layerGroup) error {
		i, err := types.LoadImage(path)
		if err != nil {
			return fmt.Errorf("types.LoadImage(%q): %w", path, err)
		}
		im := &types.Image{Image: i}
		im.IgnoreSpuriousAlpha(254)
		l.ims = append(l.ims, im)
		return nil
	}

	var layers []layerGroup

	// Handle gradients manually
	{
		l := layerGroup{name: "Gradients"}
		for _, n := range []string{"Cosmic", "Enlightened", "Jade"} {
			if err := loadLayer(filepath.Join(layerDir, "Gradients", n+".png"), &l); err != nil {
				return nil, err
			}
		}
		layers = append(layers, l)
	}

	for _, g := range gs {
		// Exclude backgrounds since they are all mono-colour or gradients
		if g.Type == "Background" {
			continue
		}

		l := layerGroup{name: g.Type}
		for _, v := range g.NonZeroValues {
			path := filepath.Join(layerDir, g.Type, v+".png")

			// Use premultiplied bodies for enlightened to avoid bad visuals
			// with the proof background.
			// This leaves normal birds completely unaffected.
			if g.Type == "Body" && strings.HasPrefix(v, "Enlightened") {
				fg, err := types.LoadImage(path)
				if err != nil {
					return nil, fmt.Errorf("types.LoadImage(%q): %w", path, err)
				}

				bg, err := types.LoadImage(filepath.Join(layerDir, "Gradients", "Enlightened.png"))
				if err != nil {
					return nil, fmt.Errorf("types.LoadImage(%q): %w", path, err)
				}

				i := blend(bg, fg)

				im := &types.Image{Image: i}
				im.IgnoreSpuriousAlpha(254)
				l.ims = append(l.ims, im)

				continue
			}

			if err := loadLayer(path, &l); err != nil {
				return nil, err
			}
		}
		layers = append(layers, l)
	}

	// Add collective background
	{
		l := layerGroup{name: "Special"}
		if err := loadLayer(filepath.Join(assetsDir, "collective.small.png"), &l); err != nil {
			return nil, err
		}
		layers = append(layers, l)
	}

	return layers, nil
}

func blend(bg, fg image.Image) image.Image {
	r := bg.Bounds()
	i := image.NewRGBA(r)

	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			// Undoing the alpha pre-multiplication
			f := color.NRGBAModel.Convert(fg.At(x, y)).(color.NRGBA)

			// If there is no foreground we leave the canvas empty
			if f.A == 0 {
				continue
			}

			// Manual alpha blending
			b := color.NRGBAModel.Convert(bg.At(x, y)).(color.NRGBA)
			na := uint(255 - f.A)
			R := (uint(b.R)*na + uint(f.R)*uint(f.A)) + 0x80
			G := (uint(b.G)*na + uint(f.G)*uint(f.A)) + 0x80
			B := (uint(b.B)*na + uint(f.B)*uint(f.A)) + 0x80

			R = ((R >> 8) + R) >> 8
			G = ((G >> 8) + G) >> 8
			B = ((B >> 8) + B) >> 8

			i.Set(x, y, color.RGBA{uint8(R), uint8(G), uint8(B), 255})
		}
	}

	return i
}

func packLayers(layers []layerGroup) ([]*aggregators.BucketStorage, error) {
	sort.Slice(layers, func(i, j int) bool {
		return layers[i].name < layers[j].name
	})

	var images []*types.Image
	for _, l := range layers {
		images = append(images, l.ims...)
	}

	buckets, err := aggregators.GroupIntoIndexedBuckets(images, maxBucketSize)
	if err != nil {
		return nil, fmt.Errorf("aggregators.GroupIntoIndexedBuckets(%T, %d): %w", images, maxBucketSize, err)
	}

	stores, err := aggregators.GroupIntoStorages(buckets, maxStorageSize, maxBucketsPerStorage, "Layer")
	if err != nil {
		return nil, fmt.Errorf("aggregators.GroupIntoStorages(%T, %v, %v, %q): %w", buckets, maxStorageSize, maxBucketsPerStorage, "Layer", err)
	}

	return stores, nil
}

func processLayers(fTypes []types.FeatureGroup, assetsDir, outDir string) ([]string, error) {
	layers, err := getLayers(fTypes, assetsDir)
	if err != nil {
		return nil, fmt.Errorf("getLayers(%T): %w", fTypes, err)
	}

	stores, err := packLayers(layers)
	if err != nil {
		return nil, fmt.Errorf("packLayers(%T): %w", layers, err)
	}

	if err = PrintStorageStats(stores); err != nil {
		return nil, fmt.Errorf("PrintStorageStats(%T): %w", stores, err)
	}

	return storage.WriteGroupStorage("Layer", layers, stores, outDir)
}
