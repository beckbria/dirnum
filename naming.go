package main

import (
	"sort"
)

type RenameEntry struct {
	oldName, newName string
}

func ComputeRenames(fileNames []string, unused []int) []RenameEntry {
	files := parseFileNames(fileNames)
	renumberMinorVersions(files)

	// Fill in gaps in major numbers.

	// First, backtrack to determine how many entries we need to fill
	majorIdx := len(files) - 1
	unusedIdx := 0
	for ; unusedIdx < len(unused) && majorIdx > 0; unusedIdx++ {
		if unused[unusedIdx] > files[majorIdx].major {
			// We've filled in to a continuous loop
			break
		}
		majorIdx--
	}

	// Now rename files in order
	for len(unused) > 0 && majorIdx < len(files) {
		firstUnused := unused[0]
		unused = unused[1:]

		// Change the major version to the unused value
		for oldMajor := files[majorIdx].major; majorIdx < len(files) && files[majorIdx].major == oldMajor; majorIdx++ {
			files[majorIdx].major = firstUnused
		}
	}

	return changedNames(files)
}

func parseFileNames(fileNames []string) PFnpSlice {
	files := make(PFnpSlice, 0)
	for _, f := range fileNames {
		n, err := ParseFileName(f)
		if err == nil {
			// Don't try to rename files which aren't named correctly.  Errors are displayed
			// before this function and controlled by the quiet flag.
			files = append(files, n)
		}
	}

	// Sort the list by major/minor version
	sort.Sort(files)
	return files
}

// Computes the number of digits required by the major/minor version. That is, if the largest major version is 100, 3 digits are required
// to represent the major version (in base 10).  For each distinct major version, the number of digits required to represent the minor
// version are computed.  Thus, "0-0", "0-1", "1-0", "1-1", ..., "1-10" would return [0: 1, 1: 2] because major version 1 requires one
// digit to represent the minor version while major version 2 requires 2.
//
// We intentionally ignore the edge case where filling the gaps will reduce the number of digits required - if so, the extra digit
// will likely be required soon enough.  If it's particularly important, running the tool a second time will remove the extra digit.
func computeDigitCounts(files PFnpSlice) (int, map[int]int) {
	majorDigits := 0
	minorDigits := make(map[int]int)
	for _, f := range files {
		majorDigits = max(majorDigits, f.majorDigits)
		minorDigits[f.major] = max(minorDigits[f.major], f.minorDigits)
	}
	return majorDigits, minorDigits
}

// Renumbers the minor version of all files.  If only one file exists for a given major version, then the minor version is cleared.
// If multiple exist, they are numbered starting at 0.
func renumberMinorVersions(files PFnpSlice) {
	majorDigits, minorDigits := computeDigitCounts(files)

	previousMajor := NoVersion
	for i, f := range files {
		f.minorDigits = minorDigits[f.major]
		f.majorDigits = majorDigits
		if f.major != previousMajor {
			// This is the first of a series.  Determine if we need to start counting
			if (i == len(files)-1) || f.major != files[i+1].major {
				f.minor = NoVersion
			} else {
				f.minor = 0
			}
			previousMajor = f.major
		} else {
			// Claim the next available minor version
			f.minor = files[i-1].minor + 1
		}
	}
}

// Determine any files whose names changed
func changedNames(files PFnpSlice) []RenameEntry {
	renames := make([]RenameEntry, 0)
	for _, f := range files {
		old := f.originalName
		new := f.String()
		if old != new {
			renames = append(renames, RenameEntry{oldName: old, newName: new})
		}
	}
	return renames
}

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}
