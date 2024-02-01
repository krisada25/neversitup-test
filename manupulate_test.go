package main

func Permute(input string) []string {
	var result []string
	chars := []rune(input)
	n := len(chars)

	swap := func(i, j int) {
		chars[i], chars[j] = chars[j], chars[i]
	}

	var generatePermutations func(int)
	generatePermutations = func(index int) {
		if index == n {
			result = append(result, string(chars))
			return
		}
		used := make(map[rune]bool)
		for i := index; i < n; i++ {
			if used[chars[i]] {
				continue
			}
			used[chars[i]] = true
			swap(index, i)
			generatePermutations(index + 1)
			swap(index, i)
		}
	}

	generatePermutations(0)
	return result
}
