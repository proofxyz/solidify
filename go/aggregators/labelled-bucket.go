package aggregators

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/proofxyz/solidify/go/deflate"
	"github.com/proofxyz/solidify/go/storage"
)

// LabelledBucket stores fields with a fixed size. Field access/identification
// is achieved by prepending a big-endian uint16 label to each field data blob.
// | label 0 (2 bytes) | blob 0 (N bytes) | label 1 (2 bytes) | blob 1 (N bytes) | ...
type LabelledBucket struct {
	raw       bytes.Buffer
	fields    []storage.LabelledField
	fieldSize uint8
}

// AddField adds a field to the bucket
func (b *LabelledBucket) AddField(f storage.LabelledField) error {
	d, err := f.Encode()
	if err != nil {
		return err
	}

	fieldSize := uint8(len(d))

	if b.fieldSize == 0 {
		b.fieldSize = fieldSize
	}

	if fieldSize != b.fieldSize {
		return fmt.Errorf("all fields need to be of same size after encoding: got %d, want %d", fieldSize, b.fieldSize)
	}

	b.fields = append(b.fields, f)

	if err = binary.Write(&b.raw, binary.BigEndian, f.Label()); err != nil {
		return fmt.Errorf("binary.Write(%T, %v, %v): %w", &b.raw, binary.BigEndian, f.Label(), err)
	}

	b.raw.Write(d)

	return nil
}

// Labels returns the labels of all fields in the bucket
func (b *LabelledBucket) Labels() []uint16 {
	labels := make([]uint16, len(b.fields))
	for i, v := range b.fields {
		labels[i] = v.Label()
	}
	return labels
}

// NumFields returns the number of fields in the bucket
func (b *LabelledBucket) NumFields() int {
	return len(b.fields)
}

// Data returns the encoded and compressed data blob of the bucket
func (b *LabelledBucket) Data() ([]byte, error) {
	d, err := deflate.Deflate(bytes.NewReader(b.raw.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("deflate.Deflate([data]): %w", err)
	}

	return d.Data, nil
}

// UncompressedSize returns the size of uncompressed data in the bucket
func (b *LabelledBucket) UncompressedSize() int {
	return b.raw.Len()
}

// GroupIntoLabelledBuckets groups labelled fields into LabelledBucket by
// limiting the raw data size in each bucket.
func GroupIntoLabelledBuckets[F storage.LabelledField](fs []F, maxBucketSize int) ([]*LabelledBucket, error) {
	var buckets []*LabelledBucket
	b := new(LabelledBucket)

	for _, f := range fs {
		err := b.AddField(f)
		if err != nil {
			return nil, fmt.Errorf("%T.AddField(%v): %w", b, f, err)
		}
		if b.UncompressedSize() > maxBucketSize {
			buckets = append(buckets, b)
			b = new(LabelledBucket)
		}
	}

	if b.UncompressedSize() > 0 {
		buckets = append(buckets, b)
	}

	return buckets, nil
}
