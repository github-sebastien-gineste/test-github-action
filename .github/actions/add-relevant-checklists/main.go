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

type FileFilter interface {
	Filter(string) bool
}

type RegexFilenameFilter struct {
	// return *true* for the Filter function if this regex match
	regexAccept *regexp.Regexp
	// return *false* for the Filter function if this regex match
	regexReject *regexp.Regexp // -> Take precedence over regexAccept
}

func NewFilenameMatchesFilter(accepts ...string) RegexFilenameFilter {
	if len(accepts) == 0 {
		accepts = []string{".*"}
	}
	return RegexFilenameFilter{regexp.MustCompile(strings.Join(accepts, "|")), nil}
}

func (filter RegexFilenameFilter) exclude(excludeFilenames ...string) RegexFilenameFilter {
	filter.regexReject = regexp.MustCompile(strings.Join(excludeFilenames, "|"))
	return filter
}

func (filter RegexFilenameFilter) Filter(filename string) bool {
	if filter.regexAccept.MatchString(filename) {
		return !(filter.regexReject != nil && filter.regexReject.MatchString(filename))
	}
	return false
}

type FileStatusFilter struct {
	// return *true* for the Filter function if the file status match the list
	acceptedList []string
	// return *false* for the Filter function if the file status match the list
	rejectedList []string // -> Take precedence over acceptedList
}

func NewFileStatusMatchesFilter(status ...string) FileStatusFilter {
	if len(status) == 0 {
		status = []string{"*"}
	}
	return FileStatusFilter{status, nil}
}

func (filter FileStatusFilter) exclude(excludeFileStatus ...string) FileStatusFilter {
	filter.rejectedList = excludeFileStatus
	return filter
}

func (filter FileStatusFilter) Filter(status string) bool {
	if filter.acceptedList == nil && filter.rejectedList == nil {
		return true
	}
	if !isValidFileStatus(status) {
		fmt.Println("The file status " + status + " is not valid")
		return false
	}
	if filter.rejectedList != nil {
		for _, rejectedStatus := range filter.rejectedList {
			if status == rejectedStatus {
				return false
			}
		}
	}
	if filter.acceptedList != nil {
		if len(filter.acceptedList) == 1 && filter.acceptedList[0] == "*" {
			return true
		}
		for _, includedStatus := range filter.acceptedList {
			if status == includedStatus {
				return true
			}
		}
	}

	return false
}

func isValidFileStatus(changeType string) bool {
	allowedTypes := []string{"added", "removed", "modified", "renamed", "copied", "changed", "unchanged"}

	for _, t := range allowedTypes {
		if changeType == t {
			return true
		}
	}

	return false
}

type CheckList struct {
	Filename string
	// If one filename on the PR diff return true on its Filter() function, the checklist will be included
	FilenameFilter FileFilter
	// if one filename on the PR diff return false on its Filter() function, the checklist will not be included
	GlobalFilenameFilter FileFilter // -> Take precedence over the FilenameFilter
	// if this filter is defined, the file is already Included by the FilenameFilter and the file status doesn't match this filter, the checklist will not be included
	FileStatusFilter FileFilter
}

type CheckListPlan struct {
	Filename string
	Title    string
	Action   PlanAction
	Present  bool
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

const protoCheckListFilename = "proto_checklist.md"

var allCheckLists = []CheckList{
	{
		Filename:         "proto_creation_checklist.md",
		FilenameFilter:   NewFilenameMatchesFilter(`^(domains|framework)\/.*\/src\/main\/protobuf\/.*\.proto$`),
		FileStatusFilter: NewFileStatusMatchesFilter("added"),
	}, {
		Filename:         "proto_update_checklist.md",
		FilenameFilter:   NewFilenameMatchesFilter(`^(domains|framework)\/.*\/src\/main\/protobuf\/.*\.proto$`),
		FileStatusFilter: NewFileStatusMatchesFilter().exclude("added"),
	}, {
		Filename:       "implementation_rpc_checklist.md",
		FilenameFilter: NewFilenameMatchesFilter(`Handler\.scala$`),
	}, {
		Filename:       "development_conf_checklist.md",
		FilenameFilter: NewFilenameMatchesFilter(`\.conf$`).exclude(`api-domains\.conf$`, `api-domains-migrations\.conf$`, `^(domains|framework)\/.*\/src\/main\/resources\/reference\.conf$`),
	}, {
		Filename:             "production_conf_checklist.md",
		FilenameFilter:       NewFilenameMatchesFilter(`(.*_bakery.*)`, `api-domains\.conf$`, `api-domains-migrations\.conf$`, `^(domains|framework)\/.*\/src\/main\/resources\/reference\.conf$`),
		GlobalFilenameFilter: NewFilenameMatchesFilter().exclude(`\.sql$`),
	}, {
		Filename:       "sql_migration_checklist.md",
		FilenameFilter: NewFilenameMatchesFilter(`\.sql$`),
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
	files, err := github.GetDiffCommitFiles(client, ctx, prData.Owner, prData.Repo, prData.PRNumber)
	if err != nil {
		fmt.Println(err, "Error while retrieving the files diff of the PR")
		return "", err
	}

	fmt.Println("There is ", len(files), " files in the diff of the PR")

	checkListsPlan, err := getCheckListsPlan(currentPRBody, files, false)
	if err != nil {
		fmt.Println(err, "Error while getting the plan of the checklists")
		return "", err
	}

	checkListsPlan, err = filterProtoChecklistPlan(checkListsPlan, currentPRBody, false)
	if err != nil {
		fmt.Println(err, "Error while filtering the plan of the checklists")
		return "", err
	}

	currentPRBody, err = applyCheckListsPlan(currentPRBody, checkListsPlan, false)
	if err != nil {
		fmt.Println(err, "Error while synchronising the checklists")
		return "", err
	}

	return currentPRBody, nil
}

func getCheckListsPlan(currentPRBody string, files []github.CommitFiles, ignoreLog bool) ([]CheckListPlan, error) {
	if !ignoreLog {
		fmt.Println("\nPlan:")
	}

	checkListsPlan := []CheckListPlan{}

	for _, checkListItem := range allCheckLists {
		checkListItemPlan, err := getCheckListPlan(currentPRBody, checkListItem, files, ignoreLog)
		if err != nil {
			fmt.Println(err, "Error while getting the plan of the checklist item "+checkListItem.Filename)
			return []CheckListPlan{}, err
		}
		checkListsPlan = append(checkListsPlan, checkListItemPlan)
	}

	return checkListsPlan, nil
}

func getCheckListPlan(prBody string, checkListItem CheckList, files []github.CommitFiles, ignoreLog bool) (CheckListPlan, error) {
	checkListTitle, err := getFirstLine(checkListItem.Filename)
	if err != nil {
		return CheckListPlan{}, err
	}
	isCheckListNeeded := isCheckListNeeded(checkListItem, files)
	isChecklistAlreadyPresent := strings.Contains(prBody, checkListTitle)
	checkListPlan := CheckListPlan{checkListItem.Filename, checkListTitle, IGNORED, isChecklistAlreadyPresent}

	if isChecklistAlreadyPresent && !isCheckListNeeded {
		checkListPlan.Action = REMOVED
	} else if isCheckListNeeded && !isChecklistAlreadyPresent {
		checkListPlan.Action = ADDED
	}

	printPlanLog(checkListItem.Filename, isCheckListNeeded, isChecklistAlreadyPresent, checkListPlan.Action, ignoreLog)
	return checkListPlan, nil
}

func printPlanLog(CheckListFilename string, isCheckListNeeded bool, isChecklistAlreadyPresent bool, decision PlanAction, ignoreLog bool) {
	if ignoreLog {
		return
	}

	logText := "- Checklist " + CheckListFilename + " : "

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

func isCheckListNeeded(checkListItem CheckList, files []github.CommitFiles) bool {
	isGlobalFilenameFilterPresent := checkListItem.GlobalFilenameFilter != nil
	isStatusFilterPresent := checkListItem.FileStatusFilter != nil
	isCheckListNeeded := false

	for _, file := range files {
		if isGlobalFilenameFilterPresent && !checkListItem.GlobalFilenameFilter.Filter(file.Filename) {
			isCheckListNeeded = false
			break
		} else if checkListItem.FilenameFilter.Filter(file.Filename) {
			if isStatusFilterPresent && !checkListItem.FileStatusFilter.Filter(file.Status) {
				continue
			}
			isCheckListNeeded = true
			if !isGlobalFilenameFilterPresent {
				break
			}
		}
	}
	return isCheckListNeeded
}

func filterProtoChecklistPlan(checkListsPlan []CheckListPlan, prBody string, ignoreLog bool) ([]CheckListPlan, error) {
	creationProtoChecklisIndex := getCheckListPlanIndexIfItWillBePresentInThePR(checkListsPlan, "proto_creation_checklist.md")
	updateProtoCheckListIndex := getCheckListPlanIndexIfItWillBePresentInThePR(checkListsPlan, "proto_update_checklist.md")

	if creationProtoChecklisIndex != -1 && updateProtoCheckListIndex != -1 {
		if !ignoreLog {
			fmt.Println("\nproto_update_checklist.md and proto_creation_checklist.md will be both present, replacing them by a the proto_checklist.md checklist")
			fmt.Println("\nRectification of the plan:")
		}

		removeCheckListItemByCheckListPlanIndex(checkListsPlan, creationProtoChecklisIndex)
		printPlanLog(checkListsPlan[creationProtoChecklisIndex].Filename, true, checkListsPlan[creationProtoChecklisIndex].Present, checkListsPlan[creationProtoChecklisIndex].Action, ignoreLog)

		removeCheckListItemByCheckListPlanIndex(checkListsPlan, updateProtoCheckListIndex)
		printPlanLog(checkListsPlan[updateProtoCheckListIndex].Filename, true, checkListsPlan[updateProtoCheckListIndex].Present, checkListsPlan[updateProtoCheckListIndex].Action, ignoreLog)

		newProtoCheckListPlan, err := addCheckListPlanItemWithoutFilter(prBody, ignoreLog)
		if err != nil {
			return []CheckListPlan{}, err
		}
		checkListsPlan = append(checkListsPlan, newProtoCheckListPlan)
		printPlanLog(protoCheckListFilename, true, newProtoCheckListPlan.Present, newProtoCheckListPlan.Action, ignoreLog)
	}

	return checkListsPlan, nil
}

func addCheckListPlanItemWithoutFilter(prBody string, ignoreLog bool) (CheckListPlan, error) {
	checkListTitle, err := getFirstLine(protoCheckListFilename)
	if err != nil {
		return CheckListPlan{}, err
	}

	isChecklistAlreadyPresent := strings.Contains(prBody, checkListTitle)
	checkListPlan := CheckListPlan{protoCheckListFilename, checkListTitle, ADDED, isChecklistAlreadyPresent}
	if isChecklistAlreadyPresent {
		checkListPlan.Action = IGNORED
	}

	return checkListPlan, nil
}

func removeCheckListItemByCheckListPlanIndex(checkListsPlan []CheckListPlan, index int) {
	if checkListsPlan[index].Action == IGNORED {
		checkListsPlan[index].Action = REMOVED
	} else if checkListsPlan[index].Action == ADDED {
		checkListsPlan[index].Action = IGNORED
	}
}

func getCheckListPlanIndexIfItWillBePresentInThePR(checkListsPlan []CheckListPlan, checkListFilename string) int {
	for index, checkListItemPlan := range checkListsPlan {
		if (checkListItemPlan.Present && checkListItemPlan.Action == IGNORED) || (!checkListItemPlan.Present && checkListItemPlan.Action == ADDED) {
			if checkListItemPlan.Filename == checkListFilename {
				return index
			}
		}
	}
	return -1
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
