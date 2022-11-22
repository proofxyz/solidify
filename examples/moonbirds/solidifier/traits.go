package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/proofxyz/solidify/go/aggregators"
	"github.com/proofxyz/solidify/go/storage"
	"github.com/proofxyz/solidify/go/types"
)

type traitGroup struct {
	name   string
	values []types.StringField
}

func (t traitGroup) Name() string {
	return t.name
}

func (t traitGroup) NumFields() int {
	return len(t.values)
}

func getTraits(gs []types.FeatureGroup) []traitGroup {
	traits := make([]traitGroup, len(gs))

	for i, g := range gs {
		if g.Type == "Gradients" {
			continue
		}

		vs := make([]string, len(g.NonZeroValues))
		copy(vs, g.NonZeroValues)

		// Need to trim colors info for Beak and Eyes metadata
		if g.Type == "Eyes" || g.Type == "Beak" {
			for i, v := range vs {
				v = strings.Split(v, " - ")[0]

				if strings.Contains(v, "(") || strings.Contains(v, "Glitch") {
					v = ""
				}

				vs[i] = v
			}
		}

		ts := make([]types.StringField, len(vs))
		for i, v := range vs {
			ts[i] = types.StringField(v)
		}

		traits[i] = traitGroup{
			name:   g.Type,
			values: ts,
		}
	}

	return traits
}

func packTraits(traits []traitGroup) ([]*aggregators.BucketStorage, error) {
	sort.Slice(traits, func(i, j int) bool {
		return traits[i].name < traits[j].name
	})

	store := aggregators.NewBucketStorage("TraitBucketStorage0")
	for _, trait := range traits {
		b := new(aggregators.IndexedBucket)

		for _, v := range trait.values {
			err := b.AddField(v)
			if err != nil {
				return nil, fmt.Errorf("%T.AddField(%v): %w", b, v, err)
			}
		}

		store.AddBucket(b)
	}

	return []*aggregators.BucketStorage{store}, nil
}

func processTraits(fTypes []types.FeatureGroup, outDir string) ([]string, error) {
	traits := getTraits(fTypes)
	stores, err := packTraits(traits)
	if err != nil {
		return nil, fmt.Errorf("packTraits(%v): %w", traits, err)
	}

	if err := PrintStorageStats(stores); err != nil {
		return nil, fmt.Errorf("PrintStorageStats(%T): %w", stores, err)
	}

	fnames, err := storage.WriteGroupStorage("Trait", traits, stores, outDir)
	if err != nil {
		return nil, fmt.Errorf("storage.WriteGroupStorage(%q, %T, %T, %q): %w", "Trait", traits, stores, outDir, err)
	}

	return fnames, nil
}
