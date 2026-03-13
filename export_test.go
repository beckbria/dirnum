package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestPlanExport(t *testing.T) {
	files := []string{
		"1-foo.jpg",
		"2-0-bar.jpg",
		"2-1.jpg",
		"2-2.jpg",
		"3-foo, bar.jpg",
	}

	actual := PlanExport(files, "bar", 0)
	expected := map[string][]string{
		"bar": {
			"2-0-bar.jpg",
			"2-1.jpg",
			"2-2.jpg",
			"3-foo, bar.jpg",
		},
	}

	assert.Equal(t, expected, actual)
}

func TestPlanExportEmptyPrefix(t *testing.T) {
	files := []string{
		"1-foo.jpg",
		"2-0-bar.jpg",
		"2-1.jpg",
		"2-2.jpg",
		"3-foo, bar.jpg",
	}

	actual := PlanExport(files, "", 0)
	expected := map[string][]string{
		"foo": {
			"1-foo.jpg",
			"3-foo, bar.jpg",
		},
		"bar": {
			"2-0-bar.jpg",
			"2-1.jpg",
			"2-2.jpg",
			"3-foo, bar.jpg",
		},
	}

	assert.Equal(t, expected, actual)
}

func TestPlanExportMinCount(t *testing.T) {
	files := []string{
		"1-foo.jpg",
		"2-0-bar.jpg",
		"2-1.jpg",
		"2-2.jpg",
		"3-foo, bar.jpg",
		"4-baz.jpg",
	}

	actual := PlanExport(files, "", 2)
	expected := map[string][]string{
		"foo": {
			"1-foo.jpg",
			"3-foo, bar.jpg",
		},
		"bar": {
			"2-0-bar.jpg",
			"2-1.jpg",
			"2-2.jpg",
			"3-foo, bar.jpg",
		},
	}
	assert.Equal(t, expected, actual)
}

