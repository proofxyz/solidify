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

const genDst = "./gen/"

type testDataGroup struct {
	name   string
	values []types.StringField
}

func (t testDataGroup) Name() string {
	return t.name
}

func (t testDataGroup) NumFields() int {
	return len(t.values)
}

func run() error {
	gs := []testDataGroup{
		{
			name:   "FOO",
			values: []types.StringField{"foo0", "foo1"},
		},
		{
			name:   "BAR",
			values: []types.StringField{"bar0", "bar1", "bar2"},
		},
		{
			name:   "QUX",
			values: []types.StringField{"qux0"},
		},
	}

	addToStorage := func(s *aggregators.BucketStorage, gs []testDataGroup) error {
		for _, g := range gs {
			b := new(aggregators.IndexedBucket)
			for _, f := range g.values {
				err := b.AddField(f)
				if err != nil {
					return fmt.Errorf("%T,AddField(%T): %w", b, f, err)
				}

			}
			s.AddBucket(b)
		}
		return nil
	}

	ss := []*aggregators.BucketStorage{
		aggregators.NewBucketStorage("GroupStorage0"),
		aggregators.NewBucketStorage("GroupStorageB"),
	}
	if err := addToStorage(ss[0], gs[:2]); err != nil {
		return err
	}
	if err := addToStorage(ss[1], gs[2:]); err != nil {
		return err
	}

	fNames, err := storage.WriteGroupStorage("GroupStorage", gs, ss, genDst)
	if err != nil {
		return fmt.Errorf("storage.WriteGroupStorage(%q, %T, %T, %q): %w", "Group", gs, ss, genDst, err)
	}

	if err := storage.FormatSol(fNames); err != nil {
		return fmt.Errorf("storage.FormatSol(%v): %w", fNames, err)
	}

	return nil
}
