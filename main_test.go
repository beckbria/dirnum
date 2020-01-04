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

	noErrors := make(validationErrors)
	err, _ := validate(valid)
	assert.Equal(t, noErrors, err)
}

func TestInvalidFiles(t *testing.T) {
	invalid := []string{
		"0.jpg",
		"00.jpg",
		"000.jpg",
		"00000.jpg",
		"0_0.jpg",
		"foo.jpg",
		"0000foo.jpg"}

	for _, i := range invalid {
		err, _ := validate([]string{i})
		_, found := err[i]
		assert.Equal(t, true, found, i)
	}
}
