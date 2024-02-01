package main

import (
	"regexp"
)

func countSmileys(arr []string) int {
	pattern := `[:;][-~]?[)D]`
	re := regexp.MustCompile(pattern)

	count := 0

	for _, face := range arr {
		if re.MatchString(face) {
			count++
		}
	}

	return count
}
