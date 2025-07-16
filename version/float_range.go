// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package version

import (
	"fmt"
	"strings"
)

// FloatBehavior Specifies the floating behavior type.
type FloatBehavior int

const (
	// None Lowest version, no float
	None FloatBehavior = iota
	// Prerelease Highest matching pre-release label
	Prerelease FloatBehavior = iota
	// Revision x.y.z.*
	Revision FloatBehavior = iota
	// Patch  x.y.*
	Patch FloatBehavior = iota
	// Minor x.*
	Minor FloatBehavior = iota
	// Major *
	Major FloatBehavior = iota
	// AbsoluteLatest Float major and pre-release, *-*
	AbsoluteLatest FloatBehavior = iota

	// PrereleaseRevision Float revision and pre-release x.y.z.*-*
	PrereleaseRevision FloatBehavior = iota

	// PrereleasePatch Float patch and pre-release x.y.*-*
	PrereleasePatch FloatBehavior = iota

	// PrereleaseMinor Float minor and pre-release x.*-*
	PrereleaseMinor FloatBehavior = iota

	// PrereleaseMajor Float major and prerelease, but only with partial prerelease *-rc.*. *-*
	PrereleaseMajor FloatBehavior = iota
)

// FloatRange The floating subset of a version range.
type FloatRange struct {

	// MinVersion The minimum version of the float range. This is null for cases such as *
	MinVersion *Version

	// FloatBehavior The Defined float behavior
	FloatBehavior FloatBehavior

	// OriginalReleasePrefix The original release label. Invalid labels are allowed here.
	OriginalReleasePrefix string
}

// NewFloatRangeFrom Create a floating range.
func NewFloatRangeFrom(floatBehavior FloatBehavior) *FloatRange {
	pre := ""
	if floatBehavior != None {
		pre = "0"
	}
	return NewFloatRange(floatBehavior, NewVersionFrom(0, 0, 0, pre, ""), "")
}

// NewFloatRangeFromVersion Create a floating range.
func NewFloatRangeFromVersion(floatBehavior FloatBehavior, version *Version) *FloatRange {
	return NewFloatRange(floatBehavior, version, "")

}

// NewFloatRange Create a floating range.
func NewFloatRange(floatBehavior FloatBehavior, minVersion *Version, releasePrefix string) *FloatRange {
	floatRange := &FloatRange{
		FloatBehavior:         floatBehavior,
		MinVersion:            minVersion,
		OriginalReleasePrefix: releasePrefix,
	}
	if strings.TrimSpace(releasePrefix) == "" && minVersion != nil &&
		strings.TrimSpace(minVersion.Semver.Prerelease()) != "" {
		//  use the actual label if one was not given
		floatRange.OriginalReleasePrefix = minVersion.Semver.Prerelease()
	}
	if floatBehavior == AbsoluteLatest && strings.TrimSpace(releasePrefix) == "" {
		floatRange.OriginalReleasePrefix = ""
	}
	return floatRange
}

// HasMinVersion  True if a min range exists.
func (f *FloatRange) HasMinVersion() bool {
	return f.MinVersion != nil
}

// ParseFloatRange Parse a floating version into a FloatRange
func ParseFloatRange(versionString string) (*FloatRange, error) {
	floatRange, ok := TryParseFloatRange(versionString)
	if !ok {
		return nil, fmt.Errorf("%s is not a valid float range string", versionString)
	}
	return floatRange, nil
}

// TryParseFloatRange Parse a floating version into a FloatRange
func TryParseFloatRange(versionString string) (*FloatRange, bool) {
	if strings.TrimSpace(versionString) == "" {
		return nil, false
	}
	var floatRange *FloatRange
	firstStarPosition := strings.Index(versionString, "*")
	lastStarPosition := strings.LastIndex(versionString, "*")
	var releasePrefix string
	if len(versionString) == 1 && firstStarPosition == 0 {
		floatRange = NewFloatRangeFromVersion(Major, NewVersionFrom(0, 0, 0, "", ""))
	} else if strings.EqualFold(versionString, "*-*") {
		v, _ := Parse("0.0.0-0")
		floatRange = NewFloatRangeFromVersion(AbsoluteLatest, v)
	} else if firstStarPosition != lastStarPosition && lastStarPosition != -1 && strings.Index(versionString, "+") == -1 {
		behavior := None
		dashPosition := strings.Index(versionString, "-")
		var actualVersion string
		// Last star is at the end of the full string
		// First star is right before the first dash.
		if dashPosition != -1 && lastStarPosition == (len(versionString)-1) && firstStarPosition == (dashPosition-1) {

			// Get the stable part.
			// Get the part without the *
			stablePart := versionString[:dashPosition-1]
			stablePart += "0"
			var versionParts = calculateVersionParts(stablePart)
			switch versionParts {
			case 1:
				behavior = PrereleaseMajor
				break
			case 2:
				behavior = PrereleaseMinor
				break
			case 3:
				behavior = PrereleasePatch
				break
			case 4:
				behavior = PrereleaseRevision
				break
			default:
				break
			}

			releaseVersion := versionString[dashPosition+1:]
			releasePrefix = releaseVersion[:len(releaseVersion)-1]
			releasePart := releasePrefix
			if len(releasePrefix) == 0 || strings.HasSuffix(releasePrefix, ".") {
				// 1.0.0-* scenario, an empty label is not a valid version.
				releasePart += "0"
			}
			actualVersion = stablePart + "-" + releasePart
		}
		if version, err := Parse(actualVersion); err == nil {
			floatRange = NewFloatRange(behavior, version, releasePrefix)
		}
	} else if lastStarPosition == len(versionString)-1 && strings.Index(versionString, "+") == -1 {
		// A single * can only appear as the last char in the string.
		// * cannot appear in the metadata section after the +

		behavior := None
		actualVersion := versionString[:len(versionString)-1]
		if strings.Index(versionString, "-") == -1 {
			// replace the * with a 0
			actualVersion += "0"
			versionParts := calculateVersionParts(actualVersion)
			if versionParts == 2 {
				behavior = Minor
			} else if versionParts == 3 {
				behavior = Patch
			} else if versionParts == 4 {
				behavior = Revision
			}
		} else {
			behavior = Prerelease

			// check for a prefix
			if strings.Index(versionString, "-") == strings.LastIndex(versionString, "-") {
				releasePrefix = actualVersion[strings.LastIndex(versionString, "-")+1:]

				// For numeric labels 0 is the lowest. For alpha-numeric - is the lowest.
				if len(releasePrefix) == 0 || strings.HasSuffix(actualVersion, ".") {
					// 1.0.0-* scenario, an empty label is not a valid version.
					actualVersion += "0"
				} else if strings.HasSuffix(actualVersion, "-") {

					// Append a dash to allow floating on the next character.
					actualVersion += "-"
				}
			}
		}
		if version, err := Parse(actualVersion); err == nil {
			floatRange = NewFloatRange(behavior, version, releasePrefix)
		}
	} else {
		// normal version parse
		if version, err := Parse(versionString); err == nil {
			// there is no float range for this version
			floatRange = NewFloatRangeFromVersion(None, version)
		}
	}
	return floatRange, floatRange != nil
}

func calculateVersionParts(line string) int {
	count := 1
	if strings.TrimSpace(line) == "" {
		return count
	}
	runes := []rune(line)
	for i := 0; i < len(line); i++ {
		if runes[i] == '.' {
			count++
		}
	}
	return count
}
