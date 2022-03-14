package vfs

import (
	"archive/zip"
	"io"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestReadAtFS(t *testing.T) {
	require := require.New(t)

	zip, err := zip.OpenReader("./testdata/sample.zip")
	require.NoError(err)

	readat := NewReadAt(zip, "")

	err = fstest.TestFS(readat,
		"testdata/quote1.txt",
		"testdata/already-compressed.jpg",
		"testdata/proverbs",
	)
	require.NoError(err)

	zf, err := zip.Open("testdata/quote1.txt")
	require.NoError(err)

	_, ok := zf.(io.ReaderAt)
	require.False(ok, "zip should not implement ReaderAt")

	raf, err := readat.Open("testdata/quote1.txt")
	require.NoError(err)

	ra, ok := raf.(io.ReaderAt)
	require.True(ok, "ReaderAt must be present")

	c, ok := raf.(io.Closer)
	require.True(ok, "Closer must be present")
	defer c.Close()

	data := make([]byte, 7)
	n, err := ra.ReadAt(data, 8)
	require.NoError(err)
	require.Equal(7, n)
	require.Equal("generic", string(data))

}
