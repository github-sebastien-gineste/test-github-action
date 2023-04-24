package main

import (
	"strings"
	"testing"
)

func helperManageAddCheckList(t *testing.T, prStartBody string, diffFilenames []string, allCheckListFilesNameNeeded []string) (string, string) {
	allCheckList := initAllCheckList()
	updatedPRBody := prStartBody

	// Got
	for _, checkListItem := range allCheckList {
		updatedPRBodyWithThisItem, err := manageCheckListItem(updatedPRBody, checkListItem, diffFilenames)
		updatedPRBody = updatedPRBodyWithThisItem
		if err != nil {
			t.Error("Checklist item file is empty for the filename :", *checkListItem.Filename)
		}
	}

	// Want
	newBodyPRContent := ""
	for _, filename := range allCheckListFilesNameNeeded {
		content, err := getFileContent(filename)
		if err != nil {
			t.Error("Checklist item file is empty for the filename :", filename)
		}
		newBodyPRContent += "\n" + content
	}

	want := prStartBody + newBodyPRContent
	got := updatedPRBody

	return got, want
}

func TestAllCheckListPresence(t *testing.T) {
	allCheckList := initAllCheckList()
	for _, checkListItem := range allCheckList {
		if checkListItem.Filename == nil || checkListItem.Title == nil || checkListItem.RegexDiffFiles == nil {
			t.Error("Checklist item is not complete")
		} else {
			_, err := getFileContent(*checkListItem.Filename)
			if err != nil {
				t.Error("Checklist item file is empty for the filename :", *checkListItem.Filename)
			}
		}
	}
}

func TestAddingCheckListProtoAndSql(t *testing.T) {
	diffFilenames := []string{"test.sql", "test2.proto", "test.txt"}
	allCheckListFilesNameNeeded := []string{"proto_checklist.md", "sql_migration_checklist.md"}

	prStartBody := `Start body 
	Test test 
	Test test`

	got, want := helperManageAddCheckList(t, prStartBody, diffFilenames, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestAddingCheckListConfAndImplementaitonRPC(t *testing.T) {
	diffFilenames := []string{"test.conf", "test2.txt", "domains/UserHandler.scala"}
	allCheckListFilesNameNeeded := []string{"implementation_rpc_checklist.md", "development_conf_checklist.md"}

	prStartBody := `Start body 
	Test test 
	Test test`

	got, want := helperManageAddCheckList(t, prStartBody, diffFilenames, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestRemovingProtoCheckList(t *testing.T) {
	diffFilenames := []string{"test.txt", "test2.txt", "domains/User.scala"}
	allCheckListFilesNameNeeded := []string{}

	protoCheckListItem := initAllCheckList()[0]
	if !strings.Contains(*protoCheckListItem.RegexDiffFiles, ".proto") {
		t.Error("Proto checklist item is not the first one, there is ", *protoCheckListItem.Filename, " in its place")
	}

	contentProto, err := getFileContent(*protoCheckListItem.Filename)
	if err != nil {
		t.Error("Checklist item file is empty for the filename :", protoCheckListItem.Filename)
	}

	prStartBody := `Start body 
	Test test 
	Test test` + "\n" + contentProto

	got, want := helperManageAddCheckList(t, prStartBody, diffFilenames, allCheckListFilesNameNeeded)

	// remove the proto checklist in the want
	want = removeCheckList(want, protoCheckListItem)

	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestAddingProtoAndRemoveSQLCheckList(t *testing.T) {
	diffFilenames := []string{"test.txt", "test2.proto", "domains/User.scala"}
	allCheckListFilesNameNeeded := []string{"proto_checklist.md"}

	sqlCheckListItem := initAllCheckList()[4]
	if !strings.Contains(*sqlCheckListItem.RegexDiffFiles, ".sql") {
		t.Error("SQL checklist item is not in the 4th index, there is ", *sqlCheckListItem.Filename, " in its place")
	}

	contentSQL, err := getFileContent(*sqlCheckListItem.Filename)
	if err != nil {
		t.Error("Checklist item file is empty for the filename :", sqlCheckListItem.Filename)
	}

	prStartBody := `Start body 
	Test test 
	Test test` + "\n" + contentSQL

	got, want := helperManageAddCheckList(t, prStartBody, diffFilenames, allCheckListFilesNameNeeded)

	// remove the sql checklist in the want
	want = removeCheckList(want, sqlCheckListItem)

	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}
