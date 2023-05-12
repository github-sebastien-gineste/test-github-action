package github

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

type GithubClient github.Client
type IssueComment *github.IssueComment

type PullRequestData struct {
	PRNumber int
	Owner    string
	Repo     string
	PR       *github.PullRequest
}

func ConnectClient() (*GithubClient, context.Context) {
	ctx := context.Background()

	token := os.Getenv(GITHUB_TOKEN)
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	githubClient := GithubClient(*github.NewClient(tc))

	return &githubClient, ctx
}

func GetPullRequestData(client *GithubClient, ctx context.Context) PullRequestData {
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

func GetDiffFilesNames(client *GithubClient, ctx context.Context, owner string, repo string, prNumber int) ([]string, error) {
	opt := &github.ListOptions{
		PerPage: 100,
	}
	var filenames []string

	for {
		files, response, err := client.PullRequests.ListFiles(ctx, owner, repo, prNumber, opt)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			filenames = append(filenames, *file.Filename)
		}

		if response.NextPage == 0 {
			break
		}
		opt.Page = response.NextPage
	}

	return filenames, nil
}

func UpdatePRBody(client *GithubClient, ctx context.Context, owner string, repo string, pr *github.PullRequest, newbody string) error {
	updatedPR := &github.PullRequest{
		Title: pr.Title,
		Body:  github.String(newbody),
		State: pr.State,
		Base:  pr.Base,
	}

	_, _, err := client.PullRequests.Edit(ctx, owner, repo, pr.GetNumber(), updatedPR)
	if err != nil {
		return err
	}
	return nil
}

func GetListPRComments(client *GithubClient, ctx context.Context, owner string, repo string, pr *github.PullRequest) ([]IssueComment, error) {
	comments, _, err := client.Issues.ListComments(ctx, owner, repo, pr.GetNumber(), &github.IssueListCommentsOptions{})
	if err != nil {
		return []IssueComment{}, err
	}

	var issueComments []IssueComment

	for _, comment := range comments {
		issueComment := IssueComment(comment)
		issueComments = append(issueComments, issueComment)
	}

	return issueComments, nil
}

func CreateRepoStatue(client *GithubClient, ctx context.Context, owner string, repo string, pr *github.PullRequest, status string, description string) {

	// Création du statut du check
	statusInput := &github.RepoStatus{
		State:       github.String(status),
		Description: github.String(description),
		Context:     github.String("Checkbox check"),
	}

	fmt.Println("Create statucs: ", statusInput)

	_, _, err := client.Repositories.CreateStatus(ctx, owner, repo, *pr.Head.SHA, statusInput)
	if err != nil {
		panic(err)
	}
}

func CreateCheckRun(client *GithubClient, ctx context.Context, owner string, repo string, sha string) {

	// Crée une check run in progress
	opt := github.CreateCheckRunOptions{
		Name:    "My check",
		HeadSHA: sha,
		Status:  github.String("in_progress"),
		Output: &github.CheckRunOutput{
			Title:   github.String("Check in progress"),
			Summary: github.String("The check is in progress..."),
		},
	}

	checkRun, response, err := client.Checks.CreateCheckRun(ctx, owner, repo, opt)
	if err != nil {
		fmt.Println("Error creating check run:", err)
		os.Exit(1)
	}

	fmt.Println(response)
	fmt.Println(checkRun)
	fmt.Println("\n\n")

	result, resp, err := client.Checks.ListCheckSuitesForRef(ctx, owner, repo, sha, nil)

	fmt.Println(result)
	fmt.Println(resp)
	fmt.Println("\n\n")

	re, resp, err := client.Checks.ListCheckRunsForRef(ctx, owner, repo, sha, nil)

	fmt.Println(re)
	fmt.Println(resp)

}
