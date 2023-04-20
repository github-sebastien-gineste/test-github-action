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
	Title          *string `json:"title,omitempty"`
	RegexBody      *string `json:"regex,omitempty"`
	RegexFileNames *string `json:"regex_file_names,omitempty"`
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

	currentBody := pr.GetBody()

	checkList := []CheckList{
		{
			Title:          stringPtr("proto_checklist.md"),
			RegexBody:      stringPtr(`^.*# Checklist for a proto PR.*$`),
			RegexFileNames: stringPtr(`^.*.proto$`),
		}, {
			Title:          stringPtr("implementation_rpc_checklist.md"),
			RegexBody:      stringPtr(`^.*# Checklist for a change in development configuration.*$`),
			RegexFileNames: stringPtr(`^.*Handler.scala$`),
		},
		{
			Title:          stringPtr("development_conf_checklist.md"),
			RegexBody:      stringPtr(`^.*# Checklist for an implementation PR.*$`),
			RegexFileNames: stringPtr(`^.*.conf$`), // TODO: remove the api-domains.conf in the detection
		},
		{
			Title:          stringPtr("production_conf_checklist.md"),
			RegexBody:      stringPtr(`^.*# Checklist for a change in production's configuration.*$`),
			RegexFileNames: stringPtr(`^.*api-domains.conf$`),
		},
		{
			Title:          stringPtr("sql_migration_checklist.md"),
			RegexBody:      stringPtr(`^.*# Checklist for a PR containing SQL migrations.*$`),
			RegexFileNames: stringPtr(`^.*.sql$`),
		},
	}

	for _, checkListItem := range checkList {
		if lineMatchesRegex(currentBody, regexp.MustCompile(*checkListItem.RegexBody)) {
			// check if we need to remove the checklist
			checklist_justify_presence := false
			for _, filename := range filenames {
				if lineMatchesRegex(filename, regexp.MustCompile(*checkListItem.RegexFileNames)) {
					checklist_justify_presence = true
				}
			}
			if !checklist_justify_presence {
				// remove the checklist
				newBody := ""
				step := 0
				for _, line := range strings.Split(currentBody, "\n") {
					if step == 0 && lineMatchesRegex(line, regexp.MustCompile(*checkListItem.RegexBody)) {
						step = 1
					} else if step == 1 && strings.HasPrefix(line, "#") {
						step = 2
						newBody += line + "\n"
					} else if step != 1 {
						newBody += line + "\n"
					}
				}
				currentBody = newBody
			}
		} else {
			// check if we need to add the checklist
			checklist_justify_presence := false
			for _, filename := range filenames {
				if lineMatchesRegex(filename, regexp.MustCompile(*checkListItem.RegexFileNames)) {
					checklist_justify_presence = true
				}
			}
			if checklist_justify_presence {
				// add the checklist
				currentBody += "\n" + getFileContent(*checkListItem.Title)
			}
		}
	}

	// ---- End ----
	fmt.Println("Fichiers modifiés :" + filesStr)

	updatePR(client, ctx, owner, repo, pr, currentBody)
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

func getFileContent(filename string) string {
	file, err := os.Open("../template/" + filename)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var bodyLines []string
	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}

	return strings.Join(bodyLines, "\n")
}
