package main

import "strconv"

type RenameEntry struct {
	oldName, newName string
}

func ComputeRenames(fileNames []string, unused []int) []RenameEntry {
	files := ParseFileNames(fileNames)
	renumberMinorVersions(files)

	// Fill in gaps in major numbers.
	// Determine what major version to use to begin filling holes
	majorIdx := len(files) - 1 // Skip the first value since the largest major version will fit in that slot
	for unusedIdx := 0; unusedIdx < len(unused); unusedIdx++ {
		// If the "hole" is larger than what we would fill it with, we've completed the sequence
		if majorIdx <= 0 || unused[unusedIdx] > files[majorIdx].major {
			break
		}

		// Ensure that we're on the first file with this major version
		for ; majorIdx > 0 && files[majorIdx-1].major == files[majorIdx].major; majorIdx-- {
		}

		majorIdx--
	}
	// Undo our last shift - it's the one that pushed us too far
	if majorIdx < len(files)-1 {
		majorIdx++
	}
	// Rename the files to fill in the holes
	for _, u := range unused {
		if majorIdx >= len(files) || u > files[majorIdx].major {
			break
		}

		for oldMajor := files[majorIdx].major; majorIdx < len(files) && files[majorIdx].major == oldMajor; majorIdx++ {
			files[majorIdx].major = u
		}
	}

	return changedNames(files)
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

// Computes the renames needed to append one major group to another.
func ComputeAppend(fileNames []string, from, onto int) []RenameEntry {
	files := ParseFileNames(fileNames)
	
	ontoMaxMinor := NoVersion
	hasOntoFiles := false
	
	// Find the max minor version in the 'onto' group
	for _, f := range files {
		if f.major == onto {
			hasOntoFiles = true
			if f.minor > ontoMaxMinor {
				ontoMaxMinor = f.minor
			}
		}
	}
	
	if hasOntoFiles {
		if ontoMaxMinor == NoVersion {
			// If the onto group has only NoVersion files, we will number the existing one(s) as 0, and start appending at 1
			for _, f := range files {
				if f.major == onto {
					f.minor = 0
					// Once we have a minor version, we need some minorDigits. We'll set it to 1, and let computeDigitCounts adjust if needed
					f.minorDigits = 1
				}
			}
			ontoMaxMinor = 0
		}
	} else {
		ontoMaxMinor = -1
	}
	
	// 'from' files that need to be appended
	var fromFiles []*FileNamePieces
	for _, f := range files {
		if f.major == from {
			fromFiles = append(fromFiles, f)
		}
	}
	
	// Append 'from' files to 'onto'
	// Minor numbers are sequentially assigned starting after ontoMaxMinor
	for _, f := range fromFiles {
		ontoMaxMinor++
		f.major = onto
		f.minor = ontoMaxMinor
		f.majorDigits = len(strconv.Itoa(f.major))
		f.minorDigits = len(strconv.Itoa(f.minor))
	}
	
	// Now all files have correct major and minor numbers.
	// We need to re-sort the files so that `computeDigitCounts` and `changedNames` process things in the correct order,
	// though `computeDigitCounts` actually doesn't care about order. We'll leave it since `files` are already collected.
	// Oh wait, `changedNames` doesn't strictly depend on order, it just compares `originalName` to `f.String()`.

	// Recompute and apply the required padding digits for major and minor
	majorDigits, minorDigits := computeDigitCounts(files)
	var relevantFiles PFnpSlice
	for _, f := range files {
		f.majorDigits = majorDigits
		f.minorDigits = minorDigits[f.major]
		// After changing 'from' files to 'onto', any file with an originally 'from' or 'onto'
		// major number will now have f.major == onto. We restrict renames to only these files.
		if f.major == onto {
			relevantFiles = append(relevantFiles, f)
		}
	}
	
	return changedNames(relevantFiles)
}
