package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func Manipulate(input []string) []string {
	permutations := Permute(input[0])
	sort.Strings(permutations)
	return permutations
}

func getInput() []string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter strings separated by spaces:")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	return strings.Split(input, " ")
}

func main() {
	input := getInput()
	fmt.Println(Manipulate(input))
}
