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
	comments, err := github.GetListPRComments(client, ctx, prData.Owner, prData.Repo, prData.PR)
	if err != nil {
		fmt.Println(err, "Error while  getting the comments lists of the PR")
		panic(err)
	}

	fmt.Println("Search for unchecked checkboxes...")
	uncheckedCheckboxes := findUncheckedCheckboxes(prbody, comments)
	uncheckedCheckboxesStr := strings.Join(uncheckedCheckboxes, "\n")

	fmt.Println("Unchecked checkboxes : ", uncheckedCheckboxesStr)

	fmt.Println("PR SHA : ", *prData.PR.Base.SHA)
	fmt.Println("PR HEAD SHA : ", *prData.PR.Head.SHA)

	Ids, err := github.GetJobIDsForPR(client, ctx, prData.PRNumber, prData.Owner, prData.Repo, *prData.PR.Head.SHA)
	if err != nil {
		fmt.Println(err, "Error while getting the job IDs for the PR")
		panic(err)
	}
	fmt.Println("Job IDs : ", Ids)

	id := Ids[0]

	if len(uncheckedCheckboxes) > 0 {
		//github.CreateCheckRun(client, ctx, prData.Owner, prData.Repo, *prData.PR.Head.SHA, "failure", uncheckedCheckboxesStr)
		//github.UpdatePRBody(client, ctx, prData.Owner, prData.Repo, prData.PR, prbody+" ")
		fmt.Println("reRunJob PR body ")
		github.ReRun(client, ctx, prData.Owner, prData.Repo, id)
		panic("PR body contains unchecked checklist")
	}

	fmt.Println("\nPR body does not contain unchecked checklist")
}

func findUncheckedCheckboxes(prBody string, comments []github.IssueComment) []string {
	uncheckedCheckboxes := findUncheckedCheckboxesInPrBody(prBody)
	uncheckedCheckboxes = append(uncheckedCheckboxes, findUncheckedCheckboxesInComment(comments)...)

	return uncheckedCheckboxes
}

func findUncheckedCheckboxesInPrBody(prBody string) []string {
	lines := strings.Split(prBody, "\n")
	uncheckedCheckboxes := []string{}

	for _, line := range lines {
		if strings.Contains(line, CHECKBOX) {
			uncheckedCheckboxes = append(uncheckedCheckboxes, line)
		}
	}
	return uncheckedCheckboxes
}

func findUncheckedCheckboxesInComment(comments []github.IssueComment) []string {
	uncheckedCheckboxes := []string{}

	for _, comment := range comments {
		fmt.Println("test ", *comment.Body)
		uncheckedCheckboxes = append(uncheckedCheckboxes, findUncheckedCheckboxesInPrBody(*comment.Body)...)
	}

	return uncheckedCheckboxes
}
