package main

func FindOddNumber(arr []int) int {
	counts := make(map[int]int)

	for _, num := range arr {
		counts[num]++
	}

	for num, count := range counts {
		if count%2 != 0 {
			return num
		}
	}

	return 0
}

func countOccurrences(arr []int, target int) int {
	count := 0
	for _, num := range arr {
		if num == target {
			count++
		}
	}
	return count
}
