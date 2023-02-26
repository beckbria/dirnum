package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func RenameFile(oldName, newName string, dirName *string) {
	oldPath := *dirName + string(os.PathSeparator) + oldName
	newPath := *dirName + string(os.PathSeparator) + newName
	fmt.Printf("Renaming %s to %s\n", oldPath, newPath)
	os.Rename(oldPath, newPath)
}

func ReadFileNames(dir string) ([]string, error) {
	fileNames := make([]string, 0)
	files, err := ioutil.ReadDir(dir)
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