package download

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/joedorseyjr/syncd/internal/runner"
)

// CacheFilename generates brew's cache filename for a given URL, name, and version.
// Formula: SHA256(url)--name--version.bottle.tar.gz
// Cask: SHA256(url)--name--version.ext (preserves extension from URL)
func CacheFilename(url, name, version string, isCask bool) string {
	hash := sha256.Sum256([]byte(url))
	prefix := fmt.Sprintf("%x", hash)
	if isCask {
		ext := filepath.Ext(url)
		return fmt.Sprintf("%s--%s--%s%s", prefix, name, version, ext)
	}
	return fmt.Sprintf("%s--%s--%s.bottle.tar.gz", prefix, name, version)
}

// CacheDir returns brew's download cache directory by running `brew --cache`.
func CacheDir(r runner.CommandRunner) (string, error) {
	out, err := r.Run("brew", "--cache")
	if err != nil {
		return "", err
	}
	dir := strings.TrimSpace(string(out))
	return filepath.Join(dir, "downloads"), nil
}
