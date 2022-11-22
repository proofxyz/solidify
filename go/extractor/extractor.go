// Package extractor provides routines to extract token features for the solidify
// toolchain from a standard metadata JSON files or token feature maps.
package extractor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/proofxyz/solidify/go/types"
)

type featureType = string
type featureValue = string

// A FeaturesMap maps feature types to values, defining a token.
// e.g. Background -> Black, Body -> Robot, etc.
type FeaturesMap = map[string]string

// ParseFeaturesJSON parses features from a JSON schema into a list of FeatureMaps.
// e.g. JSON = `[{"Body": "Robot"},{"Body": "Alien"}]`
func ParseFeaturesJSON(r io.Reader) ([]FeaturesMap, error) {
	// Parse into generic type
	var fMap []map[featureType]featureValue
	if err := json.NewDecoder(r).Decode(&fMap); err != nil {
		return nil, fmt.Errorf("json.NewDecoder([data]).Decode(%T): %w", &fMap, err)
	}
	return fMap, nil
}

// LoadAndParseFeatureJSON loads the features from a JSON file at a given path and
// parses the content into a list of FeaturesMap.
func LoadAndParseFeatureJSON(path string) ([]FeaturesMap, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("os.Open(%q): %w", path, err)
	}
	defer f.Close()

	return ParseFeaturesJSON(f)
}

// ExtractFeatureGroups retrieves groups of features (types with their different
// values) from a given set of FeaturesMaps.
func ExtractFeatureGroups(fMaps []FeaturesMap) []types.FeatureGroup {
	valSet := make(map[featureType]map[featureValue]bool)

	// First build sets of feature values depending on type to get unique values
	for _, f := range fMaps {
		for k, v := range f {
			if _, ok := valSet[k]; !ok {
				valSet[k] = map[string]bool{}
			}
			valSet[k][v] = true
		}
	}

	// Retrieve all feature types from the set and sort alphabetically
	var fTypes []types.FeatureGroup
	for k := range valSet {
		fTypes = append(fTypes, types.FeatureGroup{Type: k})
	}
	sort.Slice(fTypes, func(i, j int) bool {
		return fTypes[i].Type < fTypes[j].Type
	})

	for i, t := range fTypes {
		// Get feature values for a given type (excluding "None" since we will
		// add it later again to always have it at first position) and sort
		// alphabetically.
		var vals []featureValue
		for k := range valSet[t.Type] {
			if k == "None" {
				continue
			}
			vals = append(vals, k)
		}
		sort.Strings(vals)
		fTypes[i].NonZeroValues = vals
	}

	return fTypes
}

// EncodeFeatures converts FeatureMaps into uint slices where each element
// corresponds to the index at which the respective feature (type and value) appears
// in the feature groups list.
func EncodeFeatures(fMaps []FeaturesMap, fTypes []types.FeatureGroup) ([][]uint8, error) {
	// This will be our reverse mapping translating feature values into numbers
	// unique to each type.
	enumeration := make(map[featureType]map[featureValue]uint8)

	for _, t := range fTypes {
		// Enumerate feature values and build the lookup table
		tmp := make(map[string]uint8)
		for i, v := range t.Values() {
			tmp[v] = uint8(i)
		}
		enumeration[t.Type] = tmp
	}

	// Apply the features enumeration that was determined before.
	// Missing features get the default value 0
	featureSets := make([][]uint8, len(fMaps))
	for i, fMap := range fMaps {
		featureSet := make([]uint8, len(fTypes))

		for j, t := range fTypes {
			featureValue, ok := fMap[t.Type]
			if !ok {
				// Implicit 0 for missing types
				continue
			}

			featureSet[j], ok = enumeration[t.Type][featureValue]
			if !ok {
				return nil, fmt.Errorf("unknown feature value: type %v has no value %s", t, featureValue)
			}
		}
		featureSets[i] = featureSet
	}

	return featureSets, nil
}
