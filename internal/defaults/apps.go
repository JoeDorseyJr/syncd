package defaults

// KnownApps maps preference domains to the app that must be restarted.
// Empty string means no restart needed.
var KnownApps = map[string]string{
	"com.apple.dock":                                     "Dock",
	"com.apple.finder":                                   "Finder",
	"com.apple.systemuiserver":                           "SystemUIServer",
	"com.apple.menuextra.clock":                          "SystemUIServer",
	"com.apple.Safari":                                   "Safari",
	"com.apple.Terminal":                                  "Terminal",
	"com.apple.screencapture":                            "SystemUIServer",
	"com.apple.driver.AppleBluetoothMultitouch.trackpad": "",
	"com.apple.AppleMultitouchTrackpad":                  "",
	"com.apple.HIToolbox":                                "",
	"com.apple.desktopservices":                          "Finder",
	"com.apple.ActivityMonitor":                          "Activity Monitor",
	"com.apple.TextEdit":                                 "TextEdit",
	"NSGlobalDomain":                                     "",
}
