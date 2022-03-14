package vfs

import (
	"errors"
	"io/fs"
	"path"
	"strings"
	"sync"
)

var ErrFSNotFound = errors.New("FS not found")

type FSFactory func(f fs.File) (fs.FS, error)

var _ fs.FS = &Recursive{}

type Recursive struct {
	root      fs.FS
	factories map[string]FSFactory

	mu        sync.RWMutex
	instances map[string]fs.FS
}

func NewRecursive(root fs.FS, fsFactories map[string]FSFactory) *Recursive {
	return &Recursive{
		root:      root,
		factories: fsFactories,
		instances: make(map[string]fs.FS),
	}
}

func (vfs *Recursive) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	f, err := vfs.parseElements(name)
	if err != nil {
		return nil, err
	}

	return &recursiveFile{
		factories: vfs.factories,
		name:      name,
		File:      f,
	}, nil
}

func (vfs *Recursive) parseElements(name string) (fs.File, error) {
	elementKeys := strings.Split(name, "/")

	actualFS := vfs.root
	key := "."

	if len(elementKeys) == 0 {
		return actualFS.Open(name)
	}

	for _, ekey := range elementKeys {
		key = path.Join(key, ekey)

		if !isChildFS(key, vfs.factories) {
			continue
		}

		f, err := actualFS.Open(key)
		if err != nil {
			return nil, err
		}

		childFS, err := vfs.getFS(key, f)
		if err != nil {
			return nil, err
		}

		actualFS = childFS
		key = "."

	}

	return actualFS.Open(key)
}

func (vfs *Recursive) getFS(name string, f fs.File) (fs.FS, error) {
	vfs.mu.RLock()
	fs, ok := vfs.instances[name]
	vfs.mu.RUnlock()

	if ok {
		return fs, nil
	}

	ext := path.Ext(name)
	factory, ok := vfs.factories[ext]

	if !ok {
		return nil, ErrFSNotFound
	}

	fs, err := factory(f)
	if err != nil {
		return nil, err
	}

	vfs.mu.Lock()
	vfs.instances[name] = fs
	vfs.mu.Unlock()

	return fs, nil
}

func isChildFS(name string, factories map[string]FSFactory) bool {
	ext := path.Ext(name)
	_, ok := factories[ext]
	return ok
}

var _ fs.File = &recursiveFile{}
var _ fs.ReadDirFile = &recursiveFile{}

type recursiveFile struct {
	name string
	fs.File

	factories map[string]FSFactory
}

func (f *recursiveFile) Stat() (fs.FileInfo, error) {
	stat, err := f.File.Stat()
	if err != nil {
		return nil, err
	}

	return newNameableFileInfo(f.name, stat), nil
}

func (f *recursiveFile) ReadDir(n int) ([]fs.DirEntry, error) {
	dir, ok := f.File.(fs.ReadDirFile)
	if !ok {
		return nil, &fs.PathError{Op: "readdir", Path: f.name, Err: errors.New("not a directory")}
	}

	entries, err := dir.ReadDir(n)
	if err != nil {
		return nil, err
	}

	var out []fs.DirEntry
	// if we are returning a folder, we need to check
	// if some of the content might be a file from factories
	for _, e := range entries {
		if !isChildFS(e.Name(), f.factories) {
			out = append(out, e)
			continue
		}

		newEntry := newDirEntry(newDirFileInfo(e.Name()))
		out = append(out, newEntry)
	}

	return out, nil
}
