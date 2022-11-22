// The validateAttributes binary compares token attributes stored in a CSV
// file with ones found in a reference JSON file.
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"

	"github.com/proofxyz/solidify/examples/moonbirds/testing/attributes"
)

type arrayFlags []string

func (i arrayFlags) String() string {
	return fmt.Sprint([]string(i))
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type config struct {
	tokenID                            int
	testAttributesCSV, refMetadataJSON string
	ignoreRefAttributes                []string
}

func main() {
	cfg := config{}

	var ignoreRefAttributes arrayFlags
	flag.IntVar(&cfg.tokenID, "tokenID", -1, "The tokenID")
	flag.StringVar(&cfg.testAttributesCSV, "testAttributesCSVPath", "", "The path to the JSON file containing reference metadata")
	flag.StringVar(&cfg.refMetadataJSON, "refMetadataPath", "", "The path to the JSON file containing reference metadata")
	flag.Var(&ignoreRefAttributes, "ignoreRefAttribute", "Ignores the given attribute in the reference file (can be specified multiple times)")
	flag.Parse()

	cfg.ignoreRefAttributes = ignoreRefAttributes

	if err := cfg.validateAttribtues(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("0")
}

func (c *config) validateAttribtues() error {
	got, err := attributes.ReadAttributesCSV(c.testAttributesCSV)
	if err != nil {
		return fmt.Errorf("attributes.ReadAttributesCSV(%q): %w", c.testAttributesCSV, err)

	}

	want, err := attributes.LoadReferenceAttributes(c.refMetadataJSON, c.tokenID, c.ignoreRefAttributes)
	if err != nil {
		return fmt.Errorf("attributes.LoadReferenceAttributes(%q, %d, %v): %w", c.refMetadataJSON, c.tokenID, c.ignoreRefAttributes, err)
	}

	if err := attributes.ValidateAttributes(want, got); err != nil {
		return fmt.Errorf("attributes.ValidateAttributes(%T, %d): %w", got, c.tokenID, err)
	}

	return nil
}
