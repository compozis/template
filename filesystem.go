package template

import "io"

type File interface {
	io.Closer
	io.Reader
}

type FileSystem interface {
	Open(name string) (File, error)
}

