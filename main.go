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
		ren := ComputeRenames(fileNames, unused)
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
