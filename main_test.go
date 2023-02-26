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
		RenameEntry{oldName: "5-0-Foo.jpg", newName: "0-0-Foo.jpg"},
		RenameEntry{oldName: "5-1.jpg", newName: "0-1.jpg"},
		RenameEntry{oldName: "5-2.jpg", newName: "0-2.jpg"},
		RenameEntry{oldName: "6.jpg", newName: "3.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeRenames(files, []int{0, 3, 4}))
}

func TestRenameMajorVersionDigits(t *testing.T) {
	files := []string{"0.jpg", "1.jpg", "2.jpg", "3.jpg", "4.jpg", "5.jpg", "6.jpg", "7.jpg", "8.jpg", "9.jpg", "10.jpg"}
	expected := []RenameEntry{
		RenameEntry{oldName: "0.jpg", newName: "00.jpg"},
		RenameEntry{oldName: "1.jpg", newName: "01.jpg"},
		RenameEntry{oldName: "2.jpg", newName: "02.jpg"},
		RenameEntry{oldName: "3.jpg", newName: "03.jpg"},
		RenameEntry{oldName: "4.jpg", newName: "04.jpg"},
		RenameEntry{oldName: "5.jpg", newName: "05.jpg"},
		RenameEntry{oldName: "6.jpg", newName: "06.jpg"},
		RenameEntry{oldName: "7.jpg", newName: "07.jpg"},
		RenameEntry{oldName: "8.jpg", newName: "08.jpg"},
		RenameEntry{oldName: "9.jpg", newName: "09.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeRenames(files, []int{}))
}

func TestRenameMinorVersionDigits(t *testing.T) {
	files := []string{"0.jpg", "1-0.jpg", "2-1.jpg", "2-3.jpg", "3-0.jpg", "3-1.jpg", "3-2.jpg", "3-3.jpg", "3-4.jpg", "3-5.jpg", "3-6.jpg", "3-7.jpg", "3-8.jpg", "3-9.jpg", "3-10.jpg"}
	expected := []RenameEntry{
		RenameEntry{oldName: "1-0.jpg", newName: "1.jpg"},
		RenameEntry{oldName: "2-1.jpg", newName: "2-0.jpg"},
		RenameEntry{oldName: "2-3.jpg", newName: "2-1.jpg"},
		RenameEntry{oldName: "3-0.jpg", newName: "3-00.jpg"},
		RenameEntry{oldName: "3-1.jpg", newName: "3-01.jpg"},
		RenameEntry{oldName: "3-2.jpg", newName: "3-02.jpg"},
		RenameEntry{oldName: "3-3.jpg", newName: "3-03.jpg"},
		RenameEntry{oldName: "3-4.jpg", newName: "3-04.jpg"},
		RenameEntry{oldName: "3-5.jpg", newName: "3-05.jpg"},
		RenameEntry{oldName: "3-6.jpg", newName: "3-06.jpg"},
		RenameEntry{oldName: "3-7.jpg", newName: "3-07.jpg"},
		RenameEntry{oldName: "3-8.jpg", newName: "3-08.jpg"},
		RenameEntry{oldName: "3-9.jpg", newName: "3-09.jpg"},
	}
	assert.ElementsMatch(t, expected, ComputeRenames(files, []int{}))
}
