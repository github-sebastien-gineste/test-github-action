package main

import (
	"strings"
	"testing"
)

func helperManageAddCheckList(t *testing.T, prStartBody string, diffFilenames []string, allCheckListFilesNameNeeded []string) (string, string) {
	updatedPRBody := prStartBody

	// Got
	checkListsPlan, err := getPlanCheckLists(updatedPRBody, diffFilenames, true)
	if err != nil {
		t.Error("Error while getting the plan of the checklists :", err)
	}

	updatedPRBody, err = applyPlanCheckLists(updatedPRBody, checkListsPlan, true)
	if err != nil {
		t.Error("Error while synchronising the checklists : ", err)
	}

	// Want
	newBodyPRContent := ""
	for _, filename := range allCheckListFilesNameNeeded {
		content, err := getFileContent(filename)
		if err != nil {
			t.Error("Checklist item file is empty :", err)
		}
		newBodyPRContent += "\n" + content
	}

	want := prStartBody + newBodyPRContent
	got := updatedPRBody

	return got, want
}

func TestAllCheckListPresence(t *testing.T) {
	for _, checkListItem := range allCheckLists {
		if checkListItem.Filename == "" || checkListItem.RegexDiffFiles == nil {
			t.Error("Checklist item is not complete")
		} else {
			_, err := getFileContent(checkListItem.Filename)
			if err != nil {
				t.Error("Checklist item file is empty for the filename :", checkListItem.Filename)
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

func TestAddingCheckListProductionWithApiDomainsConf(t *testing.T) {
	diffFilenames := []string{"test.txt", "api-domains.conf", "text.md"}
	allCheckListFilesNameNeeded := []string{"production_conf_checklist.md"}

	prStartBody := `Start body 
	Test test 
	Test test`

	got, want := helperManageAddCheckList(t, prStartBody, diffFilenames, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestAddingCheckListProductionWithApiDomainsMigrationsConf(t *testing.T) {
	diffFilenames := []string{"test.txt", "test.md", "api-domains-migrations.conf"}
	allCheckListFilesNameNeeded := []string{"production_conf_checklist.md"}

	prStartBody := `Start body 
	Test test 
	Test test`

	got, want := helperManageAddCheckList(t, prStartBody, diffFilenames, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestAddingCheckListProductionWith_bakeryFolder(t *testing.T) {
	diffFilenames := []string{"test.txt", "test.go", "_bakery/test/text.txt"}
	allCheckListFilesNameNeeded := []string{"production_conf_checklist.md"}

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

	protoCheckListItem := allCheckLists[0]
	if !strings.Contains(protoCheckListItem.RegexDiffFiles.String(), ".proto") {
		t.Error("Proto checklist item is not the first one, there is ", protoCheckListItem.Filename, " in its place")
	}

	contentProto, err := getFileContent(protoCheckListItem.Filename)
	if err != nil {
		t.Error("Checklist item file is empty for the filename :", protoCheckListItem.Filename)
	}

	prStartBody := `Start body 
	Test test 
	Test test` + "\n" + contentProto

	got, want := helperManageAddCheckList(t, prStartBody, diffFilenames, allCheckListFilesNameNeeded)

	// remove the proto checklist in the want
	protoTitle := strings.Split(contentProto, "\n")[0]
	want = removeCheckList(want, protoTitle)

	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestAddingProtoAndRemoveSQLCheckList(t *testing.T) {
	diffFilenames := []string{"test.txt", "test2.proto", "domains/User.scala"}
	allCheckListFilesNameNeeded := []string{"proto_checklist.md"}

	sqlCheckListItem := allCheckLists[4]
	if !strings.Contains(sqlCheckListItem.RegexDiffFiles.String(), ".sql") {
		t.Error("SQL checklist item is not in the 4th index, there is ", sqlCheckListItem.Filename, " in its place")
	}

	contentSQL, err := getFileContent(sqlCheckListItem.Filename)
	if err != nil {
		t.Error("Checklist item file is empty for the filename :", sqlCheckListItem.Filename)
	}

	prStartBody := `Start body 
	Test test 
	Test test` + "\n" + contentSQL

	got, want := helperManageAddCheckList(t, prStartBody, diffFilenames, allCheckListFilesNameNeeded)

	// remove the sql checklist in the want
	sqlTitle := strings.Split(contentSQL, "\n")[0]
	want = removeCheckList(want, sqlTitle)

	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}
