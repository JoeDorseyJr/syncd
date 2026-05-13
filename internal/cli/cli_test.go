package cli

import (
	"strings"
	"testing"
)

func TestConfirm_Yes(t *testing.T) {
	r := strings.NewReader("y\n")
	if !confirm(r) {
		t.Error("expected confirm to return true for 'y'")
	}
}

func TestConfirm_YesFull(t *testing.T) {
	r := strings.NewReader("yes\n")
	if !confirm(r) {
		t.Error("expected confirm to return true for 'yes'")
	}
}

func TestConfirm_No(t *testing.T) {
	r := strings.NewReader("n\n")
	if confirm(r) {
		t.Error("expected confirm to return false for 'n'")
	}
}

func TestConfirm_Empty(t *testing.T) {
	r := strings.NewReader("\n")
	if confirm(r) {
		t.Error("expected confirm to return false for empty input (default N)")
	}
}

func TestConfirm_EOF(t *testing.T) {
	r := strings.NewReader("")
	if confirm(r) {
		t.Error("expected confirm to return false on EOF")
	}
}
