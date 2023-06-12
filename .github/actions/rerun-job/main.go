package main

import (
	"actions/commons/github"
	"os"
)

const JOB_TO_RERUN = "JOB_TO_RERUN"

func main() {
	client, ctx := github.ConnectClient()

	prData := github.GetPullRequestData(client, ctx)

	jobName := os.Getenv(JOB_TO_RERUN)

	err := github.ReRunJobByJobName(client, ctx, prData, jobName)
	if err != nil {
		panic(err)
	}
}
