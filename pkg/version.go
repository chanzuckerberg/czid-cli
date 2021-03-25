package pkg

import (
	"regexp"
	"strings"
)

var Version = "unversioned"

var afterDash = regexp.MustCompile(`-.*$`)

// VersionNumber strips the leading v and anything after a - in the version
func VersionNumber() string {
	version := Version
	if strings.HasPrefix(version, "v") {
		version = version[1:]
	}
	return afterDash.ReplaceAllString(version, "")
}
