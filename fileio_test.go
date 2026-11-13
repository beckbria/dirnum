package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestFiles(t *testing.T, dir string, names []string) {
	t.Helper()
	for _, name := range names {
		err := os.WriteFile(filepath.Join(dir, name), []byte("test"), 0644)
		require.NoError(t, err)
	}
}

func TestReadFileNamesWithoutIgnoreFile(t *testing.T) {
	dir := t.TempDir()
	createTestFiles(t, dir, []string{
		"0000.jpg",
		"0001.jpg",
		"Thumbs.db",
	})

	got, err := ReadFileNames(dir)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"0000.jpg", "0001.jpg", "Thumbs.db"}, got)
}

func TestReadFileNamesWithIgnoreFile(t *testing.T) {
	dir := t.TempDir()
	ignoreContent := "\n  *.sh  \nfoo*.jpg\nnotes.txt\n\ntemp?.dat\nfile[1]+(test).txt\nThumbs.db\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ignore"), []byte(ignoreContent), 0644))

	createTestFiles(t, dir, []string{
		"0000.jpg",
		"0001.jpg",
		"Thumbs.db",
		"build.sh",
		"a.sh",
		"sh.txt",
		"script.sh.bak",
		"foo.jpg",
		"foobar.jpg",
		"foo-123.jpg",
		"barfoo.jpg",
		"foo.png",
		"notes.txt",
		"temp1.dat",
		"temp12.dat",
		"file[1]+(test).txt",
	})

	got, err := ReadFileNames(dir)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{
		"0000.jpg",
		"0001.jpg",
		"sh.txt",
		"script.sh.bak",
		"barfoo.jpg",
		"foo.png",
		"temp12.dat",
	}, got)
}
