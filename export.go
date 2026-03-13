package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PlanExport determines which files should be copied to which subdirectories based on tags.
// It returns a map of tag name to a slice of filenames.
func PlanExport(files []string, prefix string, minCount int) map[string][]string {
	stats := ComputeStats(files)

	exportPlan := make(map[string][]string)

	for _, stat := range stats {
		if !strings.HasPrefix(stat.tag, prefix) {
			continue
		}

		if len(stat.files) < minCount {
			continue
		}

		// Find the major versions of the files that explicitly have this tag
		majorVersions := make(map[int]bool)
		for _, f := range stat.files {
			parsed, err := ParseFileName(f)
			if err == nil {
				majorVersions[parsed.major] = true
			}
		}

		// Find all files that share these major versions
		var filesToExport []string
		for _, f := range files {
			parsed, err := ParseFileName(f)
			if err == nil && majorVersions[parsed.major] {
				filesToExport = append(filesToExport, f)
			}
		}

		if len(filesToExport) > 0 {
			exportPlan[stat.tag] = filesToExport
		}
	}

	return exportPlan
}

// ExportTags copies files into subdirectories based on their tags and their associated major versions.
func ExportTags(dir string, files []string, prefix string, minCount int) error {
	exportPlan := PlanExport(files, prefix, minCount)

	if len(exportPlan) == 0 {
		fmt.Println("No tags matching the given prefix were found.")
		return nil
	}

	// Check for conflicting directories
	var conflictingDirs []string
	for tag := range exportPlan {
		targetDir := filepath.Join(dir, tag)
		info, err := os.Stat(targetDir)
		if err == nil && info.IsDir() {
			conflictingDirs = append(conflictingDirs, tag)
		}
	}

	if len(conflictingDirs) > 0 {
		return fmt.Errorf("cannot proceed, the following matching subdirectories already exist: %s", strings.Join(conflictingDirs, ", "))
	}

	// Count totals for prompt
	numFiles := 0
	for _, filesToExport := range exportPlan {
		numFiles += len(filesToExport)
	}

	// Prompt user for confirmation
	q := fmt.Sprintf("This will create %d subdirectories containing a total of %d files.  Continue?", len(exportPlan), numFiles)
	if !prompt(q) {
		return nil
	}

	// Execute the plan
	for tag, filesToExport := range exportPlan {
		targetDir := filepath.Join(dir, tag)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
		}

		for _, f := range filesToExport {
			src := filepath.Join(dir, f)
			dst := filepath.Join(targetDir, f)
			fmt.Printf("Copying %s to %s\n", f, filepath.Join(tag, f))
			if err := CopyFile(src, dst); err != nil {
				return fmt.Errorf("failed to copy %s to %s: %w", src, dst, err)
			}
		}
	}

	return nil
}
