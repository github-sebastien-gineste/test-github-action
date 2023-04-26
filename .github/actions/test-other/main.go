package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const PR_NUMBER = "PR_NUMBER"
const OWNER = "OWNER"
const REPO = "REPO"
const GITHUB_TOKEN = "GITHUB_TOKEN"
const CHECKBOX = "\n- [ ]"

type PullRequestData struct {
	prNumber int
	owner    string
	repo     string
	pr       *github.PullRequest
}

func main() {
	ctx := context.Background()
	client := ConnectClient(ctx)

	prData := getPullRequestData(client, ctx)

	prbody := *prData.pr.Body

	if isContainsUncheckedCheckBox(prbody) {
		panic("PR body contains unchecked checklist")
	}
}

func isContainsUncheckedCheckBox(prBody string) bool {
	return strings.Contains(prBody, CHECKBOX)
}

func ConnectClient(ctx context.Context) *github.Client {
	token := os.Getenv(GITHUB_TOKEN)
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func getPullRequestData(client *github.Client, ctx context.Context) PullRequestData {
	prNumberStr := os.Getenv(PR_NUMBER)
	prNumber, err := strconv.Atoi(prNumberStr)
	if err != nil {
		panic(err)
	}
	owner := os.Getenv(OWNER)
	repo := os.Getenv(REPO)
	pr, _, err := client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		fmt.Println(err, "Error while retrieving the PR informations")
		panic(err)
	}

	return PullRequestData{
		prNumber: prNumber,
		owner:    owner,
		repo:     repo,
		pr:       pr,
	}
}
