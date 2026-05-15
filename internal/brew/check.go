package brew

import (
	"errors"
	"fmt"
	"os/exec"
)

// CheckAvailable verifies that Homebrew is installed and accessible.
// Returns a user-friendly error with install guidance if not found.
func CheckAvailable() error {
	_, err := exec.LookPath("brew")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("Homebrew is not installed. Install it from https://brew.sh")
		}
		return fmt.Errorf("checking for Homebrew: %w", err)
	}
	return nil
}
