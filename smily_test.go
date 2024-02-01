package main

import (
	"testing"
)

func TestCountSmileys(t *testing.T) {
	testCases := []struct {
		input    []string
		expected int
	}{
		{[]string{":)", ";(", ";}", ":-D"}, 2},
		{[]string{";D", ":-(", ":-)", ";~)"}, 3},
		{[]string{";]", ":[", ";*", ":$", ";-D"}, 1},
	}

	for _, tc := range testCases {
		result := countSmileys(tc.input)
		if result != tc.expected {
			t.Errorf("Unexpected result for %v: got %d, want %d", tc.input, result, tc.expected)
		}
	}
}
