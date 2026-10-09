package main

import (
	"runtime/debug"
	"strings"
	"testing"
)

func TestVersionFromBuildInfo(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{Version: "v1.2.3"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "fca2ef0a8e55aa87bf8ea00c85a105aff7e3ead7"},
			{Key: "vcs.time", Value: "2026-10-08T15:26:06Z"},
		},
	}

	v, c, d := versionFromBuildInfo(info, "dev", "", "")
	if v != "v1.2.3" || c != "fca2ef0" || d != "2026-10-08T15:26:06Z" {
		t.Errorf("go install builds should use the embedded build information, got %q %q %q", v, c, d)
	}

	v, c, d = versionFromBuildInfo(info, "1.0.0", "abcdef1", "2026-01-01")
	if v != "1.0.0" || c != "abcdef1" || d != "2026-01-01" {
		t.Errorf("ldflags values should win, got %q %q %q", v, c, d)
	}

	v, _, _ = versionFromBuildInfo(&debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, "dev", "", "")
	if v != "dev" {
		t.Errorf("local builds should stay dev, got %q", v)
	}
}

func TestVersion(t *testing.T) {
	if got := Version(); !strings.HasPrefix(got, "libertea ") {
		t.Errorf("unexpected version string %q", got)
	}
}
