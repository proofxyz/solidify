// Package storage defines structures for data storage and organistion in
// Ethereum smart contracts together with routines to generate them.
package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	_ "embed"
)

var (
	tmplFuncsCommon = template.FuncMap{
		"hex": func(b []byte) string {
			return fmt.Sprintf(`hex"%x"`, b)
		},
		"numFields": func(s interface{ NumFields() int }) int {
			return s.NumFields()
		},
		"numFieldsPerBucketHex": func(s BucketStorage) string {
			nums := []byte{}
			for _, b := range s.Buckets() {
				nums = append(nums, uint8(b.NumFields()))
			}
			return fmt.Sprintf(`hex"%x"`, nums)
		},
		"printUnlessFirstCall": func(s string) func() string {
			i := -1
			return func() string {
				i++
				if i == 0 {
					return ""
				}
				return s
			}
		},
		"toLower": strings.ToLower,
		"numBits": func(x int) int {
			if x > 255 {
				return 16
			}
			return 8
		},
	}

	//go:embed templates/bucket-storage.go.tmpl
	rawBucketStorageTmpl string

	bucketStorageTmpl = template.Must(
		template.New("bucket-storage").Funcs(tmplFuncsCommon).Parse(rawBucketStorageTmpl),
	)

	//go:embed templates/storage-deployer.go.tmpl
	rawDeployerTmpl string

	deployerTmpl = template.Must(
		template.New("storage-deployer").Funcs(tmplFuncsCommon).Parse(rawDeployerTmpl),
	)
)

// A Field represents arbitrary data that can be represented in a binary format.
type Field interface {
	Encode() ([]byte, error)
}

// A Bucket aggregates and compresses a list of Fields, adding additional indexing metadata.
type Bucket interface {
	Data() ([]byte, error)
	UncompressedSize() int
	NumFields() int
}

// A BucketStorage is a named list of Buckets that will mapped to a single contract file.
type BucketStorage interface {
	Name() string
	Buckets() []Bucket
	NumFields() int
	Size() (int, error)
}

// WriteBucketStorage writes the storage contract file for a given BucketStorage.
func WriteBucketStorage(s BucketStorage, w io.Writer) error {
	return bucketStorageTmpl.Execute(w, struct{ Store BucketStorage }{s})
}

// WriteStorageDeployer writes a helper contract to deploy a set of BucketStorage contracts located at storagePath.
func WriteStorageDeployer[S BucketStorage](name string, storagePath string, stores []S, w io.Writer) error {
	return deployerTmpl.Execute(w,
		struct {
			Name        string
			StoragePath string
			Stores      []BucketStorage
		}{
			Name:        name,
			StoragePath: storagePath,
			Stores:      convertStorages(stores),
		})
}

func convertStorages[S BucketStorage](s []S) []BucketStorage {
	ss := make([]BucketStorage, len(s))
	for i, v := range s {
		ss[i] = v
	}
	return ss
}

func addTemplateFuncs(m1, m2 template.FuncMap) template.FuncMap {
	ret := make(template.FuncMap)
	for t, v := range m1 {
		ret[t] = v
	}
	for t, v := range m2 {
		ret[t] = v
	}
	return ret
}

func createFile(outputDir, name string) (*os.File, error) {
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("os.Mkdir(%q): %w", outputDir, err)
	}

	path := filepath.Join(outputDir, fmt.Sprintf("%s.sol", name))
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("os.Create(%s): %w", path, err)
	}

	return f, nil
}

// A fileGenerator writes files and tracks the paths of the files it creates.
type fileGenerator struct {
	created []string
}

// writeSolFile creates a new file <dir>/<name>.sol and passes it to the write()
// callback for generation. Any error returned by write() will be propagated.
func (g *fileGenerator) writeSolFile(dir, name string, write func(*os.File) error) (retErr error) {
	f, err := createFile(dir, name)
	if err != nil {
		return fmt.Errorf("createFile(%q, %q): %v", dir, name, err)
	}
	defer func() {
		if err := f.Close(); retErr == nil {
			retErr = err
		}
	}()

	if err := write(f); err != nil {
		return err
	}
	g.created = append(g.created, f.Name())
	return nil
}

// annotateNonNil returns nil if err == nil, otherwise it annotates the error
// with the specified prefix.
func annotateNonNil(err error, prefixFormat string, a ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(prefixFormat+": %v", append(a, err)...)
}
