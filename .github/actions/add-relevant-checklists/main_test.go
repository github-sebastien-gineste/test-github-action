package main

import (
	"actions/commons/github"
	"strings"
	"testing"
)

func helperManageAddCheckList(t *testing.T, prStartBody string, diffFilenames []github.CommitFiles, allCheckListFilesNameNeeded []string) (string, string) {
	updatedPRBody := prStartBody

	// Got
	checkListsPlan, err := getCheckListsPlan(updatedPRBody, diffFilenames, true)
	if err != nil {
		t.Error("Error while getting the plan of the checklists :", err)
	}

	updatedPRBody, err = applyCheckListsPlan(updatedPRBody, checkListsPlan, true)
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
		if checkListItem.Filename == "" || checkListItem.InclusionFilter == nil {
			t.Error("Checklist item is not complete")
		} else {
			_, err := getFileContent(checkListItem.Filename)
			if err != nil {
				t.Error("Checklist item file is empty for the filename :", checkListItem.Filename)
			}
		}
	}
}

func TestExclusionFilterForApiDomainsConfWhenTheyAreSQLMigration(t *testing.T) {
	diffFiles := []github.CommitFiles{{Filename: "test.sql"}, {Filename: "folder/api-domains.conf"}, {Filename: "test.txt"}}
	allCheckListFilesNameNeeded := []string{"sql_migration_checklist.md"}

	prStartBody := `Start body`

	got, want := helperManageAddCheckList(t, prStartBody, diffFiles, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestExclusionWhenAFileIsBothIncludedAndExcluded(t *testing.T) {
	diffFiles := []github.CommitFiles{{Filename: "_bakery/folder/test.sql"}, {Filename: "test.txt"}}
	// the first file is both included and excluded in the "production_conf_checklist.md"
	allCheckListFilesNameNeeded := []string{"sql_migration_checklist.md"}

	prStartBody := `Start body`

	got, want := helperManageAddCheckList(t, prStartBody, diffFiles, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestAddingCheckListProtoAndSql(t *testing.T) {
	diffFiles := []github.CommitFiles{{Filename: "test.sql"}, {Filename: "domains/account/account-api/src/main/protobuf/test.proto", Status: "added"}, {Filename: "test.txt"}}
	allCheckListFilesNameNeeded := []string{"addition_proto_checklist.md", "sql_migration_checklist.md"}

	prStartBody := `Start body 
	Test test 
	Test test`

	got, want := helperManageAddCheckList(t, prStartBody, diffFiles, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestAddingCheckListEditionProto(t *testing.T) {
	diffFiles := []github.CommitFiles{{Filename: "test.md"}, {Filename: "domains/account/account-api/src/main/protobuf/test.proto", Status: "modified"}, {Filename: "test.txt"}}
	allCheckListFilesNameNeeded := []string{"edition_proto_checklist.md"}

	prStartBody := `Start body 
	Test test 
	Test test`

	got, want := helperManageAddCheckList(t, prStartBody, diffFiles, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestAddingCheckListEditionProtoAndImplementationRPC(t *testing.T) {
	diffFiles := []github.CommitFiles{
		{Filename: "queries/controllers/handlers/ListPlannedTargetingsAnonymouslyHandlerSpec.scala"},
		{Filename: "domains/test/src/main/protobuf/queries/ListPlannedTargetingsAnonymously.proto", Status: "modified"},
		{Filename: "queries/controllers/handlers/ListPlannedTargetingsAnonymouslyHandler.scala"}}
	allCheckListFilesNameNeeded := []string{"edition_proto_checklist.md", "implementation_rpc_checklist.md"}

	prStartBody := `Start body 
	Test test 
	Test test`

	got, want := helperManageAddCheckList(t, prStartBody, diffFiles, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestAddingProtoChecklistIfTheProtoIsADomainProto(t *testing.T) {
	diffFiles := []github.CommitFiles{{Filename: "domains/account/account-api/src/main/protobuf/test.proto", Status: "added"}}
	allCheckListFilesNameNeeded := []string{"addition_proto_checklist.md"}

	prStartBody := `Start body`

	got, want := helperManageAddCheckList(t, prStartBody, diffFiles, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestAddingProtoChecklistIfTheProtoIsAFrameworkProto(t *testing.T) {
	diffFiles := []github.CommitFiles{{Filename: "framework/api-commons/src/main/protobuf/test.proto", Status: "added"}}
	allCheckListFilesNameNeeded := []string{"addition_proto_checklist.md"}

	prStartBody := `Start body`

	got, want := helperManageAddCheckList(t, prStartBody, diffFiles, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestNotAddingProtoChecklistIfTheProtoIsNotDomainOrFrameworkProto(t *testing.T) {
	diffFiles := []github.CommitFiles{{Filename: "anotherdirectory/test.proto"}}
	allCheckListFilesNameNeeded := []string{}

	prStartBody := `Start body`

	got, want := helperManageAddCheckList(t, prStartBody, diffFiles, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestAddingCheckListProductionConfWithApiDomainsConf(t *testing.T) {
	diffFiles := []github.CommitFiles{{Filename: "test.txt"}, {Filename: "api-domains.conf"}, {Filename: "text.md"}}
	allCheckListFilesNameNeeded := []string{"production_conf_checklist.md"}

	prStartBody := `Start body 
	Test test 
	Test test`

	got, want := helperManageAddCheckList(t, prStartBody, diffFiles, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestAddingCheckListProductionConfWithApiDomainsMigrationsConf(t *testing.T) {
	diffFiles := []github.CommitFiles{{Filename: "test.txt"}, {Filename: "api-domains-migrations.conf"}, {Filename: "text.md"}}
	allCheckListFilesNameNeeded := []string{"production_conf_checklist.md"}

	prStartBody := `Start body 
	Test test 
	Test test`

	got, want := helperManageAddCheckList(t, prStartBody, diffFiles, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestAddingCheckListProductionConfWithImplCommonsReferenceConf(t *testing.T) {
	diffFiles := []github.CommitFiles{{Filename: "framework/impl-commons/src/main/resources/reference.conf"}}
	allCheckListFilesNameNeeded := []string{"production_conf_checklist.md"}

	prStartBody := `Start body
	Test test
	Test test`

	got, want := helperManageAddCheckList(t, prStartBody, diffFiles, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestAddingCheckListProductionConfWithDomainsReferenceConf(t *testing.T) {
	diffFiles := []github.CommitFiles{{Filename: "domains/account/account-impl/src/main/resources/reference.conf"}}
	allCheckListFilesNameNeeded := []string{"production_conf_checklist.md"}

	prStartBody := `Start body
	Test test
	Test test`

	got, want := helperManageAddCheckList(t, prStartBody, diffFiles, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestAddingCheckListDevelopmentConfWithITReferenceConf(t *testing.T) {
	diffFiles := []github.CommitFiles{{Filename: "domains/account/account-impl/src/it/resources/reference.conf"}}
	allCheckListFilesNameNeeded := []string{"development_conf_checklist.md"}

	prStartBody := `Start body
	Test test
	Test test`

	got, want := helperManageAddCheckList(t, prStartBody, diffFiles, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestAddingCheckListProductionConfWith_bakeryFolder(t *testing.T) {
	diffFiles := []github.CommitFiles{{Filename: "test.txt"}, {Filename: "test.go"}, {Filename: "_bakery/test/text.txt"}}
	allCheckListFilesNameNeeded := []string{"production_conf_checklist.md"}

	prStartBody := `Start body 
	Test test 
	Test test`

	got, want := helperManageAddCheckList(t, prStartBody, diffFiles, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestAddingCheckListConfAndImplementationRPC(t *testing.T) {
	diffFiles := []github.CommitFiles{{Filename: "test.conf"}, {Filename: "test2.txt"}, {Filename: "domains/UserHandler.scala"}}
	allCheckListFilesNameNeeded := []string{"implementation_rpc_checklist.md", "development_conf_checklist.md"}

	prStartBody := `Start body 
	Test test 
	Test test`

	got, want := helperManageAddCheckList(t, prStartBody, diffFiles, allCheckListFilesNameNeeded)
	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestRemovingProtoCheckList(t *testing.T) {
	diffFiles := []github.CommitFiles{{Filename: "test.txt"}, {Filename: "test2.txt"}, {Filename: "domains/User.scala"}}
	allCheckListFilesNameNeeded := []string{}

	protoCheckListItem := allCheckLists[0]
	if !strings.Contains(protoCheckListItem.Filename, "proto") {
		t.Error("Proto checklist item is not the first one, there is ", protoCheckListItem.Filename, " in its place")
	}

	contentProto, err := getFileContent(protoCheckListItem.Filename)
	if err != nil {
		t.Error("Checklist item file is empty for the filename :", protoCheckListItem.Filename)
	}

	prStartBody := `Start body 
	Test test 
	Test test` + "\n" + contentProto

	got, want := helperManageAddCheckList(t, prStartBody, diffFiles, allCheckListFilesNameNeeded)

	// remove the proto checklist in the want
	protoTitle := strings.Split(contentProto, "\n")[0]
	want = removeCheckList(want, protoTitle)

	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}

func TestAddingProtoAndRemoveSQLCheckList(t *testing.T) {
	diffFiles := []github.CommitFiles{{Filename: "test.txt"}, {Filename: "domains/account/account-api/src/main/protobuf/test2.proto", Status: "added"}, {Filename: "domains/User.scala"}}
	allCheckListFilesNameNeeded := []string{"addition_proto_checklist.md"}

	sqlCheckListItem := allCheckLists[5]
	if !strings.Contains(sqlCheckListItem.Filename, "sql") {
		t.Error("SQL checklist item is not in the 4th index, there is", sqlCheckListItem.Filename, "in its place")
	}

	contentSQL, err := getFileContent(sqlCheckListItem.Filename)
	if err != nil {
		t.Error("Checklist item file is empty for the filename :", sqlCheckListItem.Filename)
	}

	prStartBody := `Start body 
	Test test 
	Test test` + "\n" + contentSQL

	got, want := helperManageAddCheckList(t, prStartBody, diffFiles, allCheckListFilesNameNeeded)

	// remove the sql checklist in the want
	sqlTitle := strings.Split(contentSQL, "\n")[0]
	want = removeCheckList(want, sqlTitle)

	if got != want {
		t.Errorf("got: \n\n%q \n\n want: \n\n%q \n", got, want)
	}
}
