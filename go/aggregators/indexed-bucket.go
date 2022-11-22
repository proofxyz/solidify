package aggregators

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/proofxyz/solidify/go/deflate"
	"github.com/proofxyz/solidify/go/storage"
)

// IndexedBucket stores fields and provides a access by prepending a location header.
// The offset of the start of each field in the data blob is stored as big-endian
// uint16 in the index header in the order of the fields.
// | offset field 0 (2 bytes) | ... | offset n-1 | blob field 0 (n bytes) | ... |
// ______________________________________________^ offset 0
type IndexedBucket struct {
	payload    []byte
	fieldSizes []uint16
	fields     []storage.Field
}

// AddField adds a field to the bucket
func (b *IndexedBucket) AddField(f storage.Field) error {
	d, err := f.Encode()
	if err != nil {
		return fmt.Errorf("%T.Encode(): %w", f, err)
	}

	b.fields = append(b.fields, f)
	b.payload = append(b.payload, d...)
	b.fieldSizes = append(b.fieldSizes, uint16(len(d)))

	return nil
}

// UncompressedSize returns the size of uncompressed data in the bucket
func (b *IndexedBucket) UncompressedSize() int {
	return len(b.payload) + len(b.fields)*2
}

// NumFields returns the number of fields in the bucket
func (b *IndexedBucket) NumFields() int {
	return len(b.fields)
}

// Data returns the encoded and compressed data blob of the bucket
func (b *IndexedBucket) Data() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.Grow(len(b.fields)*2 + len(b.payload))

	// Build index header
	t := uint16(len(b.fields)) * 2
	for _, v := range b.fieldSizes {
		if err := binary.Write(buf, binary.BigEndian, t); err != nil {
			return nil, fmt.Errorf("binary.Write(%T, %v, %v): %w", buf, binary.BigEndian, t, err)
		}
		t += v
	}

	// Append data
	if _, err := buf.Write(b.payload); err != nil {
		return nil, fmt.Errorf("%T.Write(%T): %w", buf, b.payload, err)
	}

	d, err := deflate.Deflate(buf)
	if err != nil {
		return nil, fmt.Errorf("deflate.Deflate([data]): %w", err)
	}

	return d.Data, nil
}

// GroupIntoIndexedBuckets groups fields into IndexedBuckets by limiting the
// raw data size in each bucket.
func GroupIntoIndexedBuckets[F storage.Field](fs []F, maxBucketSize int) ([]*IndexedBucket, error) {
	var buckets []*IndexedBucket
	b := new(IndexedBucket)

	for _, f := range fs {
		if err := b.AddField(f); err != nil {
			return nil, fmt.Errorf("%T.AddField(%v): %w", b, f, err)
		}

		if b.UncompressedSize() > maxBucketSize {
			buckets = append(buckets, b)
			b = new(IndexedBucket)
		}
	}

	if b.UncompressedSize() > 0 {
		buckets = append(buckets, b)
	}

	return buckets, nil
}
