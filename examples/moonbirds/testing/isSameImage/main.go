// The isSameImage binary decodes two given image files and checks if they are
// pixel-wise identical.
package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/proofxyz/solidify/examples/moonbirds/testing/image"
	"github.com/proofxyz/solidify/go/types"
)

func main() {
	code, err := run(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(code)
}

func run(path1, path2 string) (int, error) {
	same, err := isSameImage(path1, path2)
	if err != nil {
		return 1, fmt.Errorf("isSameImage(%q, %q): %w", path1, path2, err)
	}

	if !same {
		return 1, nil
	}

	return 0, nil
}

func isSameImage(path1, path2 string) (bool, error) {
	im1, err := types.LoadImage(path1)
	if err != nil {
		return false, fmt.Errorf("types.LoadImage(%q): %w", path1, err)
	}

	im2, err := types.LoadImage(path2)
	if err != nil {
		return false, fmt.Errorf("types.LoadImage(%q): %w", path2, err)
	}

	diff, err := image.CompareImages(im1, im2)
	if err != nil {
		return false, fmt.Errorf("image.CompareImages(%T, %T): %w", im1, im2, err)
	}

	return diff == 0, nil
}
