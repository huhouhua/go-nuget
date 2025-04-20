// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"strings"
)

var (
	All = NewVersionRange(nil, nil, true, true)
)

type NuGetVersion struct {
	*semver.Version
}

func (v *NuGetVersion) IsSemVer2() bool {
	return v.Prerelease() != "" || v.Metadata() != ""
}

func (v *NuGetVersion) IsPrerelease() bool {
	return v.Prerelease() != ""
}

// VersionRange represents a range of versions that satisfy a given constraint.
// This is similar to NuGet's VersionRange class.
type VersionRange struct {
	// MinVersion is the minimum version in the range (inclusive)
	MinVersion *NuGetVersion
	// MaxVersion is the maximum version in the range (inclusive)
	MaxVersion *NuGetVersion
	// IncludeMin indicates whether the minimum version is included in the range
	IncludeMin bool
	// IncludeMax indicates whether the maximum version is included in the range
	IncludeMax bool
	// Float indicates the floating behavior of the version range
	Float FloatBehavior
}

// FloatBehavior represents how version floating should behave
type FloatBehavior int

const (
	// None means no floating behavior
	None FloatBehavior = iota
	// Prerelease allows floating to prerelease versions
	Prerelease
	// Patch allows floating to patch versions
	Patch
	// Minor allows floating to minor versions
	Minor
	// Major allows floating to major versions
	Major
)

// ParseVersionRange parses a version range string into a VersionRange
// Examples:
//   - "1.0.0" -> exact version
//   - "[1.0.0, 2.0.0]" -> range from 1.0.0 to 2.0.0 (inclusive)
//   - "(1.0.0, 2.0.0)" -> range from 1.0.0 to 2.0.0 (exclusive)
//   - "[1.0.0,)" -> range from 1.0.0 to infinity
//   - "(,2.0.0]" -> range from negative infinity to 2.0.0
func ParseVersionRange(rangeStr string) (*VersionRange, error) {
	rangeStr = strings.TrimSpace(rangeStr)
	if rangeStr == "" {
		return nil, fmt.Errorf("empty version range")
	}

	// Handle floating versions
	if strings.Contains(rangeStr, "*") {
		return parseFloatingVersion(rangeStr)
	}

	// Handle tilde and caret ranges
	if strings.HasPrefix(rangeStr, "~") || strings.HasPrefix(rangeStr, "^") {
		return parseFloatingRange(rangeStr)
	}

	// Handle exact version
	if !strings.HasPrefix(rangeStr, "[") && !strings.HasPrefix(rangeStr, "(") {
		v, err := semver.NewVersion(rangeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid version: %v", err)
		}
		return NewVersionRange(v, v, true, true), nil
	}

	// Parse range
	parts := strings.Split(rangeStr[1:len(rangeStr)-1], ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range format: %s", rangeStr)
	}

	minStr := strings.TrimSpace(parts[0])
	maxStr := strings.TrimSpace(parts[1])

	var min, max *semver.Version
	var err error

	if minStr != "" {
		min, err = semver.NewVersion(minStr)
		if err != nil {
			return nil, fmt.Errorf("invalid min version: %v", err)
		}
	}

	if maxStr != "" {
		max, err = semver.NewVersion(maxStr)
		if err != nil {
			return nil, fmt.Errorf("invalid max version: %v", err)
		}
	}

	return NewVersionRange(min, max, strings.HasPrefix(rangeStr, "["), strings.HasSuffix(rangeStr, "]")), nil
}

// parseFloatingVersion parses a floating version string
func parseFloatingVersion(rangeStr string) (*VersionRange, error) {
	switch {
	case rangeStr == "*":
		return &VersionRange{Float: Major}, nil
	case strings.HasPrefix(rangeStr, "*-"):
		return &VersionRange{Float: Prerelease}, nil
	case strings.Contains(rangeStr, "-*"):
		baseVersion := strings.TrimSuffix(rangeStr, "-*")
		v, err := semver.NewVersion(baseVersion)
		if err != nil {
			return nil, fmt.Errorf("invalid version in prerelease range: %v", err)
		}
		vr := NewVersionRange(v, nil, true, false)
		vr.Float = Prerelease
		return vr, nil
	case strings.HasPrefix(rangeStr, "~") || strings.HasPrefix(rangeStr, "^"):
		return parseFloatingRange(rangeStr)
	default:
		return nil, fmt.Errorf("invalid floating version format: %s", rangeStr)
	}
}

// parseFloatingRange parses a floating range string
func parseFloatingRange(rangeStr string) (*VersionRange, error) {
	var prefix string
	var floatType FloatBehavior

	switch {
	case strings.HasPrefix(rangeStr, "~"):
		prefix = "~"
		floatType = Patch
	case strings.HasPrefix(rangeStr, "^"):
		prefix = "^"
		floatType = Minor
	default:
		return nil, fmt.Errorf("invalid floating range format: %s", rangeStr)
	}

	v, err := semver.NewVersion(strings.TrimPrefix(rangeStr, prefix))
	if err != nil {
		return nil, fmt.Errorf("invalid version in %s range: %v", prefix, err)
	}

	var maxVersion *semver.Version
	if floatType == Patch {
		maxVersion = semver.New(v.Major(), v.Minor()+1, 0, "", "")
	} else { // Minor
		if v.Major() == 0 {
			maxVersion = semver.New(v.Major(), v.Minor()+1, 0, "", "")
		} else {
			maxVersion = semver.New(v.Major()+1, 0, 0, "", "")
		}
	}

	vr := NewVersionRange(v, maxVersion, true, false)
	vr.Float = floatType
	return vr, nil
}

// NewVersionRange creates a new VersionRange with the specified parameters
func NewVersionRange(min, max *semver.Version, includeMin, includeMax bool) *VersionRange {
	v := &VersionRange{
		IncludeMin: includeMin,
		IncludeMax: includeMax,
		Float:      None,
	}
	if min != nil {
		v.MinVersion = &NuGetVersion{min}
	}
	if max != nil {
		v.MaxVersion = &NuGetVersion{max}
	}

	return v
}

// Satisfies checks if a version satisfies this version range
func (vr *VersionRange) Satisfies(v *semver.Version) bool {
	if vr.Float != None {
		return vr.satisfiesFloat(v)
	}

	if vr.MinVersion != nil {
		if vr.IncludeMin {
			if v.LessThan(vr.MinVersion.Version) {
				return false
			}
		} else {
			if !v.GreaterThan(vr.MinVersion.Version) {
				return false
			}
		}
	}

	if vr.MaxVersion != nil {
		if vr.IncludeMax {
			if v.GreaterThan(vr.MaxVersion.Version) {
				return false
			}
		} else {
			if !v.LessThan(vr.MaxVersion.Version) {
				return false
			}
		}
	}

	return true
}

// satisfiesFloat checks if a version satisfies the floating behavior
func (vr *VersionRange) satisfiesFloat(v *semver.Version) bool {
	switch vr.Float {
	case Major:
		return true
	case Prerelease:
		return v.Prerelease() != ""
	case Patch:
		return v.Major() == vr.MinVersion.Major() &&
			v.Minor() == vr.MinVersion.Minor() &&
			v.Patch() >= vr.MinVersion.Patch()
	case Minor:
		if vr.MinVersion.Major() == 0 {
			return v.Major() == 0 &&
				v.Minor() == vr.MinVersion.Minor() &&
				v.Patch() >= vr.MinVersion.Patch()
		}
		return v.Major() == vr.MinVersion.Major() &&
			v.Minor() >= vr.MinVersion.Minor()
	default:
		return false
	}
}

// DoesRangeSatisfy checks if this version range satisfies another version range
func (vr *VersionRange) DoesRangeSatisfy(other *VersionRange) bool {
	// If this range has both lower and upper bounds
	if vr.MinVersion != nil && vr.MaxVersion != nil {
		// Create a new range with the bounds of this range
		rangeWithBounds := &VersionRange{
			MinVersion: vr.MinVersion,
			MaxVersion: vr.MaxVersion,
			IncludeMin: true,
			IncludeMax: true,
		}

		// Check if either the min or max version of the other range satisfies this range
		return rangeWithBounds.Satisfies(other.MinVersion.Version) || rangeWithBounds.Satisfies(other.MaxVersion.Version)
	} else {
		// If this range doesn't have both bounds, check if either bound of the other range satisfies this range
		return vr.Satisfies(other.MinVersion.Version) || vr.Satisfies(other.MaxVersion.Version)
	}
}

// String returns the string representation of the version range
func (vr *VersionRange) String() string {
	if vr.Float != None {
		return vr.stringFloat()
	}

	if vr.MinVersion != nil && vr.MaxVersion != nil && vr.MinVersion.Version.Equal(vr.MaxVersion.Version) {
		return vr.MinVersion.String()
	}

	var sb strings.Builder
	if vr.MinVersion != nil {
		if vr.IncludeMin {
			sb.WriteString("[")
		} else {
			sb.WriteString("(")
		}
		sb.WriteString(vr.MinVersion.String())
	} else {
		sb.WriteString("(")
	}

	sb.WriteString(",")

	if vr.MaxVersion != nil {
		sb.WriteString(vr.MaxVersion.String())
		if vr.IncludeMax {
			sb.WriteString("]")
		} else {
			sb.WriteString(")")
		}
	} else {
		sb.WriteString(")")
	}

	return sb.String()
}

// stringFloat returns the string representation of a floating version range
func (vr *VersionRange) stringFloat() string {
	switch vr.Float {
	case Major:
		return "*"
	case Prerelease:
		return "*-"
	case Patch:
		return "~" + vr.MinVersion.String()
	case Minor:
		return "^" + vr.MinVersion.String()
	default:
		return ""
	}
}

// FindBestMatch returns the version that best matches the range from a list of versions
func (vr *VersionRange) FindBestMatch(versions []*semver.Version) *semver.Version {
	var bestMatch *semver.Version
	for _, version := range versions {
		if vr.IsBetter(bestMatch, version) {
			bestMatch = version
		}
	}
	return bestMatch
}

// IsBetter determines if a given version is better suited to the range than a current version
func (vr *VersionRange) IsBetter(current, considering *semver.Version) bool {
	if current == considering {
		return false
	}

	// null checks
	if considering == nil {
		return false
	}

	// If the range contains only stable versions disallow prerelease versions
	if !vr.hasPrereleaseBounds() && considering.Prerelease() != "" &&
		vr.Float != Prerelease && vr.Float != Major {
		return false
	}

	if !vr.Satisfies(considering) {
		// keep null over a value outside of the range
		return false
	}

	if current == nil {
		return true
	}

	if vr.Float != None {
		// check if either version is in the floating range
		curInRange := vr.satisfiesFloat(current)
		conInRange := vr.satisfiesFloat(considering)

		if curInRange && !conInRange {
			// take the version in the range
			return false
		} else if conInRange && !curInRange {
			// take the version in the range
			return true
		} else if curInRange && conInRange {
			// prefer the highest one if both are in the range
			return current.LessThan(considering)
		} else {
			// neither are in range
			curToLower := current.LessThan(vr.MinVersion.Version)
			conToLower := considering.LessThan(vr.MinVersion.Version)

			if curToLower && !conToLower {
				// favor the version above the range
				return true
			} else if !curToLower && conToLower {
				// favor the version above the range
				return false
			} else if !curToLower && !conToLower {
				// favor the lower version if we are above the range
				return current.GreaterThan(considering)
			} else if curToLower && conToLower {
				// favor the higher version if we are below the range
				return current.LessThan(considering)
			}
		}
	}

	// Favor lower versions
	return current.GreaterThan(considering)
}

// hasPrereleaseBounds returns true if either bound is a prerelease version
func (vr *VersionRange) hasPrereleaseBounds() bool {
	return (vr.MinVersion != nil && vr.MinVersion.IsPrerelease()) ||
		(vr.MaxVersion != nil && vr.MaxVersion.IsPrerelease())
}

// ToNonSnapshotRange removes the floating snapshot part of the minimum version if it exists
func (vr *VersionRange) ToNonSnapshotRange() *VersionRange {
	if vr.MinVersion == nil && vr.MaxVersion == nil {
		return vr
	}

	minVersion := vr.MinVersion
	maxVersion := vr.MaxVersion

	// Handle floating versions
	if vr.Float != None {
		if minVersion != nil {
			var major, minor uint64
			if vr.Float == Minor {
				major = minVersion.Major() + 1
				minor = 0
			} else if vr.Float == Patch {
				major = minVersion.Major()
				minor = minVersion.Minor() + 1
			}
			maxVersion = &NuGetVersion{semver.New(major, minor, 0, "", "")}
		}
		return NewVersionRange(minVersion.Version, maxVersion.Version, true, false)
	}

	// Handle prerelease versions
	if minVersion != nil {
		minVersion = processPrereleaseVersion(minVersion)
	}
	if maxVersion != nil {
		maxVersion = processPrereleaseVersion(maxVersion)
	}

	// Create new version range with original include flags
	return NewVersionRange(minVersion.Version, maxVersion.Version, vr.IncludeMin, vr.IncludeMax)
}

// processPrereleaseVersion processes a version with prerelease information
func processPrereleaseVersion(v *NuGetVersion) *NuGetVersion {
	if v == nil || v.Prerelease() == "" {
		return v
	}

	prerelease := strings.TrimRight(v.Prerelease(), "-")
	if prerelease == "0" {
		return &NuGetVersion{semver.New(v.Major(), v.Minor(), v.Patch(), "", "")}
	}
	return &NuGetVersion{semver.New(v.Major(), v.Minor(), v.Patch(), prerelease, "")}
}

// boolToInt converts a boolean to an integer (0 or 1)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// PrettyPrint returns a human-readable string representation of the version range
func (vr *VersionRange) PrettyPrint() string {
	if vr.Float != None {
		switch vr.Float {
		case Major:
			return "Latest version"
		case Minor:
			if vr.MinVersion != nil {
				return fmt.Sprintf("Latest minor version >= %s", vr.MinVersion)
			}
			return "Latest minor version"
		case Patch:
			if vr.MinVersion != nil {
				return fmt.Sprintf("Latest patch version >= %s", vr.MinVersion)
			}
			return "Latest patch version"
		case Prerelease:
			if vr.MinVersion != nil {
				return fmt.Sprintf("Latest prerelease version >= %s", vr.MinVersion)
			}
			return "Latest prerelease version"
		}
	}

	if vr.MinVersion == nil && vr.MaxVersion == nil {
		return "Any version"
	}

	if vr.MinVersion != nil && vr.MaxVersion != nil && vr.MinVersion.Equal(vr.MaxVersion.Version) {
		return fmt.Sprintf("Version %s exactly", vr.MinVersion)
	}

	var result strings.Builder
	if vr.MinVersion != nil {
		if vr.IncludeMin {
			result.WriteString(fmt.Sprintf(">= %s", vr.MinVersion))
		} else {
			result.WriteString(fmt.Sprintf("> %s", vr.MinVersion))
		}
	}

	if vr.MaxVersion != nil {
		if result.Len() > 0 {
			result.WriteString(" and ")
		}
		if vr.IncludeMax {
			result.WriteString(fmt.Sprintf("<= %s", vr.MaxVersion))
		} else {
			result.WriteString(fmt.Sprintf("< %s", vr.MaxVersion))
		}
	}

	return result.String()
}
