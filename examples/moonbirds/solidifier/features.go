package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/daragao/merkletree"
	"github.com/proofxyz/solidify/go/aggregators"
	"github.com/proofxyz/solidify/go/extractor"
	"github.com/proofxyz/solidify/go/storage"
	"github.com/proofxyz/solidify/go/types"
	"go.uber.org/multierr"
)

const (
	numFeatureFields       = 7
	maxFeaturesBucketSize  = numFeatureFields * 200
	maxFeaturesStorageSize = 20000
)

func getFeatures(assetsDir string) ([]types.FeatureGroup, []types.Token, error) {
	featuresJSONPath := filepath.Join(assetsDir, "moonbirds-assets", "traits.json")

	fMaps, err := extractor.LoadAndParseFeatureJSON(featuresJSONPath)
	if err != nil {
		return nil, nil, fmt.Errorf("LoadAndParseFeatureJSON(%q): %w", featuresJSONPath, err)
	}

	// Removing unnecessary features
	for i, fm := range fMaps {
		fm["Body"] = fmt.Sprintf("%s - %s", fm["Body"], fm["Feathers"])
		delete(fm, "Feathers")
		delete(fm, "Specie")
		fMaps[i] = fm
	}

	fTypes := extractor.ExtractFeatureGroups(fMaps)

	// Manually put the gradients at the end of the features for easier encoding
	var bgs []string
	for _, v := range fTypes[0].NonZeroValues {
		if v == "Cosmic Purple" || v == "Enlightened Purple" || v == "Jade Green" {
			continue
		}
		bgs = append(bgs, v)
	}
	bgs = append(bgs, "Cosmic Purple", "Enlightened Purple", "Jade Green")
	fTypes[0].NonZeroValues = bgs

	features, err := extractor.EncodeFeatures(fMaps, fTypes)
	if err != nil {
		return nil, nil, err
	}

	tokens := make([]types.Token, len(features))
	for i, v := range features {
		tokens[i] = types.Token{TokenID: uint16(i), Features: v}
	}

	return fTypes, tokens, nil
}

func processFeatures(fTypes []types.FeatureGroup, tokens []types.Token, mt *merkletree.MerkleTree, outDir string) ([]string, error) {
	buckets, err := aggregators.GroupIntoLabelledBuckets(tokens, maxFeaturesBucketSize)
	if err != nil {
		return nil, fmt.Errorf("aggregators.GroupIntoLabelledBuckets(%T, %d): %w", tokens, maxFeaturesBucketSize, err)
	}

	stores, err := aggregators.GroupIntoStorages(buckets, maxFeaturesStorageSize, -1, "Features")
	if err != nil {
		return nil, fmt.Errorf("aggregators.GroupIntoStorages(%T, %v, %v, %q): %w", buckets, maxFeaturesStorageSize, -1, "Features", err)
	}

	if err := PrintStorageStats(stores); err != nil {
		return nil, fmt.Errorf("PrintStorageStats(%T): %w", stores, err)
	}

	forgeFeaturesJSON := filepath.Join(outDir, "features.json")
	if err := storage.WriteFeaturesJSONToFile(fTypes, tokens, forgeFeaturesJSON); err != nil {
		return nil, fmt.Errorf("storage.WriteAllFeaturesJSON(%T, %T, %q): %w", fTypes, tokens, forgeFeaturesJSON, err)
	}

	fs, err := storage.WriteFeaturesContracts(fTypes, stores, mt, outDir)
	if err != nil {
		return nil, fmt.Errorf("storage.WriteFeaturesContracts(%T, %T, %T, %q): %w", fTypes, stores, mt, outDir, err)
	}

	return fs, nil
}

func computeMerkleTree(tokens []types.Token) (*merkletree.MerkleTree, error) {
	mt, err := types.ComputeMerkleTree(tokens)
	if err != nil {
		return nil, fmt.Errorf("utils.ComputeMerkleTree(%T): %w", tokens, err)
	}
	return mt, nil
}

func writeMerkleProofs(tokens []types.Token, mt *merkletree.MerkleTree, outDir string) (retErr error) {
	proofs := make([][]string, len(tokens))

	for i, v := range tokens {
		path, _, err := mt.GetMerklePath(v)
		if err != nil {
			return fmt.Errorf("mt.GetMerklePath([token]): %v", err)
		}

		proof := make([]string, len(path))
		for j, x := range path {
			proof[j] = fmt.Sprintf("0x%x", x)
		}
		proofs[i] = proof
	}

	path := filepath.Join(outDir, "proofs.json")
	w, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("io.Open(%q): %w", path, err)
	}
	defer (func() {
		retErr = multierr.Combine(retErr, w.Close())
	})()

	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	if err := enc.Encode(proofs); err != nil {
		return fmt.Errorf("%T.Encode(%T): %w", enc, proofs, err)
	}

	return nil
}
