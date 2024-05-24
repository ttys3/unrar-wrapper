package unrarwrapper

import (
	"slices"
	"testing"
)

func TestNaturalCompare(t *testing.T) {
	multiVolFiles := []string{
		"basename.part20.rar",
		"basename.part30.rar",
		"basename.part1.rar",
		"basename.part2.rar",
		"basename.part3.rar",
	}
	slices.SortFunc(multiVolFiles, func(a, b string) int {
		return naturalCompare(a, b)
	})
	t.Logf("multiVolFiles: %v", multiVolFiles)

	expect := []string{
		"basename.part1.rar",
		"basename.part2.rar",
		"basename.part3.rar",
		"basename.part20.rar",
		"basename.part30.rar",
	}

	if !slices.Equal(multiVolFiles, expect) {
		t.Errorf("multiVolFiles: %v, expected: %v", multiVolFiles, expect)
	}
}