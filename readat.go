package vfs

import (
	"io"
	"io/fs"
)

var _ fs.FS = &ReadAt{}

type ReadAt struct {
	ifs       fs.FS
	tmpFolder string
}

func NewReadAt(fs fs.FS, tmp string) *ReadAt {
	return &ReadAt{
		ifs:       fs,
		tmpFolder: tmp,
	}
}

func (vfs *ReadAt) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	f, err := vfs.ifs.Open(name)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		return f, nil
	}

	_, ok := f.(io.ReaderAt)
	if ok {
		return f, nil
	}

	dtr, err := newDiskTeeReader(f, vfs.tmpFolder)
	if err != nil {
		return nil, err
	}

	return &readAtFile{
		File: f,
		dtr:  dtr,
	}, nil
}

type readAtFile struct {
	fs.File
	dtr *diskTeeReader
}

func (raf *readAtFile) Read(p []byte) (int, error) {
	return raf.dtr.Read(p)
}

func (raf *readAtFile) ReadAt(p []byte, off int64) (int, error) {
	return raf.dtr.ReadAt(p, off)
}

func (raf *readAtFile) Close() error {
	return raf.dtr.Close()
}
