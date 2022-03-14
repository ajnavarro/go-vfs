package vfs

import (
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestRecursiveFS(t *testing.T) {
	require := require.New(t)

	dirFS := os.DirFS("./testdata")

	var factories = map[string]FSFactory{
		".zip": ZipFactory,
	}

	recursive := NewRecursive(dirFS, factories)

	err := fstest.TestFS(recursive,
		"sample.zip/testdata/quote1.txt",
		"sample.zip/testdata/already-compressed.jpg",
		"sample.zip/testdata/proverbs",
		"sample.zip/testdata/quote1.txt",
		"sample.zip/testdata/already-compressed.jpg",
		"sample.zip/testdata/proverbs",
	)

	require.NoError(err)
}
