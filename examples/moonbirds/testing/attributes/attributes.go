// Package attributes provides routines to work with token attributes intended
// to be used to build testing helpers for forge.
package attributes

import (
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	_ "embed"

	"github.com/divergencetech/ethier/erc721"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/exp/slices"
)

// UnmarshalReferenceMetadataJSON loads and unmarshals a reference metadata file.
func UnmarshalReferenceMetadataJSON(path string) ([]erc721.Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("os.Open(%q): %w", path, err)
	}
	defer f.Close()

	var meta []erc721.Metadata
	if err := json.NewDecoder(f).Decode(&meta); err != nil {
		return nil, fmt.Errorf("json.NewDecoder([%q]).Decode(%T): %w", path, &meta, err)
	}

	return meta, nil
}

// ReadAttributesCSV loads and parses a CSV file containing attributes.
// Format: type, value
func ReadAttributesCSV(filePath string) ([]*erc721.Attribute, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("os.Open(%q): %w", filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%T.ReadAll(): %w", csvReader, err)
	}

	var a []*erc721.Attribute
	for _, r := range records {
		a = append(a, &erc721.Attribute{TraitType: r[0], Value: r[1]})
	}

	return a, nil
}

// ValidateAttributes validates a given set of attributes agains provided
// reference data.
func ValidateAttributes(got, want []*erc721.Attribute) error {
	less := cmpopts.SortSlices(func(x, y *erc721.Attribute) bool {
		return (strings.Compare(x.TraitType, y.TraitType) == -1)
	})

	if diff := cmp.Diff(want, got, less); diff != "" {
		return fmt.Errorf("attributes diff (-want +got)\n%v", diff)
	}

	return nil
}

// LoadReferenceAttributes loads the attributes of a given token from a reference
// metadata file.
func LoadReferenceAttributes(path string, tokenID int, ignore []string) ([]*erc721.Attribute, error) {
	ref, err := UnmarshalReferenceMetadataJSON(path)
	if err != nil {
		return nil, fmt.Errorf("helpers.UnmarshalReferenceMetadataJSON(%q): %w", path, err)
	}

	var r []*erc721.Attribute
	for _, a := range ref[tokenID].Attributes {
		if slices.Contains(ignore, a.TraitType) {
			continue
		}

		r = append(r, a)
	}

	return r, nil
}
