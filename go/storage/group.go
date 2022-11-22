package storage

import (
	"io"
	"os"
	"path/filepath"
	"text/template"

	"go.uber.org/multierr"

	_ "embed"
)

var (
	tmplFuncsGroups = addTemplateFuncs(tmplFuncsCommon, template.FuncMap{
		"bitsNumFields": func(stores []BucketStorage) int {
			for _, s := range stores {
				if s.NumFields() > 255 {
					return 16
				}
			}
			return 8
		},
	})

	//go:embed templates/sequential-storage-mapping.go.tmpl
	rawSequentialStorageMappingTmpl string

	sequentialStorageMappingTmpl = template.Must(
		template.New("sequential-storage-mapping").Funcs(tmplFuncsGroups).Parse(rawSequentialStorageMappingTmpl),
	)
)

// FieldsGroup is a generic grouping of fields (e.g. all layers with a certain
// layer type)
type FieldsGroup interface {
	Name() string
	NumFields() int
}

// WriteSequentialStorageMapping writes the storage mapping library translating between
// groups of fields (e.g. grouped by layer type) and storage coordinates sequentially.
// For example, the sequential association works like this
// FieldGroups
// |                                   BucketStorage 0
// ├── BAR                  Bucket 0 ──┤
// │   └── bax <> Field 0 ──┤          |
// │                        │          |
// └── FOO                  │          |
//     ├── bar <> Field 1 ──┘          |
//     │                    Bucket 1 ──┘
//     ├── foo <> Field 0 ──┘
//     |                               BucketStorage 1
//     │                    Bucket 0 ──┘
//     └── qux <> Field 0 ──┘
func WriteSequentialStorageMapping[G FieldsGroup, S BucketStorage](name string, groups []G, stores []S, w io.Writer) error {
	return sequentialStorageMappingTmpl.Execute(w,
		struct {
			Name         string
			FieldsGroups []G
			Stores       []BucketStorage
		}{
			Name:         name,
			FieldsGroups: groups,
			Stores:       convertStorages(stores),
		},
	)
}

// WriteGroupStorage is a convenience wrapper that writes all contracts relating
// to a grouping of fields and corresponding BucketStorages to a given output
// directory. Returns the paths of then written files.
func WriteGroupStorage[G FieldsGroup, S BucketStorage](name string, groups []G, stores []S, outputDir string) ([]string, error) {
	var fs fileGenerator
	storageSubdir := "storage"

	errs := []error{
		fs.writeSolFile(outputDir, name+"StorageDeployer", func(f *os.File) error {
			return annotateNonNil(WriteStorageDeployer(name, "./"+storageSubdir, stores, f), "storage.WriteStorageDeployer(%q, …)", name)
		}),
		fs.writeSolFile(outputDir, name+"StorageMapping", func(f *os.File) error {
			return annotateNonNil(WriteSequentialStorageMapping(name, groups, stores, f), "storage.WriteSequentialStorageMapping(%q, …)", name)
		}),
	}
	if err := multierr.Combine(errs...); err != nil {
		return nil, err
	}

	for _, s := range stores {
		if err := fs.writeSolFile(filepath.Join(outputDir, storageSubdir), s.Name(), func(f *os.File) error {
			return annotateNonNil(WriteBucketStorage(s, f), "storage.WriteBucketStorage(…)")
		}); err != nil {
			return nil, err
		}
	}

	return fs.created, nil
}
