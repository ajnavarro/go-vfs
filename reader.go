package vfs

import (
	"io"
	"io/ioutil"
	"os"
	"sync"
)

type diskTeeReader struct {
	io.ReaderAt
	io.Closer
	io.Reader

	m sync.Mutex

	fo int64
	fr *os.File
	to int64
	tr io.Reader
}

func newDiskTeeReader(r io.Reader, tmp string) (*diskTeeReader, error) {
	fr, err := ioutil.TempFile(tmp, "dtr_tmp")
	if err != nil {
		return nil, err
	}
	tr := io.TeeReader(r, fr)
	return &diskTeeReader{fr: fr, tr: tr}, nil
}

func (dtr *diskTeeReader) ReadAt(p []byte, off int64) (int, error) {
	dtr.m.Lock()
	defer dtr.m.Unlock()
	tb := off + int64(len(p))

	if tb > dtr.fo {
		w, err := io.CopyN(ioutil.Discard, dtr.tr, tb-dtr.fo)
		dtr.to += w
		if err != nil && err != io.EOF {
			return 0, err
		}
	}

	n, err := dtr.fr.ReadAt(p, off)
	dtr.fo += int64(n)
	return n, err
}

func (dtr *diskTeeReader) Read(p []byte) (n int, err error) {
	dtr.m.Lock()
	defer dtr.m.Unlock()
	// use directly tee reader here
	n, err = dtr.tr.Read(p)
	dtr.to += int64(n)
	return
}

func (dtr *diskTeeReader) Close() error {
	if err := dtr.fr.Close(); err != nil {
		return err
	}

	return os.Remove(dtr.fr.Name())
}
