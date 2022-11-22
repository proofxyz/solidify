// Package aggregators provides data aggregation and indexing for the solidify
// toolchain.
// The Ethereum smart contracts generated with the contained types are intended
// to be used with the contracts provided under https://github.com/proofxyz/solidify/contracts.
package aggregators

import (
	"fmt"

	"github.com/proofxyz/solidify/go/storage"
)

// BucketStorage is an implementation of storage.BucketStorage, storing
// a slice of buckets.
type BucketStorage struct {
	name    string
	buckets []storage.Bucket
}

// NewBucketStorage creates a new BucketStorage with given name.
func NewBucketStorage(name string) *BucketStorage {
	return &BucketStorage{name: name}
}

// AddBucket adds a bucket to the storage.
func (s *BucketStorage) AddBucket(b storage.Bucket) {
	s.buckets = append(s.buckets, b)
}

// Buckets returns the buckets in this storage.
func (s *BucketStorage) Buckets() []storage.Bucket {
	return s.buckets
}

// Name returns the name of the storage.
func (s *BucketStorage) Name() string {
	return s.name
}

// NumFields returns the number of fields in the storage.
func (s *BucketStorage) NumFields() int {
	var sum int
	for _, b := range s.buckets {
		sum += b.NumFields()
	}
	return sum
}

// Size returns the total size of encoded buckets in the storage .
func (s *BucketStorage) Size() (int, error) {
	var t int
	for _, b := range s.buckets {
		d, err := b.Data()
		if err != nil {
			return 0, fmt.Errorf("%T.Data(): %w", b, err)
		}
		t += len(d)
	}
	return t, nil
}

// GroupIntoStorages groups buckets into storages by limiting the max number of
// buckets in and total size of each storage.
func GroupIntoStorages[B storage.Bucket](buckets []B, maxStorageSize, maxBuckets int, baseName string) ([]*BucketStorage, error) {
	var stores []*BucketStorage
	storeBuf := new(BucketStorage)

	pushStorage := func(b *BucketStorage) {
		b.name = fmt.Sprintf("%sBucketStorage%d", baseName, len(stores))
		stores = append(stores, b)
	}

	for _, b := range buckets {
		storeBuf.buckets = append(storeBuf.buckets, b)

		s, err := storeBuf.Size()
		if err != nil {
			return nil, fmt.Errorf("%T.Size(): %w", storeBuf, err)
		}

		if (maxStorageSize >= 0 && s > maxStorageSize) ||
			(maxBuckets >= 0 &&
				len(storeBuf.buckets) >= maxBuckets) {
			pushStorage(storeBuf)
			storeBuf = new(BucketStorage)
		}
	}

	if len(storeBuf.buckets) > 0 {
		pushStorage(storeBuf)
	}

	return stores, nil
}
