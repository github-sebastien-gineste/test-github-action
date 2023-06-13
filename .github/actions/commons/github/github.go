package github

import (
	"context"
	"errors"
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
const JOB_TO_RERUN = "JOB_TO_RERUN"

type GithubClient github.Client
type IssueComment github.IssueComment

type CommitFiles struct {
	Filename string
	Status   string
}

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

func GetCommitFiles(client *GithubClient, ctx context.Context, owner string, repo string, prNumber int) ([]CommitFiles, error) {
	opt := &github.ListOptions{
		PerPage: 100,
	}
	var commitFiles []CommitFiles

	for {
		files, response, err := client.PullRequests.ListFiles(ctx, owner, repo, prNumber, opt)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			commitFiles = append(commitFiles, CommitFiles{Filename: *file.Filename, Status: *file.Status})
		}

		if response.NextPage == 0 {
			break
		}
		opt.Page = response.NextPage
	}

	return commitFiles, nil
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
		issueComment := IssueComment(*comment)
		issueComments = append(issueComments, issueComment)
	}

	return issueComments, nil
}

func getListCheckRunsForRef(client *GithubClient, ctx context.Context, owner string, repo string, sha string) (*github.ListCheckRunsResults, error) {
	re, _, err := client.Checks.ListCheckRunsForRef(ctx, owner, repo, sha, nil)
	return re, err
}

func getJobIDByJobNameAndRef(client *GithubClient, ctx context.Context, owner string, repo string, sha string, jobName string) (int64, error) {
	checkRuns, err := getListCheckRunsForRef(client, ctx, owner, repo, sha)
	if err != nil {
		return -1, err
	}

	if len(checkRuns.CheckRuns) == 0 {
		return -1, errors.New("No check runs found for the pull request with the sha: " + sha)
	}

	for _, checkRun := range checkRuns.CheckRuns {
		if checkRun.GetName() == jobName {
			return checkRun.GetID(), nil
		}
	}

	return -1, errors.New("No jobs found for the pull request with the sha: " + sha)
}

func reRunJobById(client *GithubClient, ctx context.Context, owner string, repo string, jobID int64) (*github.Response, error) {
	u := fmt.Sprintf("repos/%v/%v/actions/jobs/%v/rerun", owner, repo, jobID)

	underlyingGithubClient := github.Client(*client)

	req, err := underlyingGithubClient.NewRequest("POST", u, nil)
	if err != nil {
		return nil, err
	}

	return underlyingGithubClient.Do(ctx, req, nil)
}

func ReRunJob(client *GithubClient, ctx context.Context, prData PullRequestData) error {
	jobName := os.Getenv(JOB_TO_RERUN)

	if jobName == "" {
		return errors.New("No job name provided")
	}

	jobId, err := getJobIDByJobNameAndRef(client, ctx, prData.Owner, prData.Repo, *prData.PR.Head.SHA, jobName)
	if err != nil {
		return errors.Join(err, errors.New("Error while getting the job ID for the job "+jobName+" in the PR "+strconv.Itoa(prData.PRNumber)))
	}

	fmt.Println("ReRun the Job ID ", jobId, ", in the ref : ", *prData.PR.Head.SHA)

	_, err = reRunJobById(client, ctx, prData.Owner, prData.Repo, jobId)
	if err != nil {
		return errors.Join(err, errors.New("Error while re-running the job"))
	}

	fmt.Println("Job Successfully ReRun")
	return nil
}

func String(s string) *string { return &s }
