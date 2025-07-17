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
			includeMaxVersion: includeMaxVersion,
			includeMinVersion: includeMinVersion,
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

// IsFloating True if the range has a floating version above the min version.
func (r *VersionRange) IsFloating() bool {
	return r.Float != nil && r.Float.FloatBehavior != None
}

// String Normalized range string.
func (r *VersionRange) String() (string, error) {
	return r.ToNormalizedString()
}

// ToNormalizedString Normalized range string.
func (r *VersionRange) ToNormalizedString() (string, error) {
	return Format("N", *r)
}

// PrettyPrint format the version range in Pretty Print format.
func (r *VersionRange) PrettyPrint() (string, error) {
	return Format("P", *r)
}

// ToLegacyString A legacy version range compatible with NuGet 2.8.3
func (r *VersionRange) ToLegacyString() (string, error) {
	return Format("D", *r)
}

// ToLegacyShortString A short legacy version range compatible with NuGet 2.8.3.
// Ex: 1.0.0
func (r *VersionRange) ToLegacyShortString() (string, error) {
	return Format("T", *r)
}
