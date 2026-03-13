package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

var ignoreRegEx = regexp.MustCompile(`^Thumbs\.db$`)

func RenameFile(oldName, newName, dirName string) {
	oldPath := filepath.Join(dirName, oldName)
	newPath := filepath.Join(dirName, newName)
	fmt.Printf("Renaming %s to %s\n", oldPath, newPath)
	os.Rename(oldPath, newPath)
}

func ReadFileNames(dir string) ([]string, error) {
	fileNames := make([]string, 0)
	files, err := os.ReadDir(dir)
	if err != nil {
		return fileNames, err
	}
	for _, f := range files {
		n := f.Name()
		if !ignoreRegEx.MatchString(n) {
			fileNames = append(fileNames, n)
		}
	}
	return fileNames, nil
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
