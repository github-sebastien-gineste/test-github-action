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
const EVENT_NAME = "EVENT_NAME"

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

func CreateCheckRun(client *GithubClient, ctx context.Context, owner string, repo string, sha string, conclusion string, details string) {

	suiteReq := github.CreateCheckSuiteOptions{
		HeadSHA: sha,
	}

	// Create check suite
	suite, _, err := client.Checks.CreateCheckSuite(ctx, owner, repo, suiteReq)
	if err != nil {
		fmt.Printf("Error creating check suite: %v\n", err)
		return
	}

	fmt.Println(suite)

	// Crée une check run in progress
	opt := github.CreateCheckRunOptions{
		Name:       "Comments checkboxes",
		HeadSHA:    sha,
		Conclusion: github.String(conclusion),
		Output: &github.CheckRunOutput{
			Title:   github.String("Comments checkboxes"),
			Summary: github.String("Ensure that all checkboxes in comment are checked"),
			Text:    github.String(details),
		},
	}

	checkRun, response, err := client.Checks.CreateCheckRun(ctx, owner, repo, opt)
	if err != nil {
		fmt.Println("Error creating check run:", err)
		os.Exit(1)
	}

	fmt.Println(response)
	fmt.Println(checkRun)
	fmt.Print("\n\n")

}

func GetListChekRunsForRef(client *GithubClient, ctx context.Context, owner string, repo string, sha string) (*github.ListCheckRunsResults, error) {
	re, resp, err := client.Checks.ListCheckRunsForRef(ctx, owner, repo, sha, nil)

	fmt.Println(re)
	fmt.Println(resp)

	return re, err
}

func GetJobIDsForPR(client *GithubClient, ctx context.Context, prNumber int, owner string, repo string, sha string) ([]int64, error) {

	GetListChekRunsForRef(client, ctx, owner, repo, sha)

	//opt := &github.ListCheckRunsOptions{CheckName: github.String("checklistsManagement")}
	checkRuns, _, err := client.Checks.ListCheckRunsForRef(ctx, owner, repo, sha, nil)
	if err != nil {
		return nil, err
	}

	if len(checkRuns.CheckRuns) == 0 {
		return nil, fmt.Errorf("No check runs found for pull request %d", prNumber)
	}

	jobIds := make([]int64, 0)
	for _, checkRun := range checkRuns.CheckRuns {
		if checkRun.GetName() != "checklistsManagement" {
			jobIds = append(jobIds, checkRun.GetID())
		}
	}

	if len(jobIds) == 0 {
		return nil, fmt.Errorf("No jobs found for pull request %d", prNumber)
	}

	return jobIds, nil
}

func ReRun(client *GithubClient, ctx context.Context, owner string, repo string, jobID int64) (*github.Response, error) {

	eVENT_NAME := os.Getenv(EVENT_NAME)

	if eVENT_NAME == "pull_request" {
		fmt.Println("ReRun pull_request")
		return nil, nil
	}

	u := fmt.Sprintf("repos/%v/%v/actions/jobs/%v/rerun", owner, repo, jobID)

	GithubClient := github.Client(*client)

	req, err := GithubClient.NewRequest("POST", u, nil)
	if err != nil {
		return nil, err
	}

	return GithubClient.Do(ctx, req, nil)
}
