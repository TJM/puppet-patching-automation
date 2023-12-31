package version

import (
	"bytes"
	"fmt"
)

// GitCommit - The git commit that was compiled. This will be filled in by the compiler.
var GitCommit string

// Version - The main version number that is being run at the moment.
const Version = "0.13.1"

// VersionPrerelease - A pre-release marker for the version. If this is "" (empty string)
// then it means that it is a final release. Otherwise, this is a pre-release
// such as "dev" (in development), "beta", "rc1", etc.
const VersionPrerelease = "dev5"

// FormattedVersion - return formatted version string
func FormattedVersion() string {
	var versionString bytes.Buffer
	fmt.Fprintf(&versionString, "%s", Version)
	if VersionPrerelease != "" {
		fmt.Fprintf(&versionString, "-%s", VersionPrerelease)

		if GitCommit != "" {
			fmt.Fprintf(&versionString, " (%s)", GitCommit)
		}
	}

	return versionString.String()
}
