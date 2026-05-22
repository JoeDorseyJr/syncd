package download

import (
	"runtime"
	"strings"

	"github.com/joedorseyjr/syncd/internal/runner"
)

// platformFunc is overridable for testing.
var platformFunc = detectPlatform

// Platform returns the brew platform string (e.g., "arm64_sonoma", "sonoma").
func Platform(r runner.CommandRunner) string {
	return platformFunc(r)
}

func detectPlatform(r runner.CommandRunner) string {
	codename := macOSCodename(r)
	if codename == "" {
		return ""
	}
	if runtime.GOARCH == "arm64" {
		return "arm64_" + codename
	}
	return codename
}

func macOSCodename(r runner.CommandRunner) string {
	out, err := r.Run("sw_vers", "-productVersion")
	if err != nil {
		return "sequoia" // fallback to most recent known
	}
	version := strings.TrimSpace(string(out))
	major := strings.Split(version, ".")[0]
	switch major {
	case "15":
		return "sequoia"
	case "14":
		return "sonoma"
	case "13":
		return "ventura"
	case "12":
		return "monterey"
	default:
		return "sequoia" // fallback to most recent known
	}
}
