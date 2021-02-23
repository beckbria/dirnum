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
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func main() {
	dir := flag.String("dir", "", "The directory to analyze (mandatory)")
	af := flag.Bool("fix", false, "Automatically fix simple typos in file names")
	showUnused := flag.Bool("unused", false, "Print a list of major numbers missing from the sequence")
	quiet := flag.Bool("quiet", false, "Do not print validation errors encountered")
	renumber := flag.Bool("renumber", false, "Automatically renumber files to fill in gaps in major numbers")
	flag.Parse()

	if dir == nil || len(*dir) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	files, err := ioutil.ReadDir(*dir)
	if err != nil {
		log.Fatal(err)
	}
	fileNames := []string{}
	fix := af != nil && *af
	for _, f := range files {
		n := f.Name()
		if !ignoreRegEx.MatchString(n) {
			if fix {
				n = autoFix(n, dir)
			}
			fileNames = append(fileNames, n)
		}
	}

	errors, unused := validate(fileNames)
	if quiet != nil && !*quiet {
		if len(errors) == 0 {
			fmt.Println("No errors found")
		} else {
			filesWithErrors := []string{}
			for f := range errors {
				filesWithErrors = append(filesWithErrors, f)
			}
			sort.Strings(filesWithErrors)
			for _, f := range filesWithErrors {
				for _, e := range errors[f] {
					fmt.Printf("\"%s\": %s\n", f, e)
				}
			}
		}
	}

	if showUnused != nil && *showUnused {
		fmt.Print("Unused major numbers: ")
		fmt.Println(unused)
	}

	if renumber != nil && *renumber {
		ren := suggestedRenames(fileNames, unused)
		fmt.Println("\nProposed renames: ")
		for _, r := range ren {
			fmt.Printf("%s => %s\n", r.oldName, r.newName)
		}
		if prompt("Rename files?") {
			for _, r := range ren {
				renameFile(r.oldName, r.newName, dir)
			}
		}
	}
}

var (
	fileRegEx   = regexp.MustCompile("^([0-9]+)(-[0-9]+)?(-[A-Za-z][A-Za-z0-9]+)?\\.(jpg|png|gif)$")
	ignoreRegEx = regexp.MustCompile("^Thumbs\\.db$")
	autoFixes   = []*fix{
		newFix("^([0-9][0-9][0-9][0-9])_([0-9]+)\\.(jpg|png|gif)$", "%s-%s.%s"),
		newFix("^([0-9][0-9][0-9][0-9]).JPG$", "%s.jpg"),
		newFix("^([0-9][0-9][0-9][0-9])-([0-9]+).JPG$", "%s-%s.jpg")}
)

const noMinor = -99

type fix struct {
	regex       *regexp.Regexp // Pattern to match to trigger automatic filename fix
	replacement string         // Format string accepting string parameters for all the tokens in the pattern
}

type fileNamePieces struct {
	major, minor, majorDigits, minorDigits int
	originalName, descriptor, extension    string
}

func (f *fileNamePieces) String() string {
	var b strings.Builder
	b.Grow(len(f.originalName))
	b.WriteString(prependZeroes(strconv.Itoa(f.major), f.majorDigits)) // Major version
	if f.minor != noMinor {
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

func parseFileName(f string) (*fileNamePieces, error) {
	tokens := fileRegEx.FindStringSubmatch(f)
	if tokens == nil {
		return nil, fmt.Errorf("Bad filename: %s", f)
	}
	major, err := strconv.Atoi(tokens[1])
	if err != nil {
		return nil, fmt.Errorf("Invalid major version \"%s\": %s", tokens[1], f)
	}
	minor := noMinor
	if len(tokens[2]) > 0 {
		minorStr := string([]rune(tokens[2])[1:])
		m, err := strconv.Atoi(minorStr)
		if err != nil {
			return nil, fmt.Errorf("Invalid minor version \"%s\": %s", minorStr, f)
		}
		minor = m
	}
	minorDigits := 0
	if minorDigits != noMinor {
		minorDigits = len(strconv.Itoa(minor))
	}
	name := fileNamePieces{
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

// pfnpSlice represents a set of file names that can be sorted by major+minor version
type pfnpSlice []*fileNamePieces

func (s pfnpSlice) Len() int {
	return len(s)
}

func (s pfnpSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s pfnpSlice) Less(i, j int) bool {
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
	files := make(pfnpSlice, 0)
	for _, f := range fileNames {
		n, err := parseFileName(f)
		if err == nil {
			// Don't try to rename files which aren't named correctly
			files = append(files, n)
		} else {
			fmt.Printf(err.Error())
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
	// Keep track of the highest unmodified major version number
	majorIdx := len(files) - 1
	for len(unused) > 0 {
		firstUnused := unused[0]
		unused = unused[1:]
		if firstUnused > files[majorIdx].major {
			// We've filled in to a continuous loop
			break
		}
		// Change the major version to the unused value
		for oldMajor := files[majorIdx].major; majorIdx >= 0 && files[majorIdx].major == oldMajor; majorIdx-- {
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

func renameFile(oldName, newName string, dirName *string) {
	oldPath := *dirName + string(os.PathSeparator) + oldName
	newPath := *dirName + string(os.PathSeparator) + newName
	fmt.Printf("Renaming %s to %s\n", oldPath, newPath)
	os.Rename(oldPath, newPath)
}

func autoFix(oldName string, dirName *string) string {
	for _, f := range autoFixes {
		tokens := f.regex.FindStringSubmatch(oldName)
		if tokens == nil {
			continue
		}
		t := tokens[1:]
		t2 := []interface{}{}
		for _, s := range t {
			t2 = append(t2, interface{}(s))
		}
		newName := fmt.Sprintf(f.replacement, t2...)
		renameFile(oldName, newName, dirName)
		return newName
	}
	// This isn't a file we can fix
	return oldName
}

func newFix(pattern, replacement string) *fix {
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatalf("Invalid pattern: %s", pattern)
	}
	f := fix{regex: re, replacement: replacement}
	return &f
}

// seenMajorMinor maps from the major number to the minor number to the filename
type seenMajorMinor map[int]map[int]string

func (s seenMajorMinor) add(major, minor int, file string) error {
	if _, found := s[major]; !found {
		s[major] = make(map[int]string)
	}
	if _, found := s[major][minor]; found {
		return fmt.Errorf("duplicate major/minor entry")
	}
	s[major][minor] = file
	return nil
}

type validationErrors map[string][]string

func (v validationErrors) add(filename, err string) {
	if _, found := v[filename]; !found {
		v[filename] = []string{}
	}
	v[filename] = append(v[filename], err)
}

// Returns any errors found and a list of any skipped major version numbers
func validate(files []string) (validationErrors, []int) {
	errors := make(validationErrors)
	seen := make(seenMajorMinor)
	for _, f := range files {
		name, err := parseFileName(f)
		if err != nil {
			errors.add(f, err.Error())
			continue
		}
		err = seen.add(name.major, name.minor, f)
		if err != nil {
			oldFile := seen[name.major][name.minor]
			errText := ""
			if name.minor == noMinor {
				errText = fmt.Sprintf("Overridden Major Number %d for files: \"%s\", \"%s\"", name.major, oldFile, f)
			} else {
				errText = fmt.Sprintf("Duplicate Major/Minor %d-%d for files: \"%s\", \"%s\"", name.major, name.minor, oldFile, f)
			}
			errors.add(f, errText)
			errors.add(oldFile, errText)
			continue
		}
	}

	major := []int{}
	for m := range seen {
		major = append(major, m)
	}
	sort.Ints(major)

	majErrors, unused := validateMajor(major)
	for n, e := range majErrors {
		f := ""
		for _, fileName := range seen[n] {
			f = fileName
			break
		}
		errors.add(f, fmt.Sprintf(e, f))
	}

	for maj, mins := range seen {
		minor := []int{}
		for m := range mins {
			minor = append(minor, m)
		}
		sort.Ints(minor)
		minorErrors := validateMinor(minor)
		for min, e := range minorErrors {
			f := seen[maj][min]
			errors.add(f, fmt.Sprintf(e, f))
		}
	}

	sort.Ints(unused)
	return errors, unused
}

// Returns an map from major version number to error format string which accepts the file name
func validateMajor(nums []int) (map[int]string, []int) {
	unused := []int{}
	errors := make(map[int]string)
	prev := -1
	for _, n := range nums {
		if n != (prev + 1) {
			errors[n] = fmt.Sprintf("Numbering jumped from %d to %d: %%s", prev, n)
			start := prev + 1
			if start < 0 {
				start = 0
			}
			for i := start; i < n; i++ {
				unused = append(unused, i)
			}
		}
		prev = n
	}

	return errors, unused
}

// Returns an map from minor version number to error format string which accepts the file name
func validateMinor(nums []int) map[int]string {
	errors := make(map[int]string)
	if len(nums) == 1 {
		if nums[0] != noMinor {
			errors[nums[0]] = fmt.Sprintf("Minor version %d on single file: %%s", nums[0])
		}
	} else if len(nums) > 1 {
		prev := -1
		for _, n := range nums {
			if n != (prev + 1) {
				if prev == -1 || prev == noMinor {
					errors[n] = "Minor version numbering must start with 0: %s"
				} else {
					errors[n] = fmt.Sprintf("Minor numbering jumped from %d to %d: %%s", prev, n)
				}
			}
			prev = n
		}
	}

	return errors
}
