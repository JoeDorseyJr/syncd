package download

import (
	"path/filepath"
	"strings"

	"github.com/joedorseyjr/syncd/internal/runner"
)

// CacheDir returns brew's download cache directory by running `brew --cache`.
func CacheDir(r runner.CommandRunner) (string, error) {
	out, err := r.Run("brew", "--cache")
	if err != nil {
		return "", err
	}
	dir := strings.TrimSpace(string(out))
	return filepath.Join(dir, "downloads"), nil
}
