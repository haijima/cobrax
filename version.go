package cobrax

import (
	"fmt"
	"runtime/debug"
)

// VersionFunc returns the version string.
// https://goreleaser.com/cookbooks/using-main.version/
func VersionFunc(version, commit, date string) string {
	var goVersion string
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		if version == "" {
			version = buildInfo.Main.Version
		}
		goVersion = buildInfo.GoVersion

		for _, setting := range buildInfo.Settings {
			if commit == "" && setting.Key == "vcs.revision" {
				commit = setting.Value
			} else if date == "" && setting.Key == "vcs.time" {
				date = setting.Value
			}
		}
	}
	if version == "" {
		version = "unknown"
	}

	versionString := version
	if date != "" {
		versionString += fmt.Sprintf(" (%s)", date)
	}
	if commit != "" {
		versionString += fmt.Sprintf("\n%s", commit)
	}
	versionString += fmt.Sprintf("\nbuilt with %s", goVersion)
	versionString += "\n\nhttps://github.com/haijima/gh-ignore/releases/"
	return versionString
}
