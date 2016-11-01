package pkg

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// Major version number
	Major = 2
	// Minor version number
	Minor = 0
	// Build number
	Build = 0
	// Suffix - set to empty string in release tag commits.
	Suffix = "wip"
)

// Version returns the complete version number.
func Version() string {
	ver := "v" + strconv.Itoa(Major) + "." + strconv.Itoa(Minor) + "." + strconv.Itoa(Build)
	if Suffix != "" {
		ver += "-" + Suffix
	}
	return ver
}

// Compatible returns true if Major matches the major version of the given version string.
// It returns an error if the given string is not a valid version string.
func Compatible(v string) (bool, error) {
	if len(v) < 5 {
		return false, fmt.Errorf("invalid version string format %#v", v)
	}
	v = v[1:]
	elems := strings.Split(v, ".")
	if len(elems) != 3 {
		return false, fmt.Errorf("version not of the form Major.Minor.Build %#v", v)
	}
	mj, err := strconv.Atoi(elems[0])
	if err != nil {
		return false, fmt.Errorf("invalid major version number %#v, must be number", elems[0])
	}
	return mj == Major, nil
}
