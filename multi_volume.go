package unrarwrapper

import (
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// rar: first vol is basename.part1.rar
const rarMultiPartRegex = `(?mi)^(.+)\.part([0-9]+)(\.rar)$`

var rarMultiPartRx *regexp.Regexp

func init() {
	rarMultiPartRx = regexp.MustCompile(rarMultiPartRegex)
}


// rar multi volume filename format: basename.partX.rar
func multiVolArchiveDetect(targetPath string) (isMultiVol bool, firstVolFullPath string, multiVolFiles []string, err error) {
	extLower := strings.ToLower(filepath.Ext(targetPath))
	firstVolFullPath = targetPath
	if extLower != ".rar" {
		return false, targetPath, []string{targetPath}, nil
	}

	if matches := rarMultiPartRx.FindStringSubmatch(filepath.Base(targetPath)); len(matches) == 4 {
		isMultiVol = true
		multiVolBname := matches[1]
		multiVolExt := matches[3]
		// fix filename, find the first rar vol as filename
		firstVolFullPath = filepath.Join(filepath.Dir(targetPath), fmt.Sprintf("%s.part%s%s", multiVolBname, "1", multiVolExt))
		globPattern := filepath.Join(filepath.Dir(targetPath), fmt.Sprintf("%s.part*%s", multiVolBname, multiVolExt))
		multiVolFiles, _ = filepath.Glob(globPattern)
	}

	// sort multiVolFiles
	slices.SortFunc(multiVolFiles, func(a, b string) int {
		return naturalCompare(a, b)
	})
	firstVolFullPath, err = filepath.Abs(firstVolFullPath)
	return
}

var chunkifyRegexp = regexp.MustCompile(`(\d+|\D+)`)

func chunkify(s string) []string {
	return chunkifyRegexp.FindAllString(s, -1)
}

// naturalCompare returns true if the first string precedes the second one according to natural order
// code mod from https://github.com/facette/natsort/blob/master/natsort.go
func naturalCompare(a, b string) int {
	if a == b {
		return 0
	}

	chunksA := chunkify(a)
	chunksB := chunkify(b)

	nChunksA := len(chunksA)
	nChunksB := len(chunksB)

	for i := range chunksA {
		if i >= nChunksB {
			return 1
		}

		aInt, aErr := strconv.Atoi(chunksA[i])
		bInt, bErr := strconv.Atoi(chunksB[i])

		// If both chunks are numeric, compare them as integers
		if aErr == nil && bErr == nil {
			if aInt == bInt {
				if i == nChunksA-1 {
					// We reached the last chunk of A, thus B is greater than A
					return -1
				} else if i == nChunksB-1 {
					// We reached the last chunk of B, thus A is greater than B
					return 1
				}

				continue
			}

			if aInt < bInt {
				return -1
			} else {
				return 1
			}
		}

		// So far both strings are equal, continue to next chunk
		if chunksA[i] == chunksB[i] {
			if i == nChunksA-1 {
				// We reached the last chunk of A, thus B is greater than A
				return -1
			} else if i == nChunksB-1 {
				// We reached the last chunk of B, thus A is greater than B
				return 1
			}

			continue
		}

		if chunksA[i] < chunksB[i] {
			return -1
		} else {
			return 1
		}
	}

	return 1
}