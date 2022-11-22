package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/daragao/merkletree"
	"github.com/proofxyz/solidify/go/types"
	"go.uber.org/multierr"

	_ "embed"
)

var (
	tmplFuncsFeatures = addTemplateFuncs(tmplFuncsCommon, template.FuncMap{
		"lastTokenLabel": func(x any) int {
			lastLabel, err := lastTokenLabel(x)
			if err != nil {
				panic(fmt.Errorf("lastTokenLabel([stores]): %w", err))
			}
			return int(lastLabel)
		},
		"reversed": func(s []FeatureGroup) []FeatureGroup {
			a := make([]FeatureGroup, len(s))
			for i := 0; i < len(s); i++ {
				a[i] = s[len(s)-1-i]
			}
			return a
		},
		"unusedBits": func(s []FeatureGroup) int {
			return 256 - len(s)*8
		},
		"usedBits": func(s []FeatureGroup) int {
			return len(s) * 8
		},
		"sorted": func(groups []FeatureGroup) []FeatureGroup {
			gs := make([]FeatureGroup, len(groups))
			copy(gs, groups)
			sort.Slice(gs, func(i, j int) bool {
				return (strings.Compare(gs[i].Name(), gs[j].Name()) == -1)
			})
			return gs
		},
	})

	//go:embed templates/labelled-storage-mapping.go.tmpl
	rawLabelledMappingTmpl string

	labelledStorageMappingTmpl = template.Must(
		template.New("labelled-storage-mapping").Funcs(tmplFuncsFeatures).Parse(rawLabelledMappingTmpl),
	)

	//go:embed templates/features-lib.go.tmpl
	rawFeaturesLibTmpl string

	featuresLibTmpl = template.Must(
		template.New("features-lib").Funcs(tmplFuncsFeatures).Parse(rawFeaturesLibTmpl),
	)
)

func last[T any](l []T) T {
	return l[len(l)-1]
}

func lastTokenLabel(x any) (uint16, error) {
	switch v := x.(type) {
	case BucketStorage:
		return lastTokenLabel(last(v.Buckets()))
	case []Bucket:
		return lastTokenLabel(last(v))
	case LabelledBucket:
		return last(v.Labels()), nil
	default:
		return 0, fmt.Errorf("type %T not supported by lastTokenLabel()", v)
	}
}

// LabelledField is a field with an additional label
type LabelledField interface {
	Field
	Label() uint16
}

// LabelledBucket is a bucket that contains labelled fields
type LabelledBucket interface {
	Bucket
	Labels() []uint16
}

// WriteLabelledStorageMappingFeatures writes the storage mapping to retrieve data from
// storage contracts.
func WriteLabelledStorageMappingFeatures[S BucketStorage](name string, stores []S, w io.Writer) error {
	return labelledStorageMappingTmpl.Execute(w,
		struct {
			Name   string
			Stores []BucketStorage
		}{
			Name:   name,
			Stores: convertStorages(stores),
		})
}

// FeatureGroup denotes a certain types of features.
// E.g. "Background" with values "Black" and "White"
type FeatureGroup interface {
	Name() string
	NumValues() uint8
}

// WriteFeaturesLib writes a solidity file defining the Features struct and
// a helper library to work with it.
func WriteFeaturesLib[F FeatureGroup](groups []F, mt *merkletree.MerkleTree, w io.Writer) error {
	return featuresLibTmpl.Execute(w,
		struct {
			FeatureGroups []FeatureGroup
			MerkleRoot    []byte
			NumTokens     int
		}{
			FeatureGroups: convertFeatureGroups(groups),
			MerkleRoot:    mt.MerkleRoot(),
			NumTokens:     len(mt.Leafs),
		},
	)
}

// WriteFeaturesContracts is a convenience wrapper that writes all contracts
// relevant for storing and working with the features on-chain.
func WriteFeaturesContracts[G FeatureGroup, S BucketStorage](groups []G, stores []S, mt *merkletree.MerkleTree, outputDir string) ([]string, error) {
	var fs fileGenerator
	storageSubdir := "storage"

	errs := []error{
		fs.writeSolFile(outputDir, "Features", func(f *os.File) error {
			return annotateNonNil(WriteFeaturesLib(groups, mt, f), "storeate.WriteFeaturesLib(…)")
		}),
		fs.writeSolFile(outputDir, "FeaturesStorageDeployer", func(f *os.File) error {
			return annotateNonNil(WriteStorageDeployer("Features", "./"+storageSubdir, stores, f), "storage.WriteStorageDeployer(…)")
		}),
		fs.writeSolFile(outputDir, "FeaturesStorageMapping", func(f *os.File) error {
			return annotateNonNil(WriteLabelledStorageMappingFeatures("Features", stores, f), "storage.WriteStorageMappingFeatures(…)")
		}),
	}
	if err := multierr.Combine(errs...); err != nil {
		return nil, err
	}

	for _, s := range stores {
		if err := fs.writeSolFile(filepath.Join(outputDir, storageSubdir), s.Name(), func(f *os.File) error {
			return annotateNonNil(WriteBucketStorage(s, f), "storage.WriteBucketStorage(…)")
		}); err != nil {
			return nil, err
		}
	}

	return fs.created, nil
}

// WriteFeaturesJSON writes the features of a set of tokens as JSON.
// This is intended to be used with forge's vm.parseJson for testing.
func WriteFeaturesJSON[G FeatureGroup](gs []G, ts []types.Token, w io.Writer) error {
	features := make([]map[string]uint8, len(ts))
	for i, t := range ts {
		features[i] = map[string]uint8{}
		for j, g := range gs {
			features[i][g.Name()] = t.Features[j]
		}
	}

	d := map[string]any{"features": features}

	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	if err := enc.Encode(d); err != nil {
		return fmt.Errorf("%T.Encode(%T): %w", enc, d, err)
	}

	return nil
}

// WriteFeaturesJSONToFile is a convenience wrapper to write the JSON created by
// `WriteFeaturesJSON` to a file.
func WriteFeaturesJSONToFile[G FeatureGroup](gs []G, ts []types.Token, path string) (retErr error) {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf(" os.Create(%q): %w", path, err)
	}
	defer (func() {
		retErr = multierr.Combine(retErr, f.Close())
	})()

	return WriteFeaturesJSON(gs, ts, f)
}

func convertFeatureGroups[F FeatureGroup](fs []F) []FeatureGroup {
	fs2 := make([]FeatureGroup, len(fs))
	for i, v := range fs {
		fs2[i] = v
	}
	return fs2
}
