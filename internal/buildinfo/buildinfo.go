// Package buildinfo describes the build of the binary.
package buildinfo

import (
	"fmt"
	"runtime/debug"
)

// Version describes the build from the values set with -ldflags, "dev" and
// empty strings when they are not set. Binaries installed with `go install`
// get their version and commit from the build information embedded by Go.
func Version(version, commit, date string) string {
	v, c, d := version, commit, date
	if info, ok := debug.ReadBuildInfo(); ok {
		v, c, d = versionFromBuildInfo(info, v, c, d)
	}

	out := "libertea " + v
	if c != "" && d != "" {
		out += fmt.Sprintf(" (%s, %s)", c, d)
	} else if c != "" {
		out += fmt.Sprintf(" (%s)", c)
	}
	return out
}

func versionFromBuildInfo(info *debug.BuildInfo, v, c, d string) (string, string, string) {
	if v == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
		v = info.Main.Version
	}
	for _, setting := range info.Settings {
		switch {
		case setting.Key == "vcs.revision" && c == "":
			c = setting.Value
			if len(c) > 7 {
				c = c[:7]
			}
		case setting.Key == "vcs.time" && d == "":
			d = setting.Value
		}
	}
	return v, c, d
}
