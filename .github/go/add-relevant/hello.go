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

const PR_NUMBER = "PR_NUMBER"
const OWNER = "OWNER"
const REPO = "REPO"
const GITHUB_TOKEN = "GITHUB_TOKEN"
const TEMPLATE_CHECKLIST_PATH = "../template/"

type CheckList struct {
	Filename       *string
	Title          *string
	RegexDiffFiles *string
}

func initCheckList() []CheckList {
	return []CheckList{
		{
			Filename:       stringPtr("proto_checklist.md"),
			Title:          stringPtr(`# Checklist for a proto PR`),
			RegexDiffFiles: stringPtr(`\.proto$`),
		}, {
			Filename:       stringPtr("implementation_rpc_checklist.md"),
			Title:          stringPtr(`# Checklist for a change in development configuration`),
			RegexDiffFiles: stringPtr(`Handler\.scala$`),
		}, {
			Filename:       stringPtr("development_conf_checklist.md"),
			Title:          stringPtr(`# Checklist for an implementation PR`),
			RegexDiffFiles: stringPtr(`\.conf$`),
		}, {
			Filename:       stringPtr("production_conf_checklist.md"),
			Title:          stringPtr(`# Checklist for a change in production's configuration`),
			RegexDiffFiles: stringPtr(`^api-domains.conf$`),
		}, {
			Filename:       stringPtr("sql_migration_checklist.md"),
			Title:          stringPtr(`# Checklist for a PR containing SQL migrations`),
			RegexDiffFiles: stringPtr(`\.sql$`),
		},
	}
}

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

	filenames := getDiffFiles(client, ctx, owner, repo, prNumber)
	filesStr := strings.Join(filenames, "\n")

	// ---- Start ----

	// TODO - see test
	// TODO - if modif in go folder -> CI/CD test

	currentBody := pr.GetBody()
	checkList := initCheckList()

	for _, checkListItem := range checkList {
		currentBody = manageCheckListItem(currentBody, checkListItem, filenames)
	}

	// ---- End ----
	fmt.Println("Fichiers modifiés :" + filesStr)

	updatePR(client, ctx, owner, repo, pr, currentBody)
	fmt.Println("Body updated with success !")
}

type StateRemoveCheckList int64

const (
	SearchCheckList StateRemoveCheckList = iota
	RemoveCheckList
	CopyRest
)

func manageCheckListItem(prbody string, checkListItem CheckList, filenames []string) string {
	is_checklist_justify_presence := checkNeedCheckListItem(checkListItem, filenames)
	is_checklist_already_present := lineMatchesRegex(prbody, `^`+*checkListItem.Title+`.*$`)

	if is_checklist_already_present {
		if !is_checklist_justify_presence {
			prbody = removeCheckList(prbody, checkListItem)
		}
	} else if is_checklist_justify_presence {
		prbody += "\n" + getFileContent(*checkListItem.Filename)
	}

	return prbody
}

func checkNeedCheckListItem(checkListItem CheckList, filenames []string) bool {
	is_checklist_justify_presence := false
	for _, filename := range filenames {
		if lineMatchesRegex(filename, *checkListItem.RegexDiffFiles) {
			is_checklist_justify_presence = true
		}
	}
	return is_checklist_justify_presence
}

func removeCheckList(prbody string, checkListItem CheckList) string {
	// remove the checklist
	newBody := ""
	stateRemoveCheckList := SearchCheckList
	for _, line := range strings.Split(prbody, "\n") {
		switch stateRemoveCheckList {
		case SearchCheckList:
			if lineMatchesRegex(line, `^`+*checkListItem.Title+`.*$`) {
				stateRemoveCheckList = RemoveCheckList
			} else {
				newBody += line + "\n"
			}
		case RemoveCheckList:
			if strings.HasPrefix(line, "#") {
				stateRemoveCheckList = CopyRest
				newBody += line + "\n"
			}
		case CopyRest:
			newBody += line + "\n"
		}
	}
	return newBody
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

func getFileContent(filename string) string {
	file, err := os.Open(TEMPLATE_CHECKLIST_PATH + filename)
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
