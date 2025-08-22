package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"p9e.in/ugcl/packages/config"
	"p9e.in/ugcl/packages/p9log"
)

var _ config.Source = (*file)(nil)

type file struct {
	path string
}

// NewSource new a file source.
func NewSource(path string) config.Source {
	return &file{path: path}
}
func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if len(ext) > 0 {
		return ext[1:]
	}
	return ""
}

func (f *file) loadFile(path string) (*config.KeyValue, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	return &config.KeyValue{
		Key:    info.Name(),
		Format: getFileExtension(info.Name()),
		Value:  data,
	}, nil
}

func (f *file) loadDir(path string) (kvs []*config.KeyValue, err error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		// ignore hidden files
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}
		kv, err := f.loadFile(filepath.Join(path, file.Name()))
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, kv)
	}
	return
}

func (f *file) Load() (kvs []*config.KeyValue, err error) {

	fi, err := os.Stat(f.path)
	p9log.Error(err)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return f.loadDir(f.path)
	}
	kv, err := f.loadFile(f.path)
	if err != nil {
		return nil, err
	}
	return []*config.KeyValue{kv}, nil
}

func (f *file) Watch() (config.Watcher, error) {
	return newWatcher(f)
}

// newWatcher creates a new file watcher.
func newWatcher(_ *file) (config.Watcher, error) {
	// Implementation of the watcher goes here.
	// For now, return nil and a not implemented error.
	return nil, fmt.Errorf("newWatcher not implemented")
}
