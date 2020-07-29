package utils

import (
	"strconv"
	"strings"
)

// SemanticVersion defines the struct for semantic versioning.
type SemanticVersion struct {
	Major int
	Minor int
	Patch int

	Raw string

	Good bool
}

// LessThan returns if it less than input.
func (s SemanticVersion) LessThan(ss SemanticVersion) bool {
	if s.Major > ss.Major {
		return false
	}
	if s.Major < ss.Major {
		return true
	}
	if s.Minor > ss.Minor {
		return false
	}
	if s.Minor < ss.Major {
		return true
	}
	return s.Patch < ss.Patch
}

func toNumber(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return n, true
}

// NewSemanticVersion parse string and convert to SemanticVersion.
// It use a naive implementation, because our spec says it may put an
// illegal semantic version which doesn't contains PATCH.
func NewSemanticVersion(s string) (semVer SemanticVersion) {
	semVer = SemanticVersion{Raw: s}
	if len(s) == 0 {
		return
	}

	// Remove all string after '-'.
	str := strings.SplitN(s, "-", 2)

	// Split into `major.minor.(patch)`.
	parts := strings.SplitN(str[0], ".", 3)

	// Complete missing version number.
	for len(parts) < 3 {
		parts = append(parts, "0")
	}

	dest := []*int{&semVer.Major, &semVer.Minor, &semVer.Patch}
	for i, str := range parts {
		var valid bool
		*dest[i], valid = toNumber(str)
		if !valid {
			return
		}
	}

	if semVer.Major == 0 && semVer.Minor == 0 && semVer.Patch == 0 {
		return
	}

	semVer.Good = true

	return
}
