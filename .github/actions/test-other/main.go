package main

import (
	"actions/commons"
	"strings"
)

const CHECKBOX = "\n- [ ]"

func main() {
	client, ctx := commons.ConnectClient()

	prData := commons.GetPullRequestData(client, ctx)

	prbody := prData.PR.GetBody()

	if isContainsUncheckedCheckBox(prbody) {
		panic("PR body contains unchecked checklist")
	}
}

func isContainsUncheckedCheckBox(prBody string) bool {
	return strings.Contains(prBody, CHECKBOX)
}
