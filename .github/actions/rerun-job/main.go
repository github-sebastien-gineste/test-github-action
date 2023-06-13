package main

import (
	"actions/commons/github"
)

func main() {
	client, ctx := github.ConnectClient()

	prData := github.GetPullRequestData(client, ctx)

	err := github.ReRunJob(client, ctx, prData)
	if err != nil {
		panic(err)
	}
}
