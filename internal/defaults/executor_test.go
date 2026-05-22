package defaults

import (
	"fmt"
	"testing"

	"github.com/joedorseyjr/syncd/internal/runner"
)

func TestWriteDrifted_IntWrite(t *testing.T) {
	m := &runner.MockRunner{}
	m.Outputs = []runner.MockOutput{{}}
	drifted := []DriftEntry{{Domain: "com.apple.dock", Key: "tilesize", Type: "int", Desired: "48"}}

	WriteDrifted(m, drifted)

	if len(m.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(m.Calls))
	}
	c := m.Calls[0]
	expect := []string{"write", "com.apple.dock", "tilesize", "-int", "48"}
	if c.Name != "defaults" {
		t.Errorf("expected defaults, got %s", c.Name)
	}
	for i, a := range expect {
		if c.Args[i] != a {
			t.Errorf("arg[%d]: expected %q, got %q", i, a, c.Args[i])
		}
	}
}

func TestWriteDrifted_FloatWrite(t *testing.T) {
	m := &runner.MockRunner{}
	m.Outputs = []runner.MockOutput{{}}
	drifted := []DriftEntry{{Domain: "NSGlobalDomain", Key: "scaling", Type: "float", Desired: "0.5"}}

	WriteDrifted(m, drifted)

	c := m.Calls[0]
	if c.Args[3] != "-float" || c.Args[4] != "0.5" {
		t.Errorf("expected -float 0.5, got %s %s", c.Args[3], c.Args[4])
	}
}

func TestWriteDrifted_BoolTrueWrite(t *testing.T) {
	m := &runner.MockRunner{}
	m.Outputs = []runner.MockOutput{{}}
	drifted := []DriftEntry{{Domain: "com.apple.dock", Key: "autohide", Type: "bool", Desired: "true"}}

	WriteDrifted(m, drifted)

	c := m.Calls[0]
	if c.Args[3] != "-bool" || c.Args[4] != "TRUE" {
		t.Errorf("expected -bool TRUE, got %s %s", c.Args[3], c.Args[4])
	}
}

func TestWriteDrifted_BoolFalseWrite(t *testing.T) {
	m := &runner.MockRunner{}
	m.Outputs = []runner.MockOutput{{}}
	drifted := []DriftEntry{{Domain: "com.apple.dock", Key: "autohide", Type: "bool", Desired: "false"}}

	WriteDrifted(m, drifted)

	c := m.Calls[0]
	if c.Args[4] != "FALSE" {
		t.Errorf("expected FALSE, got %s", c.Args[4])
	}
}

func TestWriteDrifted_StringWrite(t *testing.T) {
	m := &runner.MockRunner{}
	m.Outputs = []runner.MockOutput{{}}
	drifted := []DriftEntry{{Domain: "NSGlobalDomain", Key: "AppleLanguages", Type: "string", Desired: "en"}}

	WriteDrifted(m, drifted)

	c := m.Calls[0]
	if c.Args[3] != "-string" || c.Args[4] != "en" {
		t.Errorf("expected -string en, got %s %s", c.Args[3], c.Args[4])
	}
}

func TestWriteDrifted_KillDeduplication(t *testing.T) {
	m := &runner.MockRunner{}
	m.Outputs = []runner.MockOutput{{}, {}, {}} // two writes + one killall
	drifted := []DriftEntry{
		{Domain: "com.apple.dock", Key: "tilesize", Type: "int", Desired: "48", Kill: []string{"Dock"}},
		{Domain: "com.apple.dock", Key: "autohide", Type: "bool", Desired: "true", Kill: []string{"Dock"}},
	}

	_, killResults := WriteDrifted(m, drifted)

	if len(killResults) != 1 {
		t.Fatalf("expected 1 kill result (deduplicated), got %d", len(killResults))
	}
	// Find the killall call
	killCalls := 0
	for _, c := range m.Calls {
		if c.Name == "killall" {
			killCalls++
		}
	}
	if killCalls != 1 {
		t.Errorf("expected 1 killall call, got %d", killCalls)
	}
}

func TestWriteDrifted_KillallFailureNonFatal(t *testing.T) {
	m := &runner.MockRunner{}
	m.Outputs = []runner.MockOutput{
		{},                                    // write succeeds
		{Err: fmt.Errorf("No matching processes")}, // killall fails
	}
	drifted := []DriftEntry{
		{Domain: "com.apple.dock", Key: "tilesize", Type: "int", Desired: "48", Kill: []string{"Dock"}},
	}

	writeResults, killResults := WriteDrifted(m, drifted)

	if writeResults[0].Err != nil {
		t.Errorf("write should succeed, got: %v", writeResults[0].Err)
	}
	if killResults[0].Err == nil {
		t.Error("killall should have error")
	}
}

func TestWriteDrifted_WriteFailureContinues(t *testing.T) {
	m := &runner.MockRunner{}
	m.Outputs = []runner.MockOutput{
		{Err: fmt.Errorf("write failed")}, // first write fails
		{},                                 // second write succeeds
	}
	drifted := []DriftEntry{
		{Domain: "bad.domain", Key: "key1", Type: "int", Desired: "1"},
		{Domain: "good.domain", Key: "key2", Type: "int", Desired: "2"},
	}

	writeResults, _ := WriteDrifted(m, drifted)

	if len(writeResults) != 2 {
		t.Fatalf("expected 2 write results, got %d", len(writeResults))
	}
	if writeResults[0].Err == nil {
		t.Error("first write should have failed")
	}
	if writeResults[1].Err != nil {
		t.Errorf("second write should succeed, got: %v", writeResults[1].Err)
	}
}

func TestWriteDrifted_NoKillWhenNoDrift(t *testing.T) {
	m := &runner.MockRunner{}
	var drifted []DriftEntry

	_, killResults := WriteDrifted(m, drifted)

	if len(killResults) != 0 {
		t.Errorf("expected no kill results, got %d", len(killResults))
	}
	if len(m.Calls) != 0 {
		t.Errorf("expected no calls, got %d", len(m.Calls))
	}
}

func TestWriteDrifted_NoKillOnWriteFailure(t *testing.T) {
	m := &runner.MockRunner{}
	m.Outputs = []runner.MockOutput{
		{Err: fmt.Errorf("write failed")},
	}
	drifted := []DriftEntry{
		{Domain: "com.apple.dock", Key: "tilesize", Type: "int", Desired: "48", Kill: []string{"Dock"}},
	}

	_, killResults := WriteDrifted(m, drifted)

	if len(killResults) != 0 {
		t.Errorf("expected no kill when write failed, got %d", len(killResults))
	}
}
