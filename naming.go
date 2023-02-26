package main

import (
	"sort"
)

type RenameEntry struct {
	oldName, newName string
}

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}

func ComputeRenames(fileNames []string, unused []int) []RenameEntry {
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
	rename := make([]RenameEntry, 0)

	if len(files) == 0 {
		return rename
	}

	// Compute the number of digits required by the major/minor version.
	// We intentionally ignore the edge case where filling the gaps will
	// reduce the number of digits required - if so, the extra digit
	// will likely be required soon enough.  If it's particularly important,
	// running the tool a second time will remove the extra digit.
	// The number of minor digits is computed for each major digit
	majorDigits := 0
	minorDigits := make(map[int]int)
	for _, f := range files {
		majorDigits = max(majorDigits, f.majorDigits)
		minorDigits[f.major] = max(minorDigits[f.major], f.minorDigits)
	}

	// Renumber all minor version numbers
	previousMajor := -1 // Negative number isn't a valid major version
	for i, f := range files {
		f.minorDigits = minorDigits[f.major]
		f.majorDigits = majorDigits
		if f.major != previousMajor {
			// This is the first of a series.  Determine if we need to start counting
			if (i == len(files)-1) || f.major != files[i+1].major {
				f.minor = NoMinorVersion
			} else {
				f.minor = 0
			}
			previousMajor = f.major
		} else {
			// Claim the next available minor version
			f.minor = files[i-1].minor + 1
		}
	}

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

	// Determine any files whose names changed.  Add them to the list
	for _, f := range files {
		old := f.originalName
		new := f.String()
		if old != new {
			rename = append(rename, RenameEntry{oldName: old, newName: new})
		}
	}
	return rename
}
