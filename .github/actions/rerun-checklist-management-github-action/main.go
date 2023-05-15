package main

import (
	"actions/commons/github"
	"fmt"
)

const JOB_TO_SEARCH = "checklistsManagement"

func main() {
	client, ctx := github.ConnectClient()

	prData := github.GetPullRequestData(client, ctx)

	jobId, err := github.GetJobIDByJobNameAndRef(client, ctx, prData.Owner, prData.Repo, *prData.PR.Head.SHA, JOB_TO_SEARCH)
	if err != nil {
		fmt.Println(err, "Error while getting the job ID for the job "+JOB_TO_SEARCH+" in the PR")
		panic(err)
	}

	fmt.Println("ReRun the Job ID ", jobId)

	_, err = github.ReRun(client, ctx, prData.Owner, prData.Repo, jobId)
	if err != nil {
		fmt.Println(err, "Error while re-running the job")
		panic(err)
	}

	fmt.Println("Job Successfully ReRun")
}
