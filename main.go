// dirnum reads the list of files in a directory and asserts that they are numbered in ascending order.
// If they are not, it lists the out-of-order file names.
// Supported groupings:
// 0000.jpg, 0001.jpg, etc.
// 0000-0.jpg, 0000-1.jpg, etc. - Minor versions for grouped files
// 0000-note.jpg, 0000-0-note.jpg, etc. - Text annotations on file names
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
)

var (
	fileRegEx   = regexp.MustCompile("^([0-9][0-9][0-9][0-9])(-[0-9]+)?(-[A-Za-z]+)?\\.(jpg|png|gif)$")
	ignoreRegEx = regexp.MustCompile("^Thumbs\\.db$")
	autoFixes   = []*fix{
		newFix("^([0-9][0-9][0-9][0-9])_([0-9]+)\\.(jpg|png|gif)$", "%s-%s.%s"),
		newFix("^([0-9][0-9][0-9][0-9]).JPG$", "%s.jpg"),
		newFix("^([0-9][0-9][0-9][0-9])-([0-9]+).JPG$", "%s-%s.jpg")}
)

func main() {
	dir := flag.String("dir", "", "The directory to analyze")
	af := flag.Bool("fix", false, "Whether to automatically fix simple typos in file names")
	pu := flag.Bool("printUnused", false, "Print a list of major numbers missing from the sequence")
	flag.Parse()

	if dir == nil || len(*dir) < 1 {
		log.Fatalf(
			"Usage: %s -dir=directory\n\n"+
				"Options:\n"+
				"  -fix=true\t\tAutomatically fix simple filename typos\n"+
				"  -printUnused=true\tPrint a list of skipped major versions\n\n", os.Args[0])
	}

	files, err := ioutil.ReadDir(*dir)
	if err != nil {
		log.Fatal(err)
	}
	fileNames := []string{}
	for _, f := range files {
		fn := f.Name()
		if !ignoreRegEx.MatchString(fn) {
			if af != nil && *af {
				fn = autoFix(fn, *dir)
			}
			fileNames = append(fileNames, fn)
		}
	}

	errors, unused := validate(fileNames)
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

	if pu != nil && *pu {
		fmt.Println(unused)
	}
}

type fix struct {
	regex       *regexp.Regexp // Pattern to match to trigger automatic filename fix
	replacement string         // Format string accepting string parameters for all the tokens in the pattern
}

func autoFix(oldName, dirName string) string {
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
		oldPath := dirName + string(os.PathSeparator) + oldName
		newPath := dirName + string(os.PathSeparator) + newName
		fmt.Printf("Renaming %s to %s\n", oldPath, newPath)
		os.Rename(oldPath, newPath)
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

const noMinor = -99

// Returns any errors found and a list of any skipped major version numbers
func validate(files []string) (validationErrors, []int) {
	errors := make(validationErrors)
	seen := make(seenMajorMinor)
	for _, f := range files {
		tokens := fileRegEx.FindStringSubmatch(f)
		if tokens == nil {
			errors.add(f, fmt.Sprintf("Bad filename: %s", f))
			continue
		}
		major, err := strconv.Atoi(tokens[1])
		if err != nil {
			errors.add(f, fmt.Sprintf("Invalid major version \"%s\": %s", tokens[1], f))
			continue
		}
		minor := noMinor
		if len(tokens[2]) > 0 {
			minorStr := string([]rune(tokens[2])[1:])
			m, err := strconv.Atoi(minorStr)
			if err != nil {
				errors.add(f, fmt.Sprintf("Invalid minor version \"%s\": %s", minorStr, f))
				continue
			}
			minor = m
		}
		e := seen.add(major, minor, f)
		if e != nil {
			oldFile := seen[major][minor]
			errText := ""
			if minor == noMinor {
				errText = fmt.Sprintf("Overridden Major Number %d for files: \"%s\", \"%s\"", major, oldFile, f)
			} else {
				errText = fmt.Sprintf("Duplicate Major/Minor %d-%d for files: \"%s\", \"%s\"", major, minor, oldFile, f)
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
