package main

import (
	"fmt"
	"strconv"
	"strings"
)

const NoMinorVersion = -99 // Indicates a file with no minor version

type FileNamePieces struct {
	major, minor, majorDigits, minorDigits int
	originalName, descriptor, extension    string
}

func (f *FileNamePieces) String() string {
	var b strings.Builder
	b.Grow(len(f.originalName))
	b.WriteString(prependZeroes(strconv.Itoa(f.major), f.majorDigits)) // Major version
	if f.minor != NoMinorVersion {
		b.WriteRune('-')
		b.WriteString(prependZeroes(strconv.Itoa(f.minor), f.minorDigits))
	}
	if len(f.descriptor) > 0 {
		// The descriptor includes the leading dash
		b.WriteString(f.descriptor)
	}
	b.WriteRune('.')
	b.WriteString(f.extension)
	return b.String()
}

func prependZeroes(n string, l int) string {
	for len(n) < l {
		n = "0" + n
	}
	return n
}

func ParseFileName(f string) (*FileNamePieces, error) {
	tokens := fileRegEx.FindStringSubmatch(f)
	if tokens == nil {
		return nil, fmt.Errorf("Bad filename: %s", f)
	}
	major, err := strconv.Atoi(tokens[1])
	if err != nil {
		return nil, fmt.Errorf("Invalid major version \"%s\": %s", tokens[1], f)
	}
	minor := NoMinorVersion
	if len(tokens[2]) > 0 {
		minorStr := string([]rune(tokens[2])[1:])
		m, err := strconv.Atoi(minorStr)
		if err != nil {
			return nil, fmt.Errorf("Invalid minor version \"%s\": %s", minorStr, f)
		}
		minor = m
	}
	minorDigits := 0
	if minor != NoMinorVersion {
		minorDigits = len(strconv.Itoa(minor))
	}
	name := FileNamePieces{
		major:        major,
		minor:        minor,
		majorDigits:  len(strconv.Itoa(major)),
		minorDigits:  minorDigits,
		descriptor:   tokens[3],
		extension:    tokens[4],
		originalName: f,
	}
	return &name, nil
}

// PFnpSlice represents a set of file names that can be sorted by major+minor version
type PFnpSlice []*FileNamePieces

func (s PFnpSlice) Len() int {
	return len(s)
}

func (s PFnpSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s PFnpSlice) Less(i, j int) bool {
	first := s[i]
	second := s[j]

	if first.major < second.major {
		return true
	} else if first.major > second.major {
		return false
	}
	// Major version is the same, compare minor version
	return first.minor < second.minor
}
