package main

import (
	"actions/commons"
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/github"
)

const TEMPLATE_CHECKLIST_PATH = "./templates/"

type CheckList struct {
	Filename              string
	RegexDiffFiles        *regexp.Regexp
	RegexNegatifDiffFiles *regexp.Regexp
}

type CheckListPlan struct {
	Filename string
	Title    string
	Action   PlanAction
}

type StateRemoveCheckList int64

const (
	SearchCheckList StateRemoveCheckList = iota
	RemoveCheckList
	CopyRest
)

type PlanAction string

const (
	ADDED   PlanAction = "ADDED"
	IGNORED PlanAction = "IGNORED"
	REMOVED PlanAction = "REMOVED"
)

var allCheckLists = []CheckList{
	{
		Filename:       "proto_checklist.md",
		RegexDiffFiles: regexp.MustCompile(`\.proto$`),
	}, {
		Filename:       "implementation_rpc_checklist.md",
		RegexDiffFiles: regexp.MustCompile(`Handler\.scala$`),
	}, {
		Filename:              "development_conf_checklist.md",
		RegexDiffFiles:        regexp.MustCompile(`\.conf$`),
		RegexNegatifDiffFiles: regexp.MustCompile(`api-domains.conf$`),
	}, {
		Filename:       "production_conf_checklist.md",
		RegexDiffFiles: regexp.MustCompile(`api-domains.conf$`),
	}, {
		Filename:       "sql_migration_checklist.md",
		RegexDiffFiles: regexp.MustCompile(`\.sql$`),
	},
}

func main() {

	client, ctx := commons.ConnectClient()

	prData := commons.GetPullRequestData(client, ctx)

	updatedPRBody, err := syncCheckLists(client, ctx, prData)
	if err != nil {
		fmt.Println(err, "Error while synchronising the checklists")
		panic(err)
	}

	err = updatePRBody(client, ctx, prData.Owner, prData.Repo, prData.PR, updatedPRBody)
	if err != nil {
		fmt.Println(err, "Error while updating the PR body")
		panic(err)
	}

	fmt.Println("Body updated with success !")
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

func syncCheckLists(client *github.Client, ctx context.Context, prData commons.PullRequestData) (string, error) {
	currentPRBody := prData.PR.GetBody()
	filenames, err := getDiffFilesNames(client, ctx, prData.Owner, prData.Repo, prData.PRNumber)
	if err != nil {
		fmt.Println(err, "Error while retrieving the files diff of the PR")
		return "", err
	}

	CheckListsPlan, err := getPlanCheckLists(currentPRBody, filenames, false)
	if err != nil {
		fmt.Println(err, "Error while getting the plan of the checklists")
		return "", err
	}

	currentPRBody, err = applyPlanCheckLists(currentPRBody, CheckListsPlan, false)
	if err != nil {
		fmt.Println(err, "Error while synchronising the checklists")
		return "", err
	}

	return currentPRBody, nil
}

func getPlanCheckLists(currentPRBody string, filenames []string, ignoreLog bool) ([]CheckListPlan, error) {
	if !ignoreLog {
		fmt.Println("Plan:")
	}

	var CheckListsPlan = []CheckListPlan{}

	for _, checkListItem := range allCheckLists {
		checkListItemPlan, err := getPlanCheckList(currentPRBody, checkListItem, filenames, ignoreLog)
		if err != nil {
			fmt.Println(err, "Error while getting the plan of the checklist item "+checkListItem.Filename)
			return []CheckListPlan{}, err
		}
		CheckListsPlan = append(CheckListsPlan, checkListItemPlan)
	}

	return CheckListsPlan, nil
}

func getPlanCheckList(prBody string, checkListItem CheckList, filenames []string, ignoreLog bool) (CheckListPlan, error) {
	checkListTitle, err := getFirstLine(checkListItem.Filename)
	if err != nil {
		return CheckListPlan{}, err
	}
	isCheckListNeeded := isCheckListNeeded(checkListItem, filenames)
	isChecklistAlreadyPresent := strings.Contains(prBody, checkListTitle)

	if isChecklistAlreadyPresent && !isCheckListNeeded {
		printPlanLog(checkListItem, isCheckListNeeded, isChecklistAlreadyPresent, REMOVED, ignoreLog)
		return CheckListPlan{checkListItem.Filename, checkListTitle, REMOVED}, nil
	} else if isCheckListNeeded && !isChecklistAlreadyPresent {
		printPlanLog(checkListItem, isCheckListNeeded, isChecklistAlreadyPresent, ADDED, ignoreLog)
		return CheckListPlan{checkListItem.Filename, "", ADDED}, nil
	} else {
		printPlanLog(checkListItem, isCheckListNeeded, isChecklistAlreadyPresent, IGNORED, ignoreLog)
		return CheckListPlan{checkListItem.Filename, "", IGNORED}, nil
	}
}

func printPlanLog(checkListItem CheckList, isCheckListNeeded bool, isChecklistAlreadyPresent bool, decision PlanAction, ignoreLog bool) {
	if ignoreLog {
		return
	}

	logText := "- Checklist " + checkListItem.Filename + " : "

	if isCheckListNeeded {
		logText += "needed and "
	} else {
		logText += "not needed and "
	}
	if isChecklistAlreadyPresent {
		logText += "present"
	} else {
		logText += "not present"
	}

	fmt.Println(logText + " => TO BE " + string(decision))
}

func isCheckListNeeded(checkListItem CheckList, filenames []string) bool {
	isCheckListNeeded := false
	for _, filename := range filenames {
		if checkListItem.RegexDiffFiles.MatchString(filename) {
			if checkListItem.RegexNegatifDiffFiles != nil && checkListItem.RegexNegatifDiffFiles.MatchString(filename) {
				continue
			}
			isCheckListNeeded = true
			break
		}
	}
	return isCheckListNeeded
}

func applyPlanCheckLists(prBody string, checkListsPlan []CheckListPlan, ignoreLog bool) (string, error) {
	if !ignoreLog {
		fmt.Println("\nApply:")
	}
	prBodyUpdated := prBody
	nbIgnored := 0

	for _, checkListItemPlan := range checkListsPlan {
		switch checkListItemPlan.Action {
		case ADDED:
			if !ignoreLog {
				fmt.Println("- Adding checklist " + checkListItemPlan.Filename)
			}
			newBody, err := addCheckList(prBodyUpdated, checkListItemPlan.Filename)
			if err != nil {
				fmt.Println(err, "Error while adding the checklist item "+checkListItemPlan.Filename)
				return "", err
			}
			prBodyUpdated = newBody
		case REMOVED:
			if !ignoreLog {
				fmt.Println("- Removing checklist " + checkListItemPlan.Filename)
			}
			prBodyUpdated = removeCheckList(prBodyUpdated, checkListItemPlan.Title)
		case IGNORED:
			nbIgnored++
		default:
			return "", errors.New("Unknown action " + string(checkListItemPlan.Action))
		}
	}

	if nbIgnored == len(checkListsPlan) && !ignoreLog {
		fmt.Println("Nothing to do")
	}

	return prBodyUpdated, nil
}

func addCheckList(prBody string, filename string) (string, error) {
	content, err := getFileContent(filename)
	if err != nil {
		return "", err
	}
	return prBody + "\n" + content, nil
}

func removeCheckList(prbody string, checkListTitle string) string {
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
