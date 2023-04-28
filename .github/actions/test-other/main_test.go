package main

import (
	"testing"
)

func TestUnCheckedCheckbox(t *testing.T) {
	got := isContainsUncheckedCheckBox("test  \n- [ ] test", true)
	want := true

	if got != want {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestCheckedCheckbox(t *testing.T) {
	got := isContainsUncheckedCheckBox("test  \n- [x] test", true)
	want := false

	if got != want {
		t.Errorf("got %v want %v", got, want)
	}
}
