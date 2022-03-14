package vfs

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"sync"
)

type Merge struct {
	mu  sync.RWMutex
	fss []fsWithKey
}

type fsWithKey struct {
	key string
	fs  fs.FS
}

func NewMerge() *Merge {
	return &Merge{}
}

func (vfs *Merge) Add(key string, filesystem fs.FS) {
	vfs.mu.Lock()
	defer vfs.mu.Unlock()

	vfs.fss = append(vfs.fss, fsWithKey{key: key, fs: filesystem})
}

func (vfs *Merge) Remove(key string) bool {
	vfs.mu.Lock()
	defer vfs.mu.Unlock()

	for i, v := range vfs.fss {
		if v.key == key {
			// re-slice
			vfs.fss = append(vfs.fss[:i], vfs.fss[i+1:]...)
			return true
		}
	}

	return false
}

func (vfs *Merge) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	vfs.mu.RLock()
	defer vfs.mu.RUnlock()

	//Special case for root path
	if name == "." {
		return vfs.createRootDir()
	}

	for _, fsk := range vfs.fss {
		if name == fsk.key {
			fsr, err := fsk.fs.Open(".")
			if err != nil {
				return nil, err
			}
			rdf, ok := fsr.(fs.ReadDirFile)
			if !ok {
				return nil, &fs.PathError{Op: "open", Path: name, Err: errors.New("not a directory")}
			}

			return &rootFile{name: name, ReadDirFile: rdf}, nil
		}

		fsName := strings.TrimLeft(name, fsk.key+"/")

		f, err := fsk.fs.Open(fsName)
		if os.IsNotExist(err) {
			continue
		}

		return f, err
	}

	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

func (vfs *Merge) createRootDir() (fs.File, error) {
	var dirs []fs.DirEntry
	for _, fsk := range vfs.fss {
		stat, err := fs.Stat(fsk.fs, ".")
		if err != nil {
			return nil, err
		}

		dirs = append(dirs, newDirEntry(newNameableFileInfo(fsk.key, stat)))
	}

	return newDir(newDirFileInfo("."), dirs), nil
}

var _ fs.File = &rootFile{}

type rootFile struct {
	fs.ReadDirFile
	name string
}

func (f *rootFile) Stat() (fs.FileInfo, error) {
	stat, err := f.ReadDirFile.Stat()
	if err != nil {
		return nil, err
	}

	return newNameableFileInfo(f.name, stat), nil
}
