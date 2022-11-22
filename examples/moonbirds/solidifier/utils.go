package main

import (
	"fmt"

	"github.com/proofxyz/solidify/go/storage"
)

func PrintStorageStats[S storage.BucketStorage](stores []S) error {
	var totalBytes int
	for _, s := range stores {
		size, err := s.Size()
		if err != nil {
			return fmt.Errorf("%T.Size(): %w", s, err)
		}
		totalBytes += size
		fmt.Printf("%s: %d B\n", s.Name(), size)
	}
	fmt.Printf("Total size: %d B\n\n", totalBytes)
	return nil
}
