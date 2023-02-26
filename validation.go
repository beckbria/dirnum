package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var fileRegEx = regexp.MustCompile("^([0-9]+)(-[0-9]+)?(-[A-Za-z][A-Za-z0-9]+)?\\.(jpg|png|gif)$")

type ValidationErrors map[string][]string

func (v ValidationErrors) add(filename, err string) {
	if _, found := v[filename]; !found {
		v[filename] = []string{}
	}
	v[filename] = append(v[filename], err)
}

func (errors ValidationErrors) String() string {
	if len(errors) == 0 {
		return "No errors found"
	} else {
		var sb strings.Builder
		filesWithErrors := []string{}
		for f := range errors {
			filesWithErrors = append(filesWithErrors, f)
		}
		sort.Strings(filesWithErrors)
		for _, f := range filesWithErrors {
			for _, e := range errors[f] {
				sb.WriteString(fmt.Sprintf("\"%s\": %s\n", f, e))
			}
		}
		return sb.String()
	}
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

// Returns any errors found and a list of any skipped major version numbers
func ValidateFileNames(files []string) (ValidationErrors, []int) {
	errors := make(ValidationErrors)
	seen := make(seenMajorMinor)
	for _, f := range files {
		name, err := ParseFileName(f)
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
