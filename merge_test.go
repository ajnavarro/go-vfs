package vfs

import (
	"archive/zip"
	"io"
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestMergeFS(t *testing.T) {
	require := require.New(t)

	zip, err := zip.OpenReader("./testdata/sample.zip")
	require.NoError(err)

	dirFS := os.DirFS("./testdata")

	merge := NewMerge()

	merge.Add("id1", zip)
	merge.Add("id2", dirFS)

	err = fstest.TestFS(merge,
		"id1/testdata/quote1.txt",
		"id1/testdata/already-compressed.jpg",
		"id1/testdata/proverbs",
		"id2/sample.zip",
	)

	require.NoError(err)

	f1, err := merge.Open("id2/sample.zip")
	require.NoError(err)

	_, ok := f1.(io.ReaderAt)
	require.True(ok, "file does not implement ReaderAt")

	_, ok = f1.(io.Seeker)
	require.True(ok, "file does not implement Seeker")

	ok = merge.Remove("id2")
	require.True(ok, "error removing filesystem")

	err = fstest.TestFS(merge,
		"id1/testdata/quote1.txt",
		"id1/testdata/already-compressed.jpg",
		"id1/testdata/proverbs",
	)

	require.NoError(err)
}
