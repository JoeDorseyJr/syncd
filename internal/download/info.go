package download

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/joedorseyjr/syncd/internal/runner"
)

// PackageURL holds the download URL and cache path info for one package.
type PackageURL struct {
	Name      string
	Version   string
	URL       string
	CachePath string // full path from `brew --cache`
	Filename  string // kept for test compatibility (basename of CachePath)
	IsCask    bool
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

	// Build URL map by name
	urlMap := make(map[string]string)
	versionMap := make(map[string]string)
	for _, f := range info.Formulae {
		url := ""
		if file, ok := f.Bottle.Stable.Files[platform]; ok {
			url = file.URL
		} else if file, ok := f.Bottle.Stable.Files["all"]; ok {
			url = file.URL
		}
		if url != "" {
			urlMap[f.Name] = url
			versionMap[f.Name] = f.Versions.Stable
		}
	}

	if len(urlMap) == 0 {
		return nil
	}

	// Get cache paths from brew --cache
	cachePaths := getBrewCachePaths(r, names, false)

	var result []PackageURL
	for _, name := range names {
		url, ok := urlMap[name]
		if !ok {
			continue
		}
		cachePath := cachePaths[name]
		if cachePath == "" {
			continue
		}
		result = append(result, PackageURL{
			Name:      name,
			Version:   versionMap[name],
			URL:       url,
			CachePath: cachePath,
			IsCask:    false,
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

	// Build URL map by name
	urlMap := make(map[string]string)
	versionMap := make(map[string]string)
	for _, c := range info.Casks {
		if c.URL != "" {
			urlMap[c.Token] = c.URL
			versionMap[c.Token] = c.Version
		}
	}

	if len(urlMap) == 0 {
		return nil
	}

	// Get cache paths from brew --cache --cask
	cachePaths := getBrewCachePaths(r, names, true)

	var result []PackageURL
	for _, name := range names {
		url, ok := urlMap[name]
		if !ok {
			continue
		}
		cachePath := cachePaths[name]
		if cachePath == "" {
			continue
		}
		result = append(result, PackageURL{
			Name:      name,
			Version:   versionMap[name],
			URL:       url,
			CachePath: cachePath,
			IsCask:    true,
		})
	}
	return result
}

// getBrewCachePaths runs `brew --cache <names...>` and returns a map of name→full cache path.
func getBrewCachePaths(r runner.CommandRunner, names []string, isCask bool) map[string]string {
	var args []string
	if isCask {
		args = append([]string{"--cache", "--cask"}, names...)
	} else {
		args = append([]string{"--cache"}, names...)
	}
	out, err := r.Run("brew", args...)
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	result := make(map[string]string, len(names))
	for i, name := range names {
		if i < len(lines) && lines[i] != "" {
			result[name] = lines[i]
		}
	}
	return result
}
