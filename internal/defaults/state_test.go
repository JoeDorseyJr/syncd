package defaults

import (
	"fmt"
	"testing"

	"github.com/joedorseyjr/syncd/internal/runner"
)

func TestReadValue_Success(t *testing.T) {
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{{Out: []byte("48\n"), Err: nil}},
	}
	val, found, err := ReadValue(mock, "com.apple.dock", "tilesize")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected found=true")
	}
	if val != "48" {
		t.Fatalf("expected '48', got '%s'", val)
	}
	if mock.Calls[0].Name != "defaults" || mock.Calls[0].Args[0] != "read" {
		t.Fatalf("unexpected call: %v", mock.Calls[0])
	}
}

func TestReadValue_Unset(t *testing.T) {
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{{Out: nil, Err: &runner.RunError{Cmd: "defaults read", Err: nil}}},
	}
	val, found, err := ReadValue(mock, "com.apple.dock", "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Fatal("expected found=false")
	}
	if val != "" {
		t.Fatalf("expected empty string, got '%s'", val)
	}
}

func TestReadValue_OtherError(t *testing.T) {
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{{Out: nil, Err: fmt.Errorf("network error")}},
	}
	_, _, err := ReadValue(mock, "com.apple.dock", "tilesize")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadType_Success(t *testing.T) {
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{{Out: []byte("Type is integer\n"), Err: nil}},
	}
	val, found, err := ReadType(mock, "com.apple.dock", "tilesize")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected found=true")
	}
	if val != "Type is integer" {
		t.Fatalf("expected 'Type is integer', got '%s'", val)
	}
}

func TestReadType_Unset(t *testing.T) {
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{{Out: nil, Err: &runner.RunError{Cmd: "defaults read-type", Err: nil}}},
	}
	val, found, err := ReadType(mock, "com.apple.dock", "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Fatal("expected found=false")
	}
	if val != "" {
		t.Fatalf("expected empty string, got '%s'", val)
	}
}

func TestReadType_OtherError(t *testing.T) {
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{{Out: nil, Err: fmt.Errorf("network error")}},
	}
	_, _, err := ReadType(mock, "com.apple.dock", "tilesize")
	if err == nil {
		t.Fatal("expected error")
	}
}
