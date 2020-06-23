package version

import (
	"fmt"
	"runtime"
)

var (
	// VerMajor is the major version
	VerMajor = 0
	// VerMinor is the monor version
	VerMinor = 1
	// VerPatch is the patch version
	VerPatch = 2
	// GitHash is the current git commit hash
	GitHash = "Unknown"
)

// Version is the semver of the release
type Version struct {
	major     int
	minor     int
	patch     int
	GitHash   string
	GoVersion string
}

// NewVersion creates a Version object
func NewVersion() *Version {
	return &Version{
		major:     VerMajor,
		minor:     VerMinor,
		patch:     VerPatch,
		GitHash:   GitHash,
		GoVersion: runtime.Version(),
	}
}

// SemVer returns Version in semver format
func (v *Version) SemVer() string {
	return fmt.Sprintf("v%d.%d.%d", v.major, v.minor, v.patch)
}

// String converts Version to a string
func (v *Version) String() string {
	return fmt.Sprintf("%s/%s (%s)", v.SemVer(), v.GitHash, v.GoVersion)
}
