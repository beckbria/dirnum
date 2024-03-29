package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidFiles(t *testing.T) {
	valid := []string{
		"0000.jpg",
		"0001.png",
		"0002.gif",
		"0003-0.jpg",
		"0003-1.png",
		"0003-2-foo.jpg",
		"0003-3-FOO.jpg",
		"0004-bar.gif"}

	noErrors := make(ValidationErrors)
	err, _ := ValidateFileNames(valid)
	assert.Equal(t, noErrors, err)
}

func TestInvalidFiles(t *testing.T) {
	invalid := []string{
		"0-0-0ff.jpg",
		"0-0foo.jpg",
		"0000.mp4",
		"foo.jpg",
		"0000foo.jpg"}

	for _, i := range invalid {
		err, _ := ValidateFileNames([]string{i})
		_, found := err[i]
		assert.Equal(t, true, found, i)
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
