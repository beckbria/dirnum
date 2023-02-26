// dirnum reads the list of files in a directory and asserts that they are numbered in ascending order.
// If they are not, it lists the out-of-order file names.
// Supported groupings:
// 0000.jpg, 0001.jpg, etc.
// 0000-0.jpg, 0000-1.jpg, etc. - Minor versions for grouped files
// 0000-note.jpg, 0000-0-note.jpg, etc. - Text annotations on file names
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

func main() {
	dir := flag.String("dir", "", "The directory to analyze (mandatory)")
	quiet := flag.Bool("quiet", false, "Do not print validation errors encountered")
	renumber := flag.Bool("renumber", true, "Renumber files to fill in gaps in major numbers")
	flag.Parse()

	if dir == nil || len(*dir) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	fileNames, err := ReadFileNames(*dir)
	if err != nil {
		log.Fatal(err)
	}

	errors, unused := ValidateFileNames(fileNames)
	// Display errors for any malformed filenames
	if quiet != nil && !*quiet {
		fmt.Println(errors)
	}

	// Determine file name changes
	if renumber != nil && *renumber {
		ren := suggestedRenames(fileNames, unused)
		fmt.Println("\nProposed renames: ")
		for _, r := range ren {
			fmt.Printf("%s => %s\n", r.oldName, r.newName)
		}
		if prompt("Rename files?") {
			for _, r := range ren {
				RenameFile(r.oldName, r.newName, dir)
			}
		}
	}
}

const noMinor = -99 // Indicates a file with no minor version

type renameEntry struct {
	oldName, newName string
}

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}

func suggestedRenames(fileNames []string, unused []int) []renameEntry {
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
	rename := make([]renameEntry, 0)

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
				f.minor = noMinor
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
			rename = append(rename, renameEntry{oldName: old, newName: new})
		}
	}
	return rename
}

// Prompts the user for a yes or no answer
func prompt(q string) bool {
	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s (y/n): ", q)
		a, err := in.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		// Replace line endings
		a = strings.Replace(a, "\n", "", -1)
		a = strings.Replace(a, "\r", "", -1)
		if len(a) != 1 {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(a))[0] {
		case 'y':
			return true
		case 'n':
			return false
		}
	}
}
