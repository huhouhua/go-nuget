// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package version

import (
	"strconv"
	"strings"
)

// Format a version range string.
func Format(format string, versionRange VersionRange) (string, error) {
	if strings.TrimSpace(format) == "" {
		format = "N"
	}
	builder := &strings.Builder{}
	runes := []rune(format)
	for i := 0; i < len(runes); i++ {
		formatVersion(builder, runes[i], &versionRange)
	}
	return builder.String(), nil
}

func formatVersion(builder *strings.Builder, r rune, versionRange *VersionRange) {
	switch r {
	case 'P':
		prettyPrint(builder, versionRange, true)
	case 'p':
		prettyPrint(builder, versionRange, false)
	case 'L':
		if versionRange.HasLowerBound() {
			appendNormalized(builder, versionRange.MinVersion)
		}
	case 'U':
		if versionRange.HasUpperBound() {
			appendNormalized(builder, versionRange.MaxVersion)
		}
	case 'S':
		getToString(builder, versionRange)
	case 'N':
		getNormalizedString(builder, versionRange)
	case 'D':
		getLegacyString(builder, versionRange)
	case 'T':
		getLegacyShortString(builder, versionRange)
	case 'A':
		getShortString(builder, versionRange)
	default:
		builder.WriteRune(r)
	}
}

func getShortString(builder *strings.Builder, versionRange *VersionRange) {
	if versionRange.HasLowerBound() && versionRange.IsMinInclusive() && !versionRange.HasUpperBound() {
		if versionRange.IsFloating() {
			versionRange.Float.string(builder)
		} else {
			appendNormalized(builder, versionRange.MinVersion)
		}
	} else if versionRange.HasLowerAndUpperBounds() && versionRange.IsMinInclusive() && versionRange.IsMaxInclusive() && versionRange.MinVersion.Semver.Equal(versionRange.MaxVersion.Semver) {
		// Floating should be ignored here.
		builder.WriteString("[")
		appendNormalized(builder, versionRange.MinVersion)
		builder.WriteString("]")
	} else {
		getNormalizedString(builder, versionRange)
	}
}

// getNormalizedString Builds a normalized string with no short hand
func getNormalizedString(builder *strings.Builder, versionRange *VersionRange) {
	if versionRange.HasLowerBound() && versionRange.IsMinInclusive() {
		builder.WriteString("[")
	} else {
		builder.WriteString("(")
	}
	if versionRange.HasLowerBound() {
		if versionRange.IsFloating() {
			versionRange.Float.string(builder)
		} else {
			appendNormalized(builder, versionRange.MinVersion)
		}
	}
	builder.WriteString(", ")
	if versionRange.HasUpperBound() {
		appendNormalized(builder, versionRange.MaxVersion)
	}
	if versionRange.HasUpperBound() && versionRange.IsMaxInclusive() {
		builder.WriteString("]")
	} else {
		builder.WriteString(")")
	}
}

// getToString Builds a string to represent the VersionRange. This string can include short hand notations.
func getToString(builder *strings.Builder, versionRange *VersionRange) {
	if versionRange.HasLowerBound() && versionRange.IsMinInclusive() && !versionRange.HasUpperBound() {
		appendNormalized(builder, versionRange.MinVersion)
	} else if versionRange.HasLowerAndUpperBounds() && versionRange.IsMaxInclusive() && versionRange.IsMinInclusive() && versionRange.MinVersion.Semver.Equal(versionRange.MaxVersion.Semver) {
		// TODO: Does this need a specific version comparison? Does metadata matter?
		builder.WriteString("[")
		appendNormalized(builder, versionRange.MinVersion)
		builder.WriteString("]")
	} else {
		getNormalizedString(builder, versionRange)
	}
}

// Creates a legacy short string that is compatible with NuGet 2.8.3
func getLegacyShortString(builder *strings.Builder, versionRange *VersionRange) {
	if versionRange.HasLowerBound() && versionRange.IsMinInclusive() && !versionRange.HasUpperBound() {
		appendNormalized(builder, versionRange.MinVersion)
	} else if versionRange.HasLowerAndUpperBounds() && versionRange.IsMinInclusive() && versionRange.IsMaxInclusive() && versionRange.MinVersion.Semver.Equal(versionRange.MaxVersion.Semver) {
		builder.WriteString("[")
		appendNormalized(builder, versionRange.MinVersion)
		builder.WriteString("]")
	} else {
		getLegacyString(builder, versionRange)
	}
}

// Creates a legacy string that is compatible with NuGet 2.8.3
func getLegacyString(builder *strings.Builder, versionRange *VersionRange) {
	if versionRange.HasLowerBound() && versionRange.IsMinInclusive() {
		builder.WriteString("[")
	} else {
		builder.WriteString("(")
	}
	if versionRange.HasLowerBound() {
		appendNormalized(builder, versionRange.MinVersion)
	}
	builder.WriteString(", ")
	if versionRange.HasUpperBound() {
		appendNormalized(builder, versionRange.MaxVersion)
	}
	if versionRange.HasUpperBound() && versionRange.IsMaxInclusive() {
		builder.WriteString("]")
	} else {
		builder.WriteString(")")
	}
}

// prettyPrint A pretty print representation of the VersionRange.
func prettyPrint(builder *strings.Builder, versionRange *VersionRange, useParentheses bool) {
	if !versionRange.HasLowerBound() && !versionRange.HasUpperBound() {
		// empty range
		return
	}
	if useParentheses {
		builder.WriteString("(")
	}
	if versionRange.HasLowerAndUpperBounds() && versionRange.MaxVersion.Semver.Equal(versionRange.MinVersion.Semver) &&
		versionRange.IsMinInclusive() &&
		versionRange.IsMaxInclusive() {
		// single version
		builder.WriteString("= ")
		appendNormalized(builder, versionRange.MinVersion)
	} else {
		// normal case with a lower, upper, or both.
		if versionRange.HasLowerBound() {
			prettyPrintBound(builder, versionRange.MinVersion, versionRange.IsMinInclusive(), ">")
		}
		if versionRange.HasLowerAndUpperBounds() {
			builder.WriteString(" && ")
		}
		if versionRange.HasUpperBound() {
			prettyPrintBound(builder, versionRange.MaxVersion, versionRange.IsMaxInclusive(), "<")
		}
	}
	if useParentheses {
		builder.WriteString(")")
	}
}

func prettyPrintBound(builder *strings.Builder, version *Version, inclusive bool, boundChar string) {
	builder.WriteString(boundChar)
	if inclusive {
		builder.WriteString("= ")
	} else {
		builder.WriteString(" ")
	}
	appendNormalized(builder, version)
}

// appendNormalized Appends a normalized version string. This string is unique for each version 'identity'
// and does not include leading zeros or metadata.
func appendNormalized(builder *strings.Builder, version *Version) {
	appendVersion(builder, version)
	if strings.TrimSpace(version.Semver.Prerelease()) != "" {
		builder.WriteString("-")
		builder.WriteString(version.Semver.Prerelease())
	}
}

func appendVersion(builder *strings.Builder, version *Version) {
	builder.WriteString(strconv.FormatInt(int64(version.Semver.Major()), 36))
	builder.WriteString(".")
	builder.WriteString(strconv.FormatInt(int64(version.Semver.Minor()), 36))
	builder.WriteString(".")
	builder.WriteString(strconv.FormatInt(int64(version.Semver.Patch()), 36))
	if version.IsLegacyVersion() {
		builder.WriteString(".")
		builder.WriteString(strconv.FormatInt(int64(version.Revision), 36))
	}
}
