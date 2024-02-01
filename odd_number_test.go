package main

import (
	"testing"
)

func TestFindOddNumber(t *testing.T) {
	tests := []struct {
		input  []int
		output int
	}{
		{[]int{7}, 7},
		{[]int{0}, 0},
		{[]int{1, 1, 2}, 2},
		{[]int{0, 1, 0, 1, 0}, 0},
		{[]int{1, 2, 2, 3, 3, 3, 4, 3, 3, 3, 2, 2, 1}, 4},
	}

	for _, test := range tests {
		result := FindOddNumber(test.input)
		if result != test.output {
			t.Errorf("For input %v, expected %d but got %d", test.input, test.output, result)
		}
	}
}
