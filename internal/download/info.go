package download

import (
	"encoding/json"
	"fmt"

	"github.com/joedorseyjr/syncd/internal/runner"
)

// PackageURL holds the download URL and cache filename info for one package.
type PackageURL struct {
	Name     string
	Version  string
	URL      string
	Filename string
	IsCask   bool
}

// GetFormulaURLs queries brew info --json=v2 for formula bottle URLs.
func GetFormulaURLs(r runner.CommandRunner, names []string, platform string) []PackageURL {
	if len(names) == 0 {
		return nil
	}
	args := append([]string{"info", "--json=v2"}, names...)
	out, err := r.Run("brew", args...)
	if err != nil {
		fmt.Printf("  ⚠ brew info failed, skipping pre-download: %v\n", err)
		return nil
	}

	var info struct {
		Formulae []struct {
			Name     string `json:"name"`
			Versions struct {
				Stable string `json:"stable"`
			} `json:"versions"`
			Bottle struct {
				Stable struct {
					Files map[string]struct {
						URL string `json:"url"`
					} `json:"files"`
				} `json:"stable"`
			} `json:"bottle"`
		} `json:"formulae"`
	}
	if err := json.Unmarshal(out, &info); err != nil {
		fmt.Printf("  ⚠ failed to parse brew info JSON: %v\n", err)
		return nil
	}

	var result []PackageURL
	for _, f := range info.Formulae {
		url := ""
		if file, ok := f.Bottle.Stable.Files[platform]; ok {
			url = file.URL
		} else if file, ok := f.Bottle.Stable.Files["all"]; ok {
			url = file.URL
		}
		if url == "" {
			continue
		}
		result = append(result, PackageURL{
			Name:     f.Name,
			Version:  f.Versions.Stable,
			URL:      url,
			Filename: CacheFilename(url, f.Name, f.Versions.Stable, false),
			IsCask:   false,
		})
	}
	return result
}

// GetCaskURLs queries brew info --json=v2 --cask for cask download URLs.
func GetCaskURLs(r runner.CommandRunner, names []string) []PackageURL {
	if len(names) == 0 {
		return nil
	}
	args := append([]string{"info", "--json=v2", "--cask"}, names...)
	out, err := r.Run("brew", args...)
	if err != nil {
		fmt.Printf("  ⚠ brew info --cask failed, skipping pre-download: %v\n", err)
		return nil
	}

	var info struct {
		Casks []struct {
			Token   string `json:"token"`
			Version string `json:"version"`
			URL     string `json:"url"`
		} `json:"casks"`
	}
	if err := json.Unmarshal(out, &info); err != nil {
		fmt.Printf("  ⚠ failed to parse brew cask info JSON: %v\n", err)
		return nil
	}

	var result []PackageURL
	for _, c := range info.Casks {
		if c.URL == "" {
			continue
		}
		result = append(result, PackageURL{
			Name:     c.Token,
			Version:  c.Version,
			URL:      c.URL,
			Filename: CacheFilename(c.URL, c.Token, c.Version, true),
			IsCask:   true,
		})
	}
	return result
}
