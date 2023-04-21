package checklist

import "testing"

func TestManageCheckListItemAddingProtoAndSql(t *testing.T) {

	prBody := `Start body 
	Test test 
	Test test`

	checkList := initCheckList()
	filenames := []string{"test.sql", "test2.proto", "test.txt"}

	for _, checkListItem := range checkList {
		prBody = manageCheckListItem(prBody, checkListItem, filenames)
	}

	want := `Start body 
	Test test 
	Test test` + "\n" +
		getFileContent("proto_checklist.md") + "\n" +
		getFileContent("sql_migration_checklist.md")

	if prBody != want {
		t.Errorf("got \n%q \n want \n%q \n", prBody, want)
	}
}
