package progress

import (
	"strings"
)

// ExtractError returns relevant error lines from captured brew output.
// Looks for "Error:" lines and surrounding context.
func ExtractError(output []byte) string {
	if len(output) == 0 {
		return ""
	}

	lines := strings.Split(strings.TrimRight(string(output), "\n"), "\n")

	// Scan for "Error:" line (case-insensitive)
	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), "error:") {
			end := i + 3 // error line + up to 2 context lines
			if end > len(lines) {
				end = len(lines)
			}
			result := lines[i:end]
			if len(result) > 5 {
				result = result[:5]
			}
			return strings.Join(result, "\n")
		}
	}

	// Fallback: last 3 lines
	start := len(lines) - 3
	if start < 0 {
		start = 0
	}
	return strings.Join(lines[start:], "\n")
}
