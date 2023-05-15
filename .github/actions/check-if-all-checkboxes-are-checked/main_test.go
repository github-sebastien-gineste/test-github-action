package main

import (
	"actions/commons/github"
	"testing"
)

func TestFindUnCheckedCheckboxInPRBody(t *testing.T) {
	got := findUncheckedCheckboxesInText("test  \n- [ ] test")
	want := []string{"- [ ] test"}

	if len(got) != len(want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestFindCheckedCheckboxInPRBody(t *testing.T) {
	got := findUncheckedCheckboxesInText("test  \n- [x] test")
	want := []string{}

	if len(got) != len(want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestFindCheckedCheckboxInComments(t *testing.T) {
	comments := []github.IssueComment{
		{
			Body: github.String("test  \n- [x] comment"),
		},
	}

	got := findUncheckedCheckboxesInComment(comments)
	want := []string{}

	if len(got) != len(want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestFindUnCheckedCheckboxInComments(t *testing.T) {
	comments := []github.IssueComment{
		{
			Body: github.String("test  \n- [ ] comment"),
		},
	}

	got := findUncheckedCheckboxesInComment(comments)
	want := []string{"- [ ] comment"}

	if len(got) != len(want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestFindUnCheckedCheckboxInPRBodyAndComments(t *testing.T) {
	comments := []github.IssueComment{
		{
			Body: github.String("test  \n- [ ] test"),
		},
	}

	got := findUncheckedCheckboxes("test  \n- [ ] comment", comments)
	want := []string{"- [ ] comment", "- [ ] test"}

	if len(got) != len(want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestFindCheckedCheckboxInPRBodyAndComments(t *testing.T) {
	comments := []github.IssueComment{
		{
			Body: github.String("test  \n- [x] test"),
		},
	}

	got := findUncheckedCheckboxes("test  \n- [x] comment", comments)
	want := []string{}

	if len(got) != len(want) {
		t.Errorf("got %v want %v", got, want)
	}
}
