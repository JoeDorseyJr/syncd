package defaults

import (
	"fmt"
	"strings"

	"github.com/joedorseyjr/syncd/internal/runner"
)

// ReadValue reads the current value of a default. Returns ("", false, nil) if unset.
func ReadValue(r runner.CommandRunner, domain, key string) (string, bool, error) {
	if runner.Verbose {
		fmt.Printf("  > defaults read %s %s\n", domain, key)
	}
	out, err := r.Run("defaults", "read", domain, key)
	if err != nil {
		if _, ok := err.(*runner.RunError); ok {
			return "", false, nil
		}
		return "", false, err
	}
	return strings.TrimSpace(string(out)), true, nil
}

// ReadType reads the stored type of a default. Returns ("", false, nil) if unset.
func ReadType(r runner.CommandRunner, domain, key string) (string, bool, error) {
	if runner.Verbose {
		fmt.Printf("  > defaults read-type %s %s\n", domain, key)
	}
	out, err := r.Run("defaults", "read-type", domain, key)
	if err != nil {
		if _, ok := err.(*runner.RunError); ok {
			return "", false, nil
		}
		return "", false, err
	}
	return strings.TrimSpace(string(out)), true, nil
}
