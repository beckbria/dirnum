package main

import (
	"sort"
	"strings"
)

// MetadataStat counts the number of times a tag is seen
type MetadataStat struct {
	tag   string
	count int
}

// ComputeStats looks through a list of filenames, gathers the list of tags, and counts how many times each is referenced.
func ComputeStats(fileNames []string) []MetadataStat {
	files := ParseFileNames(fileNames)
	descriptors := sortedDescriptors(files)
	tags := extractTags(descriptors)
	stats := countTags(files, tags)
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].count < stats[j].count
	})
	return stats
}

// extractTags looks at the descriptors in file names and attempts to extract the list of tags seen.
// Rudimentary support for multi-tag descriptors is done by a subset of one string is found as an entire entry elsewhere
// That is, the set of descriptors [JimBob, Jim, BobbySue, Jimmy] should produce the tags [Jim, Bob, BobbySue, Jimmy]
func extractTags(descriptors []string) []string {
	return []string{}
}

// sortedDescriptors extracts the descriptors from a list of files and sorts them
func sortedDescriptors(files PFnpSlice) []string {
	descriptorSet := make(map[string]bool)
	for _, f := range files {
		if len(f.descriptor) > 0 {
			descriptorSet[f.descriptor] = true
		}
	}
	descriptors := make([]string, len(descriptorSet))
	i := 0
	for d := range descriptorSet {
		descriptors[i] = d
		i++
	}
	// Shortest descriptors first (as if they appear in other descriptors), then alphabetical
	sort.Slice(descriptors, func(i, j int) bool {
		if len(descriptors[i]) == len(descriptors[j]) {
			return strings.Compare(descriptors[i], descriptors[j]) < 0
		}
		return len(descriptors[i]) < len(descriptors[j])
	})
	return descriptors
}

func countTags(files PFnpSlice, tags []string) []MetadataStat {
	stats := []MetadataStat{}

	return stats
}
