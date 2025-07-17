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

// NewFloatRangeFromVersion Create a floating range.
func NewFloatRangeFromVersion(floatBehavior FloatBehavior, version *Version) *FloatRange {
	return NewFloatRange(floatBehavior, version, nil)
}

// NewFloatRange Create a floating range.
func NewFloatRange(floatBehavior FloatBehavior, minVersion *Version, releasePrefix *string) *FloatRange {
	floatRange := &FloatRange{
		FloatBehavior: floatBehavior,
		MinVersion:    minVersion,
	}
	if releasePrefix != nil {
		floatRange.OriginalReleasePrefix = *releasePrefix
	}
	if releasePrefix == nil && minVersion != nil &&
		strings.TrimSpace(minVersion.Semver.Prerelease()) != "" {
		//  use the actual label if one was not given
		floatRange.OriginalReleasePrefix = minVersion.Semver.Prerelease()
	}
	if floatBehavior == AbsoluteLatest && releasePrefix == nil {
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
	var releasePrefix *string
	if len(versionString) == 1 && firstStarPosition == 0 {
		floatRange = NewFloatRangeFromVersion(Major, NewVersionFrom(0, 0, 0, "", ""))
	} else if strings.EqualFold(versionString, "*-*") {
		v, _ := Parse("0.0.0-0")
		floatRange = NewFloatRangeFromVersion(AbsoluteLatest, v)
	} else if firstStarPosition != lastStarPosition && lastStarPosition != -1 && !strings.Contains(versionString, "+") {
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
			case 2:
				behavior = PrereleaseMinor
			case 3:
				behavior = PrereleasePatch
			case 4:
				behavior = PrereleaseRevision
			default:
				break
			}

			releaseVersion := versionString[dashPosition+1:]
			releasePrefix = stringPtr(releaseVersion[:len(releaseVersion)-1])
			releasePart := stringValue(releasePrefix)
			if len(stringValue(releasePrefix)) == 0 || strings.HasSuffix(stringValue(releasePrefix), ".") {
				// 1.0.0-* scenario, an empty label is not a valid version.
				releasePart += "0"
			}
			actualVersion = stablePart + "-" + releasePart
		}
		if version, err := Parse(actualVersion); err == nil {
			floatRange = NewFloatRange(behavior, version, releasePrefix)
		}
	} else if lastStarPosition == len(versionString)-1 && !strings.Contains(versionString, "+") {
		// A single * can only appear as the last char in the string.
		// * cannot appear in the metadata section after the +

		behavior := None
		actualVersion := versionString[:len(versionString)-1]
		if !strings.Contains(versionString, "-") {
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
				releasePrefix = stringPtr(actualVersion[strings.LastIndex(versionString, "-")+1:])

				// For numeric labels 0 is the lowest. For alpha-numeric - is the lowest.
				if len(stringValue(releasePrefix)) == 0 || strings.HasSuffix(actualVersion, ".") {
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

// String Create a floating version string in the format: 1.0.0-alpha-*
func (f *FloatRange) String() string {
	builder := &strings.Builder{}
	f.string(builder)
	return builder.String()
}

// string Create a floating version string in the format: 1.0.0-alpha-*
func (f *FloatRange) string(builder *strings.Builder) {
	switch f.FloatBehavior {
	case None:
		appendNormalized(builder, f.MinVersion)
	case Prerelease:
		appendVersion(builder, f.MinVersion)
		builder.WriteString(fmt.Sprintf("-%s*", f.OriginalReleasePrefix))
	case Revision:
		builder.WriteString(
			fmt.Sprintf(
				"%v.%v.%v.*",
				f.MinVersion.Semver.Major(),
				f.MinVersion.Semver.Minor(),
				f.MinVersion.Semver.Patch(),
			),
		)
	case Patch:
		builder.WriteString(fmt.Sprintf("%v.%v.*", f.MinVersion.Semver.Major(), f.MinVersion.Semver.Minor()))
	case Minor:
		builder.WriteString(fmt.Sprintf("%v.*", f.MinVersion.Semver.Major()))
	case Major:
		builder.WriteString("*")
	case PrereleaseRevision:
		builder.WriteString(
			fmt.Sprintf(
				"%v.%v.%v.*-%s*",
				f.MinVersion.Semver.Major(),
				f.MinVersion.Semver.Minor(),
				f.MinVersion.Semver.Patch(),
				f.OriginalReleasePrefix,
			),
		)
	case PrereleasePatch:
		builder.WriteString(
			fmt.Sprintf(
				"%v.%v.*-%s*",
				f.MinVersion.Semver.Major(),
				f.MinVersion.Semver.Minor(),
				f.OriginalReleasePrefix,
			),
		)
	case PrereleaseMinor:
		builder.WriteString(fmt.Sprintf("%v.*-%s*", f.MinVersion.Semver.Major(), f.OriginalReleasePrefix))
	case PrereleaseMajor:
		builder.WriteString(fmt.Sprintf("*-%s*", f.OriginalReleasePrefix))
	case AbsoluteLatest:
		builder.WriteString("*-*")
	default:
		break
	}
}

func stringPtr(s string) *string {
	return &s
}
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
