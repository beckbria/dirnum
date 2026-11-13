package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ignoreList []*regexp.Regexp

func wildcardToRegex(pattern string) (*regexp.Regexp, error) {
	escaped := regexp.QuoteMeta(pattern)
	escaped = strings.ReplaceAll(escaped, `\*`, `.*`)
	escaped = strings.ReplaceAll(escaped, `\?`, `.`)
	return regexp.Compile("^" + escaped + "$")
}

func readIgnorePatterns(dir string) (ignoreList, error) {
	f, err := os.Open(filepath.Join(dir, ".ignore"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var patterns ignoreList
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		re, err := wildcardToRegex(line)
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, re)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return patterns, nil
}

func (il ignoreList) shouldIgnore(name string) bool {
	if name == ".ignore" {
		return true
	}
	for _, re := range il {
		if re.MatchString(name) {
			return true
		}
	}
	return false
}

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
	ignores, err := readIgnorePatterns(dir)
	if err != nil {
		return fileNames, err
	}
	for _, f := range files {
		n := f.Name()
		if !ignores.shouldIgnore(n) {
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
