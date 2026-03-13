package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// MetadataStat counts the number of times a tag is seen
type MetadataStat struct {
	tag   string
	files []string
}

// ComputeStats looks through a list of filenames, gathers the list of tags, and counts how many times each is referenced.
func ComputeStats(fileNames []string) []MetadataStat {
	tagMap := make(map[string][]string)
	for _, f := range fileNames {
		parsed, err := ParseFileName(f)
		if err != nil {
			continue // Skip files that don't match the expected format
		}

		desc := parsed.descriptor
		if len(desc) > 0 && desc[0] == '-' {
			desc = desc[1:] // strip leading dash
		}
		if len(desc) == 0 {
			continue
		}

		tags := strings.Split(desc, ",")
		for _, t := range tags {
			t = strings.TrimSpace(t)
			if len(t) > 0 {
				tagMap[t] = append(tagMap[t], f)
			}
		}
	}

	stats := make([]MetadataStat, 0, len(tagMap))
	for tag, files := range tagMap {
		stats = append(stats, MetadataStat{tag: tag, files: files})
	}
	return stats
}

func SortStatsAlphabetical(stats []MetadataStat) {
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].tag < stats[j].tag
	})
}

func SortStatsByFrequency(stats []MetadataStat) {
	sort.Slice(stats, func(i, j int) bool {
		if len(stats[i].files) == len(stats[j].files) {
			return stats[i].tag < stats[j].tag
		}
		return len(stats[i].files) > len(stats[j].files) // descending frequency
	})
}

func PrintTagCounts(stats []MetadataStat) {
	for _, s := range stats {
		fmt.Printf("%d\t%s\n", len(s.files), s.tag)
	}
}

func PrintTagMajorVersions(stats []MetadataStat) {
	for _, s := range stats {
		var majors []int
		majorSet := make(map[int]bool)
		for _, f := range s.files {
			parsed, err := ParseFileName(f)
			if err == nil && !majorSet[parsed.major] {
				majorSet[parsed.major] = true
				majors = append(majors, parsed.major)
			}
		}
		sort.Ints(majors)
		majorStrs := make([]string, len(majors))
		for i, m := range majors {
			majorStrs[i] = strconv.Itoa(m)
		}
		fmt.Printf("%s\t%s\n", s.tag, strings.Join(majorStrs, ", "))
	}
}
