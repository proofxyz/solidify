package main

import (
	"fmt"
	"os"

	"github.com/proofxyz/solidify/go/aggregators"
	"github.com/proofxyz/solidify/go/storage"
	"github.com/proofxyz/solidify/go/types"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

const (
	genDst                = "./gen/"
	featuresJSON          = "./gen/features.json"
	maxFeaturesBucketSize = 6
)

func run() error {

	gs := []types.FeatureGroup{
		{
			Type:          "FOO",
			NonZeroValues: []string{"foo", "bar"},
		},
		{
			Type:          "BAR",
			NonZeroValues: []string{"baz", "qab", "zaq"},
		},
		{
			Type:          "QUX",
			NonZeroValues: []string{"tud"},
		},
	}

	tokens := []types.Token{
		{TokenID: 0, Features: []uint8{0, 1, 1}},
		{TokenID: 1, Features: []uint8{2, 3, 0}},
		{TokenID: 2, Features: []uint8{1, 2, 0}},
		{TokenID: 3, Features: []uint8{1, 2, 1}}, // Not stored
		{TokenID: 4, Features: []uint8{1, 1, 1}}, // Not stored
		{TokenID: 5, Features: []uint8{1, 0, 0}}, // Not stored
		{TokenID: 6, Features: []uint8{0, 2, 0}},
		{TokenID: 7, Features: []uint8{0, 3, 1}},
	}

	mt, err := types.ComputeMerkleTree(tokens)
	if err != nil {
		return fmt.Errorf("utils.ComputeMerkleTree(%T): %w", tokens, err)
	}

	storedTokens := []types.Token{}
	storedTokens = append(storedTokens, tokens[:3]...)
	storedTokens = append(storedTokens, tokens[6:8]...)

	buckets, err := aggregators.GroupIntoLabelledBuckets(storedTokens, maxFeaturesBucketSize)
	if err != nil {
		return fmt.Errorf("utils.GroupIntoLabelledBuckets(%T, %d): %w", tokens, maxFeaturesBucketSize, err)
	}

	ss, err := aggregators.GroupIntoStorages(buckets, -1, 2, "Features")
	if err != nil {
		return fmt.Errorf("utils.GroupIntoStorages(%T, %v, %v, %q): %w", buckets, -1, 2, "Features", err)
	}

	fNames, err := storage.WriteFeaturesContracts(gs, ss, mt, genDst)
	if err != nil {
		return fmt.Errorf("storage.WriteFeaturesContracts(%q, %T, %T, %q): %w", "Group", gs, ss, genDst, err)
	}

	if err := storage.FormatSol(fNames); err != nil {
		return fmt.Errorf("utils.FormatSol(%v): %w", fNames, err)
	}

	if err := storage.WriteFeaturesJSONToFile(gs, tokens, featuresJSON); err != nil {
		return fmt.Errorf("utils.WriteAllFeaturesJSON(%T, %T, %q): %w", gs, tokens, featuresJSON, err)
	}

	return nil
}
