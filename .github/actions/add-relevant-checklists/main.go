package main

import (
	"bufio"
	"context"
	"errors"
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
const TEMPLATE_CHECKLIST_PATH = "../templates/checklists/"

const (
	SearchCheckList StateRemoveCheckList = iota
	RemoveCheckList
	CopyRest
)

type StateRemoveCheckList int64

type CheckList struct {
	Filename       string
	RegexDiffFiles *regexp.Regexp
}

var allCheckLists = []CheckList{
	{
		Filename:       "proto_checklist.md",
		RegexDiffFiles: regexp.MustCompile(`\.proto$`),
	}, {
		Filename:       "implementation_rpc_checklist.md",
		RegexDiffFiles: regexp.MustCompile(`Handler\.scala$`),
	}, {
		Filename:       "development_conf_checklist.md",
		RegexDiffFiles: regexp.MustCompile(`\.conf$`),
	}, {
		Filename:       "production_conf_checklist.md",
		RegexDiffFiles: regexp.MustCompile(`^api-domains.conf$`),
	}, {
		Filename:       "sql_migration_checklist.md",
		RegexDiffFiles: regexp.MustCompile(`\.sql$`),
	},
}

func main() {
	ctx := context.Background()
	client := connectClient(ctx)

	// Retrieve information from the Pull Request
	prNumberStr := os.Getenv(PR_NUMBER)
	prNumber, err := strconv.Atoi(prNumberStr)
	owner := os.Getenv(OWNER)
	repo := os.Getenv(REPO)
	pr, _, err := client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		fmt.Println(err, "Error while retrieving the PR informations")
		panic(err)
	}
	currentPRBody := pr.GetBody()
	filenames, err := getDiffFilesNames(client, ctx, owner, repo, prNumber)
	if err != nil {
		fmt.Println(err, "Error while retrieving the files diff of the PR")
		panic(err)
	}

	path, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(path) // for example /home/user

	path, err = os.Executable()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(path)

	// Makes the changes
	for _, checkListItem := range allCheckLists {
		currentPRBody, err = syncCheckList(currentPRBody, checkListItem, filenames)
		if err != nil {
			fmt.Println(err, "Error while synchronising the checklist item "+checkListItem.Filename)
			panic(err)
		}
	}
	err = updatePRBody(client, ctx, owner, repo, pr, currentPRBody)
	if err != nil {
		fmt.Println(err, "Error while updating the PR body")
		panic(err)
	}
	fmt.Println("Body updated with success !")
}

func connectClient(ctx context.Context) *github.Client {
	token := os.Getenv(GITHUB_TOKEN)
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func getDiffFilesNames(client *github.Client, ctx context.Context, owner string, repo string, prNumber int) ([]string, error) {
	files, _, err := client.PullRequests.ListFiles(ctx, owner, repo, prNumber, nil)
	if err != nil {
		return nil, err
	}

	var filenames []string
	for _, file := range files {
		filenames = append(filenames, *file.Filename)
	}
	return filenames, nil
}

func syncCheckList(prBody string, checkListItem CheckList, filenames []string) (string, error) {
	checkListTitle, err := getFirstLine(checkListItem.Filename)
	if err != nil {
		return "", err
	}
	isCheckListNeeded := isCheckListNeeded(checkListItem, filenames)
	isChecklistAlreadyPresent := strings.Contains(prBody, checkListTitle)

	if isChecklistAlreadyPresent {
		if !isCheckListNeeded {
			prBody = removeCheckList(prBody, checkListItem, checkListTitle)
		}
	} else if isCheckListNeeded {
		content, err := getFileContent(checkListItem.Filename)
		if err != nil {
			return "", err
		}
		prBody += "\n" + content
	}

	return prBody, nil
}

func isCheckListNeeded(checkListItem CheckList, filenames []string) bool {
	isCheckListNeeded := false
	for _, filename := range filenames {
		if checkListItem.RegexDiffFiles.MatchString(filename) {
			isCheckListNeeded = true
			break
		}
	}
	return isCheckListNeeded
}

func removeCheckList(prbody string, checkListItem CheckList, checkListTitle string) string {
	newBody := ""
	stateRemoveCheckList := SearchCheckList
	for _, line := range strings.Split(prbody, "\n") {
		switch stateRemoveCheckList {
		case SearchCheckList:
			if strings.Contains(line, checkListTitle) {
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

func getFileContent(filename string) (string, error) {
	file, err := os.Open(TEMPLATE_CHECKLIST_PATH + filename)
	if err != nil {
		return "", errors.New("The file " + TEMPLATE_CHECKLIST_PATH + filename + " does not exist")
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var bodyLines []string
	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}

	return strings.Join(bodyLines, "\n"), nil
}

func updatePRBody(client *github.Client, ctx context.Context, owner string, repo string, pr *github.PullRequest, newbody string) error {
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

func getFirstLine(filename string) (string, error) {
	file, err := os.Open(TEMPLATE_CHECKLIST_PATH + filename)
	if err != nil {
		return "", errors.New("The file " + TEMPLATE_CHECKLIST_PATH + filename + " does not exist")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return scanner.Text(), nil
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", errors.New("empty file")
}
