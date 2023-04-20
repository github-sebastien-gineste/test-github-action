package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type CheckList struct {
	Title   *string `json:"title,omitempty"`
	Regex   *string `json:"regex,omitempty"`
	Is_Used *bool   `json:"is_used,omitempty"`
}

func main() {
	// Récupérer le token d'authentification depuis les secrets de Github
	ctx := context.Background()
	client := connectClient(ctx)
	// Récupérer les informations de la Pull Request
	prNumberStr := os.Getenv("PR_NUMBER")
	prNumber, err := strconv.Atoi(prNumberStr)
	owner := os.Getenv("OWNER")
	repo := os.Getenv("REPO")
	pr, _, err := client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		fmt.Println(err)
		return
	}

	filenames := getDiffFiles(client, ctx, owner, repo, prNumber)
	filesStr := strings.Join(filenames, "\n")

	// ---- Start ----

	checkList := []CheckList{
		{Title: stringPtr("proto_checklist.md"), Regex: stringPtr(`^.*# Checklist for a proto PR.*$`), Is_Used: boolPtr(false)},
		{Title: stringPtr("development_conf_checklist"), Regex: stringPtr(`^.*# Checklist for a change in development configuration.*$`), Is_Used: boolPtr(false)},
		{Title: stringPtr("implementation_rpc_checklist"), Regex: stringPtr(`^.*# Checklist for an implementation PR.*$`), Is_Used: boolPtr(false)},
		{Title: stringPtr("production_conf_checklist"), Regex: stringPtr(`^.*# Checklist for a change in production's configuration.*$`), Is_Used: boolPtr(false)},
		{Title: stringPtr("sql_migration_checklist"), Regex: stringPtr(`^.*# Checklist for a PR containing SQL migrations.*$`), Is_Used: boolPtr(false)},
	}

	for _, checkListItem := range checkList {
		if lineMatchesRegex(pr.GetBody(), regexp.MustCompile(*checkListItem.Regex)) {
			*checkListItem.Is_Used = true // The checklist is already used
		}
		println(*checkListItem.Title + " : " + strconv.FormatBool(*checkListItem.Is_Used))
	}

	// search regex pattern in the pr body

	r := regexp.MustCompile(`^.*# Checklist for a proto PR.*$`)
	if lineMatchesRegex(pr.GetBody(), r) {
		fmt.Println("Le string contient une ligne qui correspond au regex")
	} else {
		fmt.Println("Le string ne contient pas de ligne qui correspond au regex")
	}

	// ---- End ----
	fmt.Println("Fichiers modifiés :" + filesStr)

	// Lire le contenu du fichier check.md
	file, err := os.Open("../template/proto_checklist.md")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var bodyLines []string
	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}
	body := strings.Join(bodyLines, "\n")

	updatePR(client, ctx, owner, repo, pr, body)
	fmt.Println("Body updated with success !")
}

func connectClient(ctx context.Context) *github.Client {
	token := os.Getenv("GITHUB_TOKEN")

	// Créer un client Github avec l'authentification
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func getDiffFiles(client *github.Client, ctx context.Context, owner string, repo string, prNumber int) []string {
	files, _, err := client.PullRequests.ListFiles(ctx, owner, repo, prNumber, nil)
	if err != nil {
		fmt.Println(err)
		return []string{}
	}

	var filenames []string
	for _, file := range files {
		filenames = append(filenames, *file.Filename)
	}

	return filenames
}

func updatePR(client *github.Client, ctx context.Context, owner string, repo string, pr *github.PullRequest, newbody string) {
	// Mettre à jour le corps de la Pull Request avec  le contenu du fichier  check.md
	updatedPR := &github.PullRequest{
		Title: pr.Title,
		Body:  github.String(newbody),
		State: pr.State,
		Base:  pr.Base,
	}

	_, _, err := client.PullRequests.Edit(ctx, owner, repo, pr.GetNumber(), updatedPR)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func lineMatchesRegex(s string, r *regexp.Regexp) bool {
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		if r.MatchString(line) {
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
