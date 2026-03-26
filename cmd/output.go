package cmd

import "fmt"

func skillNoun(n int) string {
	if n == 1 {
		return "skill"
	}
	return "skills"
}

func printSummary(action string, n int) {
	fmt.Printf("\n%s %d %s.\n", action, n, skillNoun(n))
}
