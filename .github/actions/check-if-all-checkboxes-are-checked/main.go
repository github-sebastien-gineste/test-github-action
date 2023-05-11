package main

import (
	"actions/commons/github"
	"fmt"
	"strings"
)

const CHECKBOX = "- [ ]"

func main() {
	client, ctx := github.ConnectClient()

	prData := github.GetPullRequestData(client, ctx)

	prbody := prData.PR.GetBody()

	fmt.Println("Search for unchecked checkboxes... ")
	uncheckedCheckboxes := findUncheckedCheckboxes(prbody)

	for _, uncheckedCheckboxe := range uncheckedCheckboxes {
		fmt.Println("  " + uncheckedCheckboxe)
	}

	if len(uncheckedCheckboxes) > 0 {
		panic("PR body contains unchecked checklist")
	}

	fmt.Println("\nPR body does not contain unchecked checklist")
}

func findUncheckedCheckboxes(prBody string) []string {
	lines := strings.Split(prBody, "\n")
	uncheckedCheckboxes := []string{}

	for _, line := range lines {
		if strings.Contains(line, CHECKBOX) {
			uncheckedCheckboxes = append(uncheckedCheckboxes, line)
		}
	}

	return uncheckedCheckboxes
}
