package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const PR_NUMBER = "PR_NUMBER"
const OWNER = "OWNER"
const REPO = "REPO"
const GITHUB_TOKEN = "GITHUB_TOKEN"

func main() {
	// Récupérer le token d'authentification depuis les secrets de Github
	ctx := context.Background()
	client := connectClient(ctx)
	// Récupérer les informations de la Pull Request
	prNumberStr := os.Getenv(PR_NUMBER)
	prNumber, err := strconv.Atoi(prNumberStr)
	owner := os.Getenv(OWNER)
	repo := os.Getenv(REPO)

	pr, _, err := client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		fmt.Println(err)
		return
	}

	prBody := pr.GetBody()

	// Vérifier si la checklist est cochée
	if lineMatchesRegex(prBody, `^- \[ \]`) {
		fmt.Println("Checklist already checked !")
		return
	}

	client.Checks.UpdateCheckRun(ctx, owner, repo, 0, github.UpdateCheckRunOptions{})

	// https://github.com/actions/starter-workflows/issues/292

	// ---- End ----

	fmt.Println("Body updated with success !")
}

func connectClient(ctx context.Context) *github.Client {
	token := os.Getenv(GITHUB_TOKEN)

	// Créer un client Github avec l'authentification
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func lineMatchesRegex(s string, regex string) bool {
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		if regexp.MustCompile(regex).MatchString(line) {
			return true
		}
	}
	return false
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
