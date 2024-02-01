package main

import (
	"sort"
)

func generatePermutations(chars []rune, start int, result *[]string) {
	if start == len(chars)-1 {
		*result = append(*result, string(chars))
		return
	}

	used := make(map[rune]bool)
	for i := start; i < len(chars); i++ {
		if used[chars[i]] {
			continue
		}
		used[chars[i]] = true
		chars[i], chars[start] = chars[start], chars[i]
		generatePermutations(chars, start+1, result)
		chars[i], chars[start] = chars[start], chars[i]
	}
}

func Manipulate(text interface{}) []string {
	var chars []rune

	switch val := text.(type) {
	case string:
		chars = []rune(val)
	case []string:
		chars = []rune(concatenateStrings(val))
	}

	var result []string
	generatePermutations(chars, 0, &result)

	sort.Strings(result)
	result = removeDuplicates(result)

	return result
}

func concatenateStrings(strings []string) string {
	var result string
	for _, str := range strings {
		result += str
	}
	return result
}

func removeDuplicates(arr []string) []string {
	unique := make(map[string]bool)
	result := make([]string, 0, len(arr))

	for _, str := range arr {
		if !unique[str] {
			unique[str] = true
			result = append(result, str)
		}
	}

	return result
}
