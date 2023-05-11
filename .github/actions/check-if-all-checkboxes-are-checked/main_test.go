package main

import (
	"testing"
)

func TestUnCheckedCheckbox(t *testing.T) {
	got := findUncheckedCheckboxesInPrBody("test  \n- [ ] test")
	want := []string{"- [ ] test"}

	if len(got) != len(want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestCheckedCheckbox(t *testing.T) {
	got := findUncheckedCheckboxesInPrBody("test  \n- [x] test")
	want := []string{}

	if len(got) != len(want) {
		t.Errorf("got %v want %v", got, want)
	}
}
