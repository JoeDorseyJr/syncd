package progress

import "strings"

// DetectPhase examines a brew output line and returns a phase name.
// Returns "" if the line doesn't indicate a phase change.
func DetectPhase(line string) string {
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "downloading"):
		return "downloading"
	case strings.Contains(lower, "pouring"):
		return "pouring"
	case strings.Contains(lower, "installing"):
		return "installing"
	case strings.Contains(lower, "built from source"), strings.Contains(lower, "built"):
		return "built"
	default:
		return ""
	}
}
