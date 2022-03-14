package vfs

import (
	"io"
	"io/fs"
	"path"
	"time"
)

var _ fs.ReadDirFile = &dir{}

func newDir(fi fs.FileInfo, dirs []fs.DirEntry) *dir {
	return &dir{
		fi:   fi,
		dirs: dirs,
	}
}

type dir struct {
	fi fs.FileInfo

	offset int
	dirs   []fs.DirEntry
}

func (d *dir) Stat() (fs.FileInfo, error) {
	return d.fi, nil
}

func (d *dir) Read([]byte) (int, error) {
	return 0, nil
}

func (d *dir) Close() error {
	return nil
}

func (d *dir) ReadDir(count int) ([]fs.DirEntry, error) {
	n := len(d.dirs) - d.offset
	if count > 0 && n > count {
		n = count
	}
	if n == 0 {
		if count <= 0 {
			return nil, nil
		}
		return nil, io.EOF
	}

	list := make([]fs.DirEntry, n)
	for i := range list {
		list[i] = d.dirs[d.offset+i]
	}
	d.offset += n

	return list, nil
}

var _ fs.DirEntry = &dirEntry{}

type dirEntry struct {
	fi fs.FileInfo
}

func newDirEntry(fi fs.FileInfo) *dirEntry {
	return &dirEntry{
		fi: fi,
	}
}

func (de *dirEntry) Name() string {
	return de.fi.Name()
}

func (de *dirEntry) IsDir() bool {
	return de.fi.IsDir()
}

func (de *dirEntry) Type() fs.FileMode {
	return de.fi.Mode().Type()
}

func (de *dirEntry) Info() (fs.FileInfo, error) {
	return de.fi, nil
}

func newNameableFileInfo(name string, fi fs.FileInfo) *nameableFileInfo {
	return &nameableFileInfo{
		name:     name,
		FileInfo: fi,
	}
}

type nameableFileInfo struct {
	fs.FileInfo
	name string
}

func (fi *nameableFileInfo) Name() string {
	base := path.Base(fi.name)
	return base
}

var _ fs.FileInfo = &dirFileInfo{}

func newDirFileInfo(name string) *dirFileInfo {
	return &dirFileInfo{name: name}
}

type dirFileInfo struct {
	name string
}

func (fi *dirFileInfo) Path() string {
	return fi.name
}

func (fi *dirFileInfo) Name() string {
	return path.Base(fi.name)
}
func (fi *dirFileInfo) Size() int64 {
	return 0
}
func (fi *dirFileInfo) Mode() fs.FileMode {
	return fs.ModeDir | 0555
}
func (fi *dirFileInfo) ModTime() time.Time {
	return time.Time{}
}
func (fi *dirFileInfo) IsDir() bool {
	return true
}
func (fi *dirFileInfo) Sys() interface{} {
	return nil
}
