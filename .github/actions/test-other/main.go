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

	fmt.Println("Search for unchecked checkboxes...")
	uncheckedCheckboxeLines := findUncheckedCheckboxes(prbody)

	for _, uncheckedCheckboxeLine := range uncheckedCheckboxeLines {
		fmt.Println("  " + uncheckedCheckboxeLine)
	}

	if len(uncheckedCheckboxeLines) > 0 {
		panic("PR body contains unchecked checklist")
	}

	fmt.Println("\nPR body does not contain unchecked checklist")
}

func findUncheckedCheckboxes(prBody string) []string {
	lines := strings.Split(prBody, "\n")
	uncheckedCheckboxeLines := []string{}

	for _, line := range lines {
		if strings.Contains(line, CHECKBOX) {
			uncheckedCheckboxeLines = append(uncheckedCheckboxeLines, line)
		}
	}

	return uncheckedCheckboxeLines
}
