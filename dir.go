package template

import (
	"os"
	"path"
)

// Dir implements FileSystem based on native filesystem.
type Dir string

func (d Dir) Open(filename string) (File, error) {
	fullPath := path.Join(string(d), filename)

	return os.Open(fullPath)
}
