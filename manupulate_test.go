package main

import (
	"reflect"
	"testing"
)

func TestPermute(t *testing.T) {
	input := "abc"
	expected := []string{"abc", "acb", "bac", "bca", "cab", "cba"}

	result := Permute(input)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}
