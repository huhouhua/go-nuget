// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package version

import (
	"fmt"
)

// VersionRange represents a range of versions that satisfy a given constraint.
// This is similar to NuGet's VersionRange class.
type VersionRange struct {
	*VersionRangeBase

	// Float Optional floating range used to determine the best version match.
	Float *FloatRange

	// OriginalString Original string being parsed to this object.
	OriginalString string
}

// NewVersionRangeFromMinVersion Creates a range that is greater than or equal to the minVersion with the given float
// behavior.
func NewVersionRangeFromMinVersion(minVersion *Version, floatRange *FloatRange) (*VersionRange, error) {
	return NewVersionRange(minVersion, nil, true, false, floatRange, "")
}

// NewVersionRange Creates a VersionRange with the given min and max.
func NewVersionRange(
	minVersion, maxVersion *Version,
	includeMinVersion, includeMaxVersion bool,
	floatRange *FloatRange,
	originalString string,
) (*VersionRange, error) {
	if floatRange != nil && minVersion == nil {
		return nil, fmt.Errorf("parameter 'minVersion' cannot be nil when parameter 'floatRange' is not nil")
	}
	v := &VersionRange{
		VersionRangeBase: &VersionRangeBase{
			includeMaxVersion: includeMinVersion,
			includeMinVersion: includeMaxVersion,
			MaxVersion:        maxVersion,
			MinVersion:        minVersion,
		},
		Float:          floatRange,
		OriginalString: originalString,
	}
	return v, nil
}

func (r *VersionRange) DoesRangeSatisfy(catalogItemLower, catalogItemUpper *Version) bool {
	// Mainly to cover the '!dependencyRange.IsMaxInclusive() && !dependencyRange.IsMinInclusive()' case
	if r.HasLowerAndUpperBounds() {
		if catalogItemVersionRange, err := NewVersionRange(catalogItemLower, catalogItemUpper, true, true, nil, ""); err != nil {
			return false
		} else {
			return catalogItemVersionRange.Satisfies(r.MinVersion) || catalogItemVersionRange.Satisfies(r.MaxVersion)
		}
	} else {
		return r.Satisfies(catalogItemLower) || r.Satisfies(catalogItemUpper)
	}
}

//// parseFloatingVersion parses a floating version string
//func parseFloatingVersion(rangeStr string) (*VersionRange, error) {
//	switch {
//	case rangeStr == "*":
//		return &VersionRange{Float: Major}, nil
//	case strings.HasPrefix(rangeStr, "*-"):
//		return &VersionRange{Float: Prerelease}, nil
//	case strings.Contains(rangeStr, "-*"):
//		baseVersion := strings.TrimSuffix(rangeStr, "-*")
//		v, err := semver.NewVersion(baseVersion)
//		if err != nil {
//			return nil, fmt.Errorf("invalid version in prerelease range: %s", err.Error())
//		}
//		vr := NewVersionRange(v, nil, true, false)
//		vr.Float = Prerelease
//		return vr, nil
//	case strings.HasPrefix(rangeStr, "~") || strings.HasPrefix(rangeStr, "^"):
//		return parseFloatingRange(rangeStr)
//	default:
//		return nil, fmt.Errorf("invalid floating version format: %s", rangeStr)
//	}
//}
//
//// parseFloatingRange parses a floating range string
//func parseFloatingRange(rangeStr string) (*VersionRange, error) {
//	var prefix string
//	var floatType FloatBehavior
//
//	switch {
//	case strings.HasPrefix(rangeStr, "~"):
//		prefix = "~"
//		floatType = Patch
//	case strings.HasPrefix(rangeStr, "^"):
//		prefix = "^"
//		floatType = Minor
//	default:
//		return nil, fmt.Errorf("invalid floating range format: %s", rangeStr)
//	}
//
//	v, err := semver.NewVersion(strings.TrimPrefix(rangeStr, prefix))
//	if err != nil {
//		return nil, fmt.Errorf("invalid version in %s range: %s", prefix, err.Error())
//	}
//
//	var maxVersion *semver.Version
//	if floatType == Patch {
//		maxVersion = semver.New(v.Major(), v.Minor()+1, 0, "", "")
//	} else { // Minor
//		if v.Major() == 0 {
//			maxVersion = semver.New(v.Major(), v.Minor()+1, 0, "", "")
//		} else {
//			maxVersion = semver.New(v.Major()+1, 0, 0, "", "")
//		}
//	}
//
//	vr := NewVersionRange(v, maxVersion, true, false)
//	vr.Float = floatType
//	return vr, nil
//}
//
//// Satisfies checks if a version satisfies this version range
//func (vr *VersionRange) Satisfies(v *semver.Version) bool {
//	if vr.Float != None {
//		return vr.satisfiesFloat(v)
//	}
//
//	if vr.MinVersion != nil {
//		if vr.IncludeMin {
//			if v.LessThan(vr.MinVersion) {
//				return false
//			}
//		} else {
//			if !v.GreaterThan(vr.MinVersion) {
//				return false
//			}
//		}
//	}
//
//	if vr.MaxVersion != nil {
//		if vr.IncludeMax {
//			if v.GreaterThan(vr.MaxVersion) {
//				return false
//			}
//		} else {
//			if !v.LessThan(vr.MaxVersion) {
//				return false
//			}
//		}
//	}
//
//	return true
//}
//
//// satisfiesFloat checks if a version satisfies the floating behavior
//func (vr *VersionRange) satisfiesFloat(v *semver.Version) bool {
//	switch vr.Float {
//	case Major:
//		return true
//	case Prerelease:
//		return v.Prerelease() != ""
//	case Patch:
//		return v.Major() == vr.MinVersion.Major() &&
//			v.Minor() == vr.MinVersion.Minor() &&
//			v.Patch() >= vr.MinVersion.Patch()
//	case Minor:
//		if vr.MinVersion.Major() == 0 {
//			return v.Major() == 0 &&
//				v.Minor() == vr.MinVersion.Minor() &&
//				v.Patch() >= vr.MinVersion.Patch()
//		}
//		return v.Major() == vr.MinVersion.Major() &&
//			v.Minor() >= vr.MinVersion.Minor()
//	default:
//		return false
//	}
//}
//
//// DoesRangeSatisfy checks if this version range satisfies another version range
//func (vr *VersionRange) DoesRangeSatisfy(other *VersionRange) bool {
//	// If this range has both lower and upper bounds
//	if vr.MinVersion != nil && vr.MaxVersion != nil {
//		// Create a new range with the bounds of this range
//		rangeWithBounds := &VersionRange{
//			MinVersion: vr.MinVersion,
//			MaxVersion: vr.MaxVersion,
//			IncludeMin: true,
//			IncludeMax: true,
//		}
//
//		// Check if either the min or max version of the other range satisfies this range
//		return rangeWithBounds.Satisfies(other.MinVersion) ||
//			rangeWithBounds.Satisfies(other.MaxVersion)
//	} else {
//		// If this range doesn't have both bounds, check if either bound of the other range satisfies this range
//		return vr.Satisfies(other.MinVersion) || vr.Satisfies(other.MaxVersion)
//	}
//}
//
//// String returns the string representation of the version range
//func (vr *VersionRange) String() string {
//	if vr.Float != None {
//		return vr.stringFloat()
//	}
//
//	if vr.MinVersion != nil && vr.MaxVersion != nil && vr.MinVersion.Equal(vr.MaxVersion) {
//		return vr.MinVersion.String()
//	}
//
//	var sb strings.Builder
//	if vr.MinVersion != nil {
//		if vr.IncludeMin {
//			sb.WriteString("[")
//		} else {
//			sb.WriteString("(")
//		}
//		sb.WriteString(vr.MinVersion.String())
//	} else {
//		sb.WriteString("(")
//	}
//
//	sb.WriteString(",")
//
//	if vr.MaxVersion != nil {
//		sb.WriteString(vr.MaxVersion.String())
//		if vr.IncludeMax {
//			sb.WriteString("]")
//		} else {
//			sb.WriteString(")")
//		}
//	} else {
//		sb.WriteString(")")
//	}
//
//	return sb.String()
//}
//
//// stringFloat returns the string representation of a floating version range
//func (vr *VersionRange) stringFloat() string {
//	switch vr.Float {
//	case Major:
//		return "*"
//	case Prerelease:
//		return "*-"
//	case Patch:
//		return "~" + vr.MinVersion.String()
//	case Minor:
//		return "^" + vr.MinVersion.String()
//	default:
//		return ""
//	}
//}
//
//// FindBestMatch returns the version that best matches the range from a list of versions
//func (vr *VersionRange) FindBestMatch(versions []*semver.Version) *semver.Version {
//	var bestMatch *semver.Version
//	for _, version := range versions {
//		if vr.IsBetter(bestMatch, version) {
//			bestMatch = version
//		}
//	}
//	return bestMatch
//}
//
//// IsBetter determines if a given version is better suited to the range than a current version
//func (vr *VersionRange) IsBetter(current, considering *semver.Version) bool {
//	if current == considering {
//		return false
//	}
//
//	// null checks
//	if considering == nil {
//		return false
//	}
//
//	// If the range contains only stable versions disallow prerelease versions
//	if !vr.hasPrereleaseBounds() && considering.Prerelease() != "" &&
//		vr.Float != Prerelease && vr.Float != Major {
//		return false
//	}
//
//	if !vr.Satisfies(considering) {
//		// keep null over a value outside of the range
//		return false
//	}
//
//	if current == nil {
//		return true
//	}
//
//	if vr.Float != None {
//		// check if either version is in the floating range
//		curInRange := vr.satisfiesFloat(current)
//
//		if curInRange {
//			// prefer the highest one if both are in the range
//			return current.LessThan(considering)
//		} else {
//			// neither are in range
//			curToLower := current.LessThan(vr.MinVersion)
//			conToLower := considering.LessThan(vr.MinVersion)
//
//			if curToLower && !conToLower {
//				// favor the version above the range
//				return true
//			} else if !curToLower && conToLower {
//				return false
//			} else if !curToLower {
//				// favor the lower version if we are above the range
//				return current.GreaterThan(considering)
//			} else {
//				// favor the higher version if we are below the range
//				return current.LessThan(considering)
//			}
//		}
//	}
//
//	// Favor lower versions
//	return current.GreaterThan(considering)
//}
//
//// hasPrereleaseBounds returns true if either bound is a prerelease version
//func (vr *VersionRange) hasPrereleaseBounds() bool {
//	return (vr.MinVersion != nil && vr.MinVersion.Prerelease() != "") ||
//		(vr.MaxVersion != nil && vr.MaxVersion.Prerelease() != "")
//}
//
//// ToNonSnapshotRange removes the floating snapshot part of the minimum version if it exists
//func (vr *VersionRange) ToNonSnapshotRange() *VersionRange {
//	if vr.MinVersion == nil && vr.MaxVersion == nil {
//		return vr
//	}
//
//	minVersion := vr.MinVersion
//	maxVersion := vr.MaxVersion
//
//	// Handle floating versions
//	if vr.Float != None {
//		if minVersion != nil {
//			var major, minor uint64
//			if vr.Float == Minor {
//				major = minVersion.Major() + 1
//				minor = 0
//			} else if vr.Float == Patch {
//				major = minVersion.Major()
//				minor = minVersion.Minor() + 1
//			}
//			maxVersion = semver.New(major, minor, 0, "", "")
//		}
//		return NewVersionRange(minVersion, maxVersion, true, false)
//	}
//
//	// Handle prerelease versions
//	if minVersion != nil {
//		minVersion = processPrereleaseVersion(minVersion)
//	}
//	if maxVersion != nil {
//		maxVersion = processPrereleaseVersion(maxVersion)
//	}
//
//	// Create new version range with original include flags
//	return NewVersionRange(minVersion, maxVersion, vr.IncludeMin, vr.IncludeMax)
//}
//
//// processPrereleaseVersion processes a version with prerelease information
//func processPrereleaseVersion(v *semver.Version) *semver.Version {
//	if v == nil || v.Prerelease() == "" {
//		return v
//	}
//
//	prerelease := strings.TrimRight(v.Prerelease(), "-")
//	if prerelease == "0" {
//		return semver.New(v.Major(), v.Minor(), v.Patch(), "", "")
//	}
//	return semver.New(v.Major(), v.Minor(), v.Patch(), prerelease, "")
//}
//
//// PrettyPrint returns a human-readable string representation of the version range
//func (vr *VersionRange) PrettyPrint() string {
//	if vr.Float != None {
//		switch vr.Float {
//		case Major:
//			return "Latest version"
//		case Minor:
//			if vr.MinVersion != nil {
//				return fmt.Sprintf("Latest minor version >= %s", vr.MinVersion)
//			}
//			return "Latest minor version"
//		case Patch:
//			if vr.MinVersion != nil {
//				return fmt.Sprintf("Latest patch version >= %s", vr.MinVersion)
//			}
//			return "Latest patch version"
//		case Prerelease:
//			if vr.MinVersion != nil {
//				return fmt.Sprintf("Latest prerelease version >= %s", vr.MinVersion)
//			}
//			return "Latest prerelease version"
//		}
//	}
//
//	if vr.MinVersion == nil && vr.MaxVersion == nil {
//		return "Any version"
//	}
//
//	if vr.MinVersion != nil && vr.MaxVersion != nil && vr.MinVersion.Equal(vr.MaxVersion) {
//		return fmt.Sprintf("Version %s exactly", vr.MinVersion)
//	}
//
//	var result strings.Builder
//	if vr.MinVersion != nil {
//		if vr.IncludeMin {
//			result.WriteString(fmt.Sprintf(">= %s", vr.MinVersion))
//		} else {
//			result.WriteString(fmt.Sprintf("> %s", vr.MinVersion))
//		}
//	}
//
//	if vr.MaxVersion != nil {
//		if result.Len() > 0 {
//			result.WriteString(" and ")
//		}
//		if vr.IncludeMax {
//			result.WriteString(fmt.Sprintf("<= %s", vr.MaxVersion))
//		} else {
//			result.WriteString(fmt.Sprintf("< %s", vr.MaxVersion))
//		}
//	}
//
//	return result.String()
//}
