package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/joedorseyjr/syncd/internal/plan"
)

func captureStdout() func() string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	return func() string {
		w.Close()
		os.Stdout = old
		buf := make([]byte, 4096)
		n, _ := r.Read(buf)
		r.Close()
		return string(buf[:n])
	}
}

func TestDisableColor_RemovesANSI(t *testing.T) {
	// Save and restore
	origG, origR, origY, origRst := Green, Red, Yellow, Reset
	defer func() { Green, Red, Yellow, Reset = origG, origR, origY, origRst }()

	DisableColor()

	if Green != "" || Red != "" || Yellow != "" || Reset != "" {
		t.Error("expected all color codes to be empty after DisableColor")
	}
}

func TestPrintSection_GreenForAdditions(t *testing.T) {
	// Ensure colors are enabled
	Green = "\033[32m"
	Reset = "\033[0m"

	p := &plan.Plan{BrewsToInstall: []string{"git"}}
	out := capturePrintPlan(p)

	if !strings.Contains(out, "\033[32m") {
		t.Errorf("expected ANSI green in output, got: %q", out)
	}
	if !strings.Contains(out, "+") {
		t.Errorf("expected + prefix in output, got: %q", out)
	}
}

func TestPrintSection_RedForRemovals(t *testing.T) {
	Red = "\033[31m"
	Reset = "\033[0m"

	p := &plan.Plan{BrewsToRemove: []string{"wget"}}
	out := capturePrintPlan(p)

	if !strings.Contains(out, "\033[31m") {
		t.Errorf("expected ANSI red in output, got: %q", out)
	}
	if !strings.Contains(out, "-") {
		t.Errorf("expected - prefix in output, got: %q", out)
	}
}

func TestPrintSection_YellowForMaintenance(t *testing.T) {
	Yellow = "\033[33m"
	Reset = "\033[0m"

	p := &plan.Plan{Autoremove: true}
	out := capturePrintPlan(p)

	if !strings.Contains(out, "\033[33m") {
		t.Errorf("expected ANSI yellow in output, got: %q", out)
	}
	if !strings.Contains(out, "~") {
		t.Errorf("expected ~ prefix in output, got: %q", out)
	}
}

func TestPrintSection_NoColorWhenDisabled(t *testing.T) {
	origG, origR, origY, origRst := Green, Red, Yellow, Reset
	defer func() { Green, Red, Yellow, Reset = origG, origR, origY, origRst }()

	DisableColor()

	p := &plan.Plan{BrewsToInstall: []string{"git"}, BrewsToRemove: []string{"wget"}, Autoremove: true}
	out := capturePrintPlan(p)

	if strings.Contains(out, "\033[") {
		t.Errorf("expected no ANSI codes when color disabled, got: %q", out)
	}
}

func capturePrintPlan(p *plan.Plan) string {
	// Capture stdout
	old := captureStdout()
	printPlan(p)
	return old()
}
