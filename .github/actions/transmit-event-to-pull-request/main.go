package main

import (
	"actions/commons/github"
	"fmt"
)

func main() {
	client, ctx := github.ConnectClient()

	prData := github.GetPullRequestData(client, ctx)

	Ids, err := github.GetJobIDsForPR(client, ctx, prData.PRNumber, prData.Owner, prData.Repo, *prData.PR.Head.SHA)
	if err != nil {
		fmt.Println(err, "Error while getting the job IDs for the PR")
		panic(err)
	}
	fmt.Println("Job IDs : ", Ids)

	id := Ids[0]

	fmt.Println("reRunJob")
	resp, err := github.ReRun(client, ctx, prData.Owner, prData.Repo, id)
	if err != nil {
		fmt.Println(err, "Error while re-running the job")
		panic(err)
	}
	fmt.Println("reRunJob resp : ", resp)

}
