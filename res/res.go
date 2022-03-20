package res

import (
	"bytes"
	"errors"
	"net/http"
	"os"
	"time"
)

var FileSystem http.FileSystem

var r *map[string][]byte

func init() {
	if r != nil {
		FileSystem = FakeFileSystem{*r}
	} else {
		FileSystem = http.Dir("html")
	}
}

func IsCacheTemplate() bool {
	return r != nil
}

var modtime = time.Now()
var errNotDir = errors.New("not a folder")

// MemoryFile Read R file
type MemoryFile struct {
	*bytes.Reader
	size  int64
	name  string
	isDir bool
}

func (m *MemoryFile) Close() error {
	return nil
}
func (m *MemoryFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, errNotDir
}
func (m *MemoryFile) Stat() (os.FileInfo, error) {
	return m, nil
}
func (m *MemoryFile) Name() string {
	return m.name
}
func (m *MemoryFile) Size() int64 {
	return m.size
}
func (m *MemoryFile) Mode() os.FileMode {
	return os.ModePerm
}
func (m *MemoryFile) ModTime() time.Time {
	return modtime
}
func (m *MemoryFile) IsDir() bool {
	return m.isDir
}
func (m *MemoryFile) Sys() interface{} {
	return nil
}

// FakeFileSystem Read R file system
type FakeFileSystem struct {
	R map[string][]byte
}

// Open open resources by R
func (ffs FakeFileSystem) Open(name string) (http.File, error) {
	if data, ok := ffs.R[name]; ok {
		if data != nil {
			return &MemoryFile{bytes.NewReader(data), int64(len(data)), name, false}, nil
		}
		return &MemoryFile{nil, 0, name, true}, nil
	}
	return nil, os.ErrNotExist
}
