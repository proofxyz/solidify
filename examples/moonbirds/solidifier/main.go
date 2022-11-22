// The solidifier binary generates all contracts and auxiliary files for
// moonbirds in-chain.
package main

import (
	"flag"
	"fmt"

	"github.com/golang/glog"
	"github.com/proofxyz/solidify/go/storage"
)

type config struct {
	outDir, assetsDir string
	writeProofs       bool
}

func main() {
	var c config
	flag.StringVar(&c.outDir, "out", "", "The output directory for the generated contract")
	flag.StringVar(&c.assetsDir, "in", "", "The input directory containing the moonbirds assets")
	flag.BoolVar(&c.writeProofs, "writeProofs", false, "Flag to enable Merkle proof generation.")
	flag.Parse()

	if err := c.run(); err != nil {
		glog.Exit(err)
	}
}

func (c config) run() error {
	fTypes, tokens, err := getFeatures(c.assetsDir)
	if err != nil {
		return fmt.Errorf("getFeatures(): %w", err)
	}

	mt, err := computeMerkleTree(tokens)
	if err != nil {
		return fmt.Errorf("computeMerkleTree([tokens]): %w", err)
	}

	if c.writeProofs {
		if err := writeMerkleProofs(tokens, mt, c.outDir); err != nil {
			return fmt.Errorf("writeMerkleProofs([tokens], [mt]): %w", err)
		}

	}

	var generatedFiles []string

	if fns, err := processLayers(fTypes, c.assetsDir, c.outDir); err != nil {
		return fmt.Errorf("processLayers(): %w", err)
	} else {
		generatedFiles = append(generatedFiles, fns...)
	}

	if fns, err := processTraits(fTypes, c.outDir); err != nil {
		return fmt.Errorf("processTraits(): %w", err)
	} else {
		generatedFiles = append(generatedFiles, fns...)
	}

	if fns, err := processFeatures(fTypes, tokens, mt, c.outDir); err != nil {
		return fmt.Errorf("processFeatures(): %w", err)
	} else {
		generatedFiles = append(generatedFiles, fns...)
	}

	if err := storage.FormatSol(generatedFiles); err != nil {
		return fmt.Errorf("FormatSol(%v): %w", generatedFiles, err)
	}

	return nil
}
