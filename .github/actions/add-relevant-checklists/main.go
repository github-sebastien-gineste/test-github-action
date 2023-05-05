package main

import (
	"actions/commons/github"
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const TEMPLATE_CHECKLIST_PATH = "./templates/"

type Filter interface {
	Filter(string) bool
}

type FilterImp struct {
	regexAccept *regexp.Regexp
	regexReject *regexp.Regexp
}

func (filter FilterImp) Filter(filename string) bool {
	if filter.regexAccept != nil && filter.regexAccept.MatchString(filename) {
		return !(filter.regexReject != nil && filter.regexReject.MatchString(filename))
	}
	return false
}

type CheckList struct {
	Filename        string
	InclusionFilter Filter
	ExclusionFilter Filter
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
		Filename:        "proto_checklist.md",
		InclusionFilter: FilterImp{regexp.MustCompile(`\.proto$`), nil},
	}, {
		Filename:        "implementation_rpc_checklist.md",
		InclusionFilter: FilterImp{regexp.MustCompile(`Handler\.scala$`), nil},
	}, {
		Filename:        "development_conf_checklist.md",
		InclusionFilter: FilterImp{regexp.MustCompile(`\.conf$`), regexp.MustCompile(`(api-domains\.conf|api-domains-migrations\.conf)$`)},
	}, {
		Filename:        "production_conf_checklist.md",
		InclusionFilter: FilterImp{regexp.MustCompile(`(.*_bakery.*)|(api-domains\.conf|api-domains-migrations\.conf)$`), nil},
		ExclusionFilter: FilterImp{regexp.MustCompile(`\.sql$`), nil},
	}, {
		Filename:        "sql_migration_checklist.md",
		InclusionFilter: FilterImp{regexp.MustCompile(`\.sql$`), nil},
	},
}

func main() {

	client, ctx := github.ConnectClient()

	prData := github.GetPullRequestData(client, ctx)

	updatedPRBody, err := syncCheckLists(client, ctx, prData)
	if err != nil {
		fmt.Println(err, "Error while synchronising the checklists")
		panic(err)
	}

	if updatedPRBody == prData.PR.GetBody() {
		fmt.Println("\nNo changes required to the body of the PR")
		return
	}

	err = github.UpdatePRBody(client, ctx, prData.Owner, prData.Repo, prData.PR, updatedPRBody)
	if err != nil {
		fmt.Println(err, "Error while updating the PR body")
		panic(err)
	}

	fmt.Println("\nBody updated with success !")
}

func syncCheckLists(client *github.GithubClient, ctx context.Context, prData github.PullRequestData) (string, error) {
	currentPRBody := prData.PR.GetBody()
	filenames, err := github.GetDiffFilesNames(client, ctx, prData.Owner, prData.Repo, prData.PRNumber)
	if err != nil {
		fmt.Println(err, "Error while retrieving the files diff of the PR")
		return "", err
	}

	fmt.Println("There is ", len(filenames), " files in the diff of the PR")

	checkListsPlan, err := getCheckListsPlan(currentPRBody, filenames, false)
	if err != nil {
		fmt.Println(err, "Error while getting the plan of the checklists")
		return "", err
	}

	currentPRBody, err = applyCheckListsPlan(currentPRBody, checkListsPlan, false)
	if err != nil {
		fmt.Println(err, "Error while synchronising the checklists")
		return "", err
	}

	return currentPRBody, nil
}

func getCheckListsPlan(currentPRBody string, filenames []string, ignoreLog bool) ([]CheckListPlan, error) {
	if !ignoreLog {
		fmt.Println("\nPlan:")
	}

	checkListsPlan := []CheckListPlan{}

	for _, checkListItem := range allCheckLists {
		checkListItemPlan, err := getCheckListPlan(currentPRBody, checkListItem, filenames, ignoreLog)
		if err != nil {
			fmt.Println(err, "Error while getting the plan of the checklist item "+checkListItem.Filename)
			return []CheckListPlan{}, err
		}
		checkListsPlan = append(checkListsPlan, checkListItemPlan)
	}

	return checkListsPlan, nil
}

func getCheckListPlan(prBody string, checkListItem CheckList, filenames []string, ignoreLog bool) (CheckListPlan, error) {
	checkListTitle, err := getFirstLine(checkListItem.Filename)
	if err != nil {
		return CheckListPlan{}, err
	}
	isCheckListNeeded := isCheckListNeeded(checkListItem, filenames)
	isChecklistAlreadyPresent := strings.Contains(prBody, checkListTitle)
	checkListPlan := CheckListPlan{checkListItem.Filename, checkListTitle, IGNORED}

	if isChecklistAlreadyPresent && !isCheckListNeeded {
		checkListPlan.Action = REMOVED
	} else if isCheckListNeeded && !isChecklistAlreadyPresent {
		checkListPlan.Action = ADDED
	}

	printPlanLog(checkListItem, isCheckListNeeded, isChecklistAlreadyPresent, checkListPlan.Action, ignoreLog)
	return checkListPlan, nil
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
		if checkListItem.InclusionFilter.Filter(filename) {
			if checkListItem.ExclusionFilter != nil && checkListItem.ExclusionFilter.Filter(filename) {
				continue
			}
			isCheckListNeeded = true
			break
		}
	}
	return isCheckListNeeded
}

func applyCheckListsPlan(prBody string, checkListsPlan []CheckListPlan, ignoreLog bool) (string, error) {
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
