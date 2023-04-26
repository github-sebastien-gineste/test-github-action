package commons

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const GITHUB_TOKEN = "GITHUB_TOKEN"
const PR_NUMBER = "PR_NUMBER"
const OWNER = "OWNER"
const REPO = "REPO"

type PullRequestData struct {
	PRNumber int
	Owner    string
	Repo     string
	PR       *github.PullRequest
}

func ConnectClient() (*github.Client, context.Context) {
	ctx := context.Background()

	token := os.Getenv(GITHUB_TOKEN)
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc), ctx
}

func GetPullRequestData(client *github.Client, ctx context.Context) PullRequestData {
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
		PRNumber: prNumber,
		Owner:    owner,
		Repo:     repo,
		PR:       pr,
	}
}
