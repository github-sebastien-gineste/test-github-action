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
	if isContainsUncheckedCheckBox(prbody, false) {
		panic("PR body contains unchecked checklist")
	}

	fmt.Println("\nPR body does not contain unchecked checklist")
}

func isContainsUncheckedCheckBox(prBody string, ignoreLog bool) bool {
	lines := strings.Split(prBody, "\n")
	isContainsUncheckedCheckBox := false

	if !ignoreLog {
		fmt.Println("Search for unchecked checkboxes...")
	}

	for _, line := range lines {
		if strings.Contains(line, CHECKBOX) {
			if !ignoreLog {
				fmt.Println("  " + line)
			}
			isContainsUncheckedCheckBox = true
		}
	}

	return isContainsUncheckedCheckBox
}
