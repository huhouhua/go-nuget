// Copyright (c) 2025 Kevin Berger <huhouhuam@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package version

// VersionRangeBase A base version range that handles ranges only and not any of the preferred version logic.
type VersionRangeBase struct {
	// MaxVersion Maximum version allowed by this range.
	MaxVersion *Version
	// MinVersion Minimum version allowed by this range.
	MinVersion        *Version
	includeMaxVersion bool
	includeMinVersion bool
}

// HasLowerBound True if MinVersion exists;
func (v *VersionRangeBase) HasLowerBound() bool {
	return v.MinVersion != nil
}

// HasUpperBound True if MaxVersion exists.
func (v *VersionRangeBase) HasUpperBound() bool {
	return v.MaxVersion != nil
}

// HasLowerAndUpperBounds True if both MinVersion and MaxVersion exist.
func (v *VersionRangeBase) HasLowerAndUpperBounds() bool {
	return v.HasLowerBound() && v.HasUpperBound()
}

// IsMinInclusive True if MinVersion exists and is included in the range.
func (v *VersionRangeBase) IsMinInclusive() bool {
	return v.HasLowerBound() && v.includeMinVersion
}

// IsMaxInclusive True if MaxVersion exists and is included in the range.
func (v *VersionRangeBase) IsMaxInclusive() bool {
	return v.HasUpperBound() && v.includeMaxVersion
}

// Satisfies Determines if an Version meets the requirements using the version comparer.
func (v *VersionRangeBase) Satisfies(version *Version) bool {
	if version == nil {
		return false
	}
	// Determine if version is in the given range using the comparer.
	condition := true
	if v.HasLowerBound() {
		if v.IsMinInclusive() {
			condition = condition && v.MinVersion.Semver.Compare(version.Semver) <= 0
		} else {
			condition = condition && v.MinVersion.Semver.Compare(version.Semver) < 0
		}
	}
	if v.HasUpperBound() {
		if v.IsMaxInclusive() {
			condition = condition && v.MaxVersion.Semver.Compare(version.Semver) >= 0
		} else {
			condition = condition && v.MaxVersion.Semver.Compare(version.Semver) > 0
		}
	}
	return condition
}
