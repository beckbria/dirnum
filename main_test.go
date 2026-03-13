package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidFiles(t *testing.T) {
	valid := []string{
		"0000.jpg",
		"0001.gif",
		"0002-0.jpg",
		"0002-1.jpeg",
		"0002-2-foo.jpg",
		"0002-3-FOO.jpg",
		"0003-bar.gif"}

	noErrors := make(ValidationErrors)
	errors, _ := ValidateFileNames(valid, false, false)
	assert.Equal(t, noErrors, errors)
}

func TestInvalidFiles(t *testing.T) {
	invalid := []string{
		"0-0-0ff.jpg",
		"0-0foo.jpg",
		"0000.mp4",
		"foo.jpg",
		"0000foo.jpg"}

	for _, i := range invalid {
		errors, _ := ValidateFileNames([]string{i}, false, false)
		_, found := errors[i]
		assert.True(t, found, i)
	}
}

func TestRenameFillGaps(t *testing.T) {
	files := []string{"1.jpg", "2-Foo.jpg", "5-0-Foo.jpg", "5-1.jpg", "5-2.jpg", "6.jpg"}
	expected := []RenameEntry{
		{oldName: "5-0-Foo.jpg", newName: "0-0-Foo.jpg"},
		{oldName: "5-1.jpg", newName: "0-1.jpg"},
		{oldName: "5-2.jpg", newName: "0-2.jpg"},
		{oldName: "6.jpg", newName: "3.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeRenames(files, []int{0, 3, 4}))
}

func TestMissingZero(t *testing.T) {
	files := []string{"0.jpg", "0-1.jpg", "0-2-Foo.jpg"}
	expectedRenames := []RenameEntry{
		{oldName: "0.jpg", newName: "0-0.jpg"},
	}
	expectedErrors := ValidationErrors{
		"0.jpg": []string{"Minor version numbering must start with 0: 0.jpg"},
	}

	errors, unused := ValidateFileNames(files, true, false)
	assert.Equal(t, expectedErrors, errors)
	assert.ElementsMatch(t, expectedRenames, ComputeRenames(files, unused))
}

func TestMissingZeroWithDescriptor(t *testing.T) {
	files := []string{"0-Bar Baz.jpg", "0-1.jpg"}
	expectedRenames := []RenameEntry{
		{oldName: "0-Bar Baz.jpg", newName: "0-0-Bar Baz.jpg"},
	}
	expectedErrors := ValidationErrors{
		"0-Bar Baz.jpg": []string{"Minor version numbering must start with 0: 0-Bar Baz.jpg"},
	}

	errors, unused := ValidateFileNames(files, true, false)
	assert.Equal(t, expectedErrors, errors)
	assert.ElementsMatch(t, expectedRenames, ComputeRenames(files, unused))
}

func TestRenameFillGapsEmpty(t *testing.T) {
	assert.ElementsMatch(t, []RenameEntry{}, ComputeRenames([]string{}, []int{}))
}

func TestRenameFillGapsExactlyOne(t *testing.T) {
	files := []string{"1.jpg"}
	expected := []RenameEntry{
		{oldName: "1.jpg", newName: "0.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeRenames(files, []int{0}))
}

func TestRenameFillGapsExactlyOneHighNumber(t *testing.T) {
	files := []string{"5.jpg"}
	expected := []RenameEntry{
		{oldName: "5.jpg", newName: "0.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeRenames(files, []int{0, 1, 2, 3, 4}))
}

func TestRenameFillGapsExactlyTwoStartHole(t *testing.T) {
	files := []string{"1.jpg", "2.jpg"}
	expected := []RenameEntry{
		{oldName: "2.jpg", newName: "0.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeRenames(files, []int{0}))
}

func TestRenameFillGapsExactlyTwoMidHole(t *testing.T) {
	files := []string{"0.jpg", "2.jpg"}
	expected := []RenameEntry{
		{oldName: "2.jpg", newName: "1.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeRenames(files, []int{1}))
}

func TestRenameFillGapsExactlyTwoTwoHoles(t *testing.T) {
	files := []string{"1.jpg", "3.jpg"}
	expected := []RenameEntry{
		{oldName: "3.jpg", newName: "0.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeRenames(files, []int{0, 2}))
}

func TestRenameFillGapsManyMinorVersions(t *testing.T) {
	files := []string{"0.jpg", "2.jpg", "4-0.jpg", "4-1.jpg", "4-2.jpg", "4-3.jpg", "4-4.jpg", "4-5.jpg"}
	expected := []RenameEntry{
		{oldName: "4-0.jpg", newName: "1-0.jpg"},
		{oldName: "4-1.jpg", newName: "1-1.jpg"},
		{oldName: "4-2.jpg", newName: "1-2.jpg"},
		{oldName: "4-3.jpg", newName: "1-3.jpg"},
		{oldName: "4-4.jpg", newName: "1-4.jpg"},
		{oldName: "4-5.jpg", newName: "1-5.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeRenames(files, []int{1, 3}))
}

func TestRenameFillGapsPenultimateManyMinorVersions(t *testing.T) {
	files := []string{"0.jpg", "2.jpg", "4-0.jpg", "4-1.jpg", "4-2.jpg", "4-3.jpg", "4-4.jpg", "4-5.jpg", "5.jpg"}
	expected := []RenameEntry{
		{oldName: "4-0.jpg", newName: "1-0.jpg"},
		{oldName: "4-1.jpg", newName: "1-1.jpg"},
		{oldName: "4-2.jpg", newName: "1-2.jpg"},
		{oldName: "4-3.jpg", newName: "1-3.jpg"},
		{oldName: "4-4.jpg", newName: "1-4.jpg"},
		{oldName: "4-5.jpg", newName: "1-5.jpg"},
		{oldName: "5.jpg", newName: "3.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeRenames(files, []int{1, 3}))
}

func TestRenameMajorVersionDigits(t *testing.T) {
	files := []string{"0.jpg", "1.jpg", "2.jpg", "3.jpg", "4.jpg", "5.jpg", "6.jpg", "7.jpg", "8.jpg", "9.jpg", "10.jpg"}
	expected := []RenameEntry{
		{oldName: "0.jpg", newName: "00.jpg"},
		{oldName: "1.jpg", newName: "01.jpg"},
		{oldName: "2.jpg", newName: "02.jpg"},
		{oldName: "3.jpg", newName: "03.jpg"},
		{oldName: "4.jpg", newName: "04.jpg"},
		{oldName: "5.jpg", newName: "05.jpg"},
		{oldName: "6.jpg", newName: "06.jpg"},
		{oldName: "7.jpg", newName: "07.jpg"},
		{oldName: "8.jpg", newName: "08.jpg"},
		{oldName: "9.jpg", newName: "09.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeRenames(files, []int{}))
}

func TestRenameMinorVersionDigits(t *testing.T) {
	files := []string{"0.jpg", "1-0.jpg", "2-1.jpg", "2-3.jpg", "3-0.jpg", "3-1.jpg", "3-2.jpg", "3-3.jpg", "3-4.jpg", "3-5.jpg", "3-6.jpg", "3-7.jpg", "3-8.jpg", "3-9.jpg", "3-10.jpg"}
	expected := []RenameEntry{
		{oldName: "1-0.jpg", newName: "1.jpg"},
		{oldName: "2-1.jpg", newName: "2-0.jpg"},
		{oldName: "2-3.jpg", newName: "2-1.jpg"},
		{oldName: "3-0.jpg", newName: "3-00.jpg"},
		{oldName: "3-1.jpg", newName: "3-01.jpg"},
		{oldName: "3-2.jpg", newName: "3-02.jpg"},
		{oldName: "3-3.jpg", newName: "3-03.jpg"},
		{oldName: "3-4.jpg", newName: "3-04.jpg"},
		{oldName: "3-5.jpg", newName: "3-05.jpg"},
		{oldName: "3-6.jpg", newName: "3-06.jpg"},
		{oldName: "3-7.jpg", newName: "3-07.jpg"},
		{oldName: "3-8.jpg", newName: "3-08.jpg"},
		{oldName: "3-9.jpg", newName: "3-09.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeRenames(files, []int{}))
}

func TestComputeStats(t *testing.T) {
	files := []string{
		"0-Foo.jpg",
		"1-0-Bar.jpg",
		"1-1-Foo, Baz.jpg",
		"2.jpg", // No stats
		"foo.jpg", // Invalid
	}
	expected := []MetadataStat{
		{tag: "Foo", files: []string{"0-Foo.jpg", "1-1-Foo, Baz.jpg"}},
		{tag: "Bar", files: []string{"1-0-Bar.jpg"}},
		{tag: "Baz", files: []string{"1-1-Foo, Baz.jpg"}},
	}
	
	actual := ComputeStats(files)
	assert.ElementsMatch(t, expected, actual)
}

func TestSortStatsAlphabetical(t *testing.T) {
	stats := []MetadataStat{
		{tag: "Foo", files: []string{"1.jpg", "2.jpg"}},
		{tag: "Bar", files: []string{"3.jpg"}},
		{tag: "Apple", files: []string{"4.jpg"}},
	}
	
	SortStatsAlphabetical(stats)
	
	expected := []MetadataStat{
		{tag: "Apple", files: []string{"4.jpg"}},
		{tag: "Bar", files: []string{"3.jpg"}},
		{tag: "Foo", files: []string{"1.jpg", "2.jpg"}},
	}
	assert.Equal(t, expected, stats)
}

func TestSortStatsByFrequency(t *testing.T) {
	stats := []MetadataStat{
		{tag: "Apple", files: []string{"4.jpg"}},
		{tag: "Foo", files: []string{"1.jpg", "2.jpg"}},
		{tag: "Bar", files: []string{"3.jpg"}},
		{tag: "Zeta", files: []string{"5.jpg", "6.jpg"}},
	}
	
	SortStatsByFrequency(stats)
	
	expected := []MetadataStat{
		{tag: "Foo", files: []string{"1.jpg", "2.jpg"}},
		{tag: "Zeta", files: []string{"5.jpg", "6.jpg"}},
		{tag: "Apple", files: []string{"4.jpg"}},
		{tag: "Bar", files: []string{"3.jpg"}},
	}
	assert.Equal(t, expected, stats)
}

func TestAppendBasic(t *testing.T) {
	files := []string{"1-0.jpg", "1-1.jpg", "2-0.jpg", "2-1.jpg", "2-2.jpg"}
	// Append 2 onto 1
	expected := []RenameEntry{
		{oldName: "2-0.jpg", newName: "1-2.jpg"},
		{oldName: "2-1.jpg", newName: "1-3.jpg"},
		{oldName: "2-2.jpg", newName: "1-4.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeAppend(files, 2, 1))
}

func TestAppendWithTags(t *testing.T) {
	files := []string{"1-0-Foo.jpg", "1-1-Bar.jpg", "2-0-Baz.jpg", "2-1.jpg"}
	// Append 2 onto 1
	expected := []RenameEntry{
		{oldName: "2-0-Baz.jpg", newName: "1-2-Baz.jpg"},
		{oldName: "2-1.jpg", newName: "1-3.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeAppend(files, 2, 1))
}

func TestAppendToSingleNoVersion(t *testing.T) {
	// 1 has no minor version. 2 will be appended.
	// 1 becomes 1-0, and 2-0 becomes 1-1, etc.
	files := []string{"1.jpg", "2-0.jpg", "2-1.jpg"}
	expected := []RenameEntry{
		{oldName: "1.jpg", newName: "1-0.jpg"},
		{oldName: "2-0.jpg", newName: "1-1.jpg"},
		{oldName: "2-1.jpg", newName: "1-2.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeAppend(files, 2, 1))
}

func TestAppendToEmpty(t *testing.T) {
	// 'onto' group doesn't exist yet, we just move 'from' to 'onto'
	// Minor numbers should start from 0 because NoVersion defaults max to -1, which increments to 0 for the first
	files := []string{"2-0.jpg", "2-1.jpg", "3-0.jpg"}
	expected := []RenameEntry{
		{oldName: "2-0.jpg", newName: "1-0.jpg"},
		{oldName: "2-1.jpg", newName: "1-1.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeAppend(files, 2, 1))
}

func TestAppendDigitsResized(t *testing.T) {
	// 1 has 9 items (0-8), when we append 2, it will exceed 9 items, requiring 2 digits for minor version.
	// So 1-0 becomes 1-00, ..., 1-8 becomes 1-08, and the appended ones become 1-09, 1-10, etc.
	files := []string{
		"1-0.jpg", "1-1.jpg", "1-2.jpg", "1-3.jpg", "1-4.jpg", "1-5.jpg", "1-6.jpg", "1-7.jpg", "1-8.jpg",
		"2-0.jpg", "2-1.jpg",
	}
	expected := []RenameEntry{
		{oldName: "1-0.jpg", newName: "1-00.jpg"},
		{oldName: "1-1.jpg", newName: "1-01.jpg"},
		{oldName: "1-2.jpg", newName: "1-02.jpg"},
		{oldName: "1-3.jpg", newName: "1-03.jpg"},
		{oldName: "1-4.jpg", newName: "1-04.jpg"},
		{oldName: "1-5.jpg", newName: "1-05.jpg"},
		{oldName: "1-6.jpg", newName: "1-06.jpg"},
		{oldName: "1-7.jpg", newName: "1-07.jpg"},
		{oldName: "1-8.jpg", newName: "1-08.jpg"},
		{oldName: "2-0.jpg", newName: "1-09.jpg"},
		{oldName: "2-1.jpg", newName: "1-10.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeAppend(files, 2, 1))
}

func TestAppendIgnoreOtherFiles(t *testing.T) {
	// 2.jpeg would normally be renamed to 2.jpg when naming normalizes extensions.
	// We verify that since major 2 is not part of the append operation, it gets ignored.
	files := []string{"1.jpg", "2.jpeg", "3.jpg"}
	expected := []RenameEntry{
		{oldName: "1.jpg", newName: "1-0.jpg"},
		{oldName: "3.jpg", newName: "1-1.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeAppend(files, 3, 1))
}
