package download

import (
	"fmt"
	"testing"

	"github.com/joedorseyjr/syncd/internal/runner"
)

func TestGetFormulaURLs_EmptyNames(t *testing.T) {
	mock := &runner.MockRunner{}
	result := GetFormulaURLs(mock, nil, "arm64_sonoma")
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
	if len(mock.Calls) != 0 {
		t.Errorf("expected no calls, got %d", len(mock.Calls))
	}
}

func TestGetFormulaURLs_ParsesBottleURL(t *testing.T) {
	jsonData := `{
		"formulae": [{
			"name": "neovim",
			"versions": {"stable": "0.10.0"},
			"bottle": {
				"stable": {
					"files": {
						"arm64_sonoma": {"url": "https://ghcr.io/neovim-0.10.0.tar.gz"},
						"sonoma": {"url": "https://ghcr.io/neovim-0.10.0-intel.tar.gz"}
					}
				}
			}
		}],
		"casks": []
	}`
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{
			{Out: []byte(jsonData)},
		},
	}

	result := GetFormulaURLs(mock, []string{"neovim"}, "arm64_sonoma")
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Name != "neovim" {
		t.Errorf("expected neovim, got %s", result[0].Name)
	}
	if result[0].Version != "0.10.0" {
		t.Errorf("expected 0.10.0, got %s", result[0].Version)
	}
	if result[0].URL != "https://ghcr.io/neovim-0.10.0.tar.gz" {
		t.Errorf("unexpected URL: %s", result[0].URL)
	}
	if result[0].IsCask {
		t.Error("expected IsCask=false")
	}
}

func TestGetFormulaURLs_FallsBackToAll(t *testing.T) {
	jsonData := `{
		"formulae": [{
			"name": "pkg",
			"versions": {"stable": "1.0"},
			"bottle": {
				"stable": {
					"files": {
						"all": {"url": "https://ghcr.io/pkg-all.tar.gz"}
					}
				}
			}
		}],
		"casks": []
	}`
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{
			{Out: []byte(jsonData)},
		},
	}

	result := GetFormulaURLs(mock, []string{"pkg"}, "arm64_sonoma")
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].URL != "https://ghcr.io/pkg-all.tar.gz" {
		t.Errorf("expected all URL, got %s", result[0].URL)
	}
}

func TestGetFormulaURLs_SkipsNoBottle(t *testing.T) {
	jsonData := `{
		"formulae": [{
			"name": "nobottle",
			"versions": {"stable": "1.0"},
			"bottle": {"stable": {"files": {}}}
		}],
		"casks": []
	}`
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{
			{Out: []byte(jsonData)},
		},
	}

	result := GetFormulaURLs(mock, []string{"nobottle"}, "arm64_sonoma")
	if len(result) != 0 {
		t.Errorf("expected 0 results, got %d", len(result))
	}
}

func TestGetFormulaURLs_MultipleFormulae(t *testing.T) {
	jsonData := `{
		"formulae": [
			{"name": "a", "versions": {"stable": "1.0"}, "bottle": {"stable": {"files": {"arm64_sonoma": {"url": "https://a.tar.gz"}}}}},
			{"name": "b", "versions": {"stable": "2.0"}, "bottle": {"stable": {"files": {"arm64_sonoma": {"url": "https://b.tar.gz"}}}}}
		],
		"casks": []
	}`
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{
			{Out: []byte(jsonData)},
		},
	}

	result := GetFormulaURLs(mock, []string{"a", "b"}, "arm64_sonoma")
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
}

func TestGetFormulaURLs_BrewInfoError(t *testing.T) {
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{
			{Err: &runner.RunError{Cmd: "brew info", Err: fmt.Errorf("fail")}},
		},
	}

	result := GetFormulaURLs(mock, []string{"neovim"}, "arm64_sonoma")
	if result != nil {
		t.Errorf("expected nil on error, got %v", result)
	}
}

func TestGetCaskURLs_EmptyNames(t *testing.T) {
	mock := &runner.MockRunner{}
	result := GetCaskURLs(mock, nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestGetCaskURLs_ParsesURL(t *testing.T) {
	jsonData := `{
		"formulae": [],
		"casks": [{
			"token": "firefox",
			"version": "126.0",
			"url": "https://cdn.mozilla.net/firefox-126.0.dmg"
		}]
	}`
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{
			{Out: []byte(jsonData)},
		},
	}

	result := GetCaskURLs(mock, []string{"firefox"})
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Name != "firefox" {
		t.Errorf("expected firefox, got %s", result[0].Name)
	}
	if result[0].Version != "126.0" {
		t.Errorf("expected 126.0, got %s", result[0].Version)
	}
	if result[0].URL != "https://cdn.mozilla.net/firefox-126.0.dmg" {
		t.Errorf("unexpected URL: %s", result[0].URL)
	}
	if !result[0].IsCask {
		t.Error("expected IsCask=true")
	}
}

func TestGetCaskURLs_MultipleCasks(t *testing.T) {
	jsonData := `{
		"formulae": [],
		"casks": [
			{"token": "a", "version": "1.0", "url": "https://a.dmg"},
			{"token": "b", "version": "2.0", "url": "https://b.pkg"}
		]
	}`
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{
			{Out: []byte(jsonData)},
		},
	}

	result := GetCaskURLs(mock, []string{"a", "b"})
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
}

func TestGetCaskURLs_BrewInfoError(t *testing.T) {
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{
			{Err: &runner.RunError{Cmd: "brew info", Err: fmt.Errorf("fail")}},
		},
	}

	result := GetCaskURLs(mock, []string{"firefox"})
	if result != nil {
		t.Errorf("expected nil on error, got %v", result)
	}
}

func TestGetCaskURLs_SkipsEmptyURL(t *testing.T) {
	jsonData := `{
		"formulae": [],
		"casks": [{"token": "nourl", "version": "1.0", "url": ""}]
	}`
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{
			{Out: []byte(jsonData)},
		},
	}

	result := GetCaskURLs(mock, []string{"nourl"})
	if len(result) != 0 {
		t.Errorf("expected 0 results, got %d", len(result))
	}
}
