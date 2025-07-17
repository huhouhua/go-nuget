// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package version

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionRange_MetadataIsIgnored_FormatRemovesMetadata(t *testing.T) {
	// Arrange
	bothMetadata, err := ParseRange("[1.0.0+A, 2.0.0+A]")
	require.NoError(t, err)

	str, err := bothMetadata.String()
	require.NoError(t, err)

	normalized, err := bothMetadata.ToNormalizedString()
	require.NoError(t, err)

	legacyString, err := bothMetadata.ToLegacyString()
	require.NoError(t, err)

	// Act & Assert
	require.Equal(t, "[1.0.0, 2.0.0]", str)
	require.Equal(t, "[1.0.0, 2.0.0]", normalized)
	require.Equal(t, "[1.0.0, 2.0.0]", legacyString)
}

func TestVersionRange_FloatAllStable_ReturnsCorrectPrints(t *testing.T) {
	// Arrange
	bothMetadata, err := ParseRange("*")
	require.NoError(t, err)

	str, err := bothMetadata.String()
	require.NoError(t, err)

	normalized, err := bothMetadata.ToNormalizedString()
	require.NoError(t, err)

	legacyString, err := bothMetadata.ToLegacyString()
	require.NoError(t, err)

	// Act & Assert
	require.Equal(t, "[*, )", str)
	require.Equal(t, "[*, )", normalized)
	require.Equal(t, "[0.0.0, )", legacyString)
}

func TestVersionRange_MetadataIsIgnored_FormatRemovesMetadata_Short(t *testing.T) {
	// Arrange
	bothMetadata, err := ParseRange("[1.0.0+A, )")
	require.NoError(t, err)

	short, err := bothMetadata.ToLegacyShortString()
	require.NoError(t, err)

	// Act & Assert
	require.Equal(t, "1.0.0", short)
}

func TestVersionRange_MetadataIsIgnored_FormatRemovesMetadata_PrettyPrint(t *testing.T) {
	// Arrange
	bothMetadata, err := ParseRange("[1.0.0+A, )")
	require.NoError(t, err)

	prettyStr, err := bothMetadata.PrettyPrint()
	require.NoError(t, err)

	// Act & Assert
	require.Equal(t, "(>= 1.0.0)", prettyStr)
}

func TestVersionRange_IncludePrerelease(t *testing.T) {
	tests := []struct {
		version string
	}{
		{
			version: "[1.0.0]",
		},
		{
			version: "[1.0.0, 2.0.0]",
		},
		{
			version: "1.0.0",
		},
		{
			version: "1.0.0-beta",
		},
		{
			version: "(1.0.0-beta, 2.0.0-alpha)",
		},
		{
			version: "(1.0.0-beta, 2.0.0)",
		},
		{
			version: "(1.0.0, 2.0.0-alpha)",
		},
		{
			version: "1.0.0-beta-*",
		},
		{
			version: "[1.0.0-beta-*, ]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			// Arrange
			versionRange, err := ParseRange(tt.version)
			require.NoError(t, err)

			normalized1, err := versionRange.ToNormalizedString()
			require.NoError(t, err)
			normalized2, err := versionRange.ToNormalizedString()
			require.NoError(t, err)
			// Act && Assert
			require.Equal(t, versionRange.IsFloating(), versionRange.IsFloating())
			require.Equal(t, versionRange.Float, versionRange.Float)
			require.Equal(t, normalized1, normalized2)
		})
	}
}
func TestParseVersionRangeSingleDigit(t *testing.T) {
	// Act
	versionInfo, err := ParseRange("[1,3)")
	require.NoError(t, err)

	require.Equal(t, "1.0.0", versionInfo.MinVersion.Semver.String())
	require.True(t, versionInfo.IsMinInclusive())
	require.Equal(t, "3.0.0", versionInfo.MaxVersion.Semver.String())
	require.False(t, versionInfo.IsMaxInclusive())
}

func TestVersionRange_MissingVersionComponents_DefaultToZero(t *testing.T) {
	tests := []struct {
		shortVersionSpec string
		longVersionSpec  string
	}{
		{
			shortVersionSpec: "0",
			longVersionSpec:  "0.0",
		},
		{
			shortVersionSpec: "1",
			longVersionSpec:  "1.0.0",
		},
		{
			shortVersionSpec: "02",
			longVersionSpec:  "2.0.0.0",
		},
		{
			shortVersionSpec: "123.456",
			longVersionSpec:  "123.456.0.0",
		},
		{
			shortVersionSpec: "[2021,)",
			longVersionSpec:  "[2021.0.0.0,)",
		},
		{
			shortVersionSpec: "[,2021)",
			longVersionSpec:  "[,2021.0.0.0)",
		},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("%s=%s", tt.shortVersionSpec, tt.longVersionSpec)
		t.Run(name, func(t *testing.T) {
			// Act
			versionRange1, err := ParseRange(tt.shortVersionSpec)
			require.NoError(t, err)
			versionRange2, err := ParseRange(tt.longVersionSpec)
			require.NoError(t, err)

			// Assert
			str1, err := versionRange1.String()
			require.NoError(t, err)
			str2, err := versionRange2.String()
			require.NoError(t, err)
			require.Equal(t, str1, str2)
		})
	}
}

func TestParseVersionRangeParts(t *testing.T) {
	tests := []struct {
		minString string
		maxString string
		minInc    bool
		maxInc    bool
	}{
		{
			minString: "1.0.0",
			maxString: "2.0.0",
			minInc:    true,
			maxInc:    true,
		},
		{
			minString: "1.0.0",
			maxString: "1.0.1",
			minInc:    false,
			maxInc:    false,
		},
		{
			minString: "1.0.0-beta+0",
			maxString: "2.0.0",
			minInc:    false,
			maxInc:    true,
		},
		{
			minString: "1.0.0-beta+0",
			maxString: "2.0.0+99",
			minInc:    false,
			maxInc:    false,
		},
		{
			minString: "1.0.0-beta+0",
			maxString: "2.0.0+99",
			minInc:    true,
			maxInc:    true,
		},
		{
			minString: "1.0.0",
			maxString: "2.0.0+99",
			minInc:    true,
			maxInc:    true,
		},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("max:%s min:%s", tt.maxString, tt.minString)
		t.Run(name, func(t *testing.T) {
			// Arrange
			minVersion, err := Parse(tt.minString)
			require.NoError(t, err)
			maxVersion, err := Parse(tt.maxString)
			require.NoError(t, err)

			// Act
			versionInfo, err := NewVersionRange(minVersion, maxVersion, tt.minInc, tt.maxInc, nil, "")
			require.NoError(t, err)

			// Assert
			require.Equal(t, minVersion, versionInfo.MinVersion)
			require.Equal(t, maxVersion, versionInfo.MaxVersion)
			require.Equal(t, tt.minInc, versionInfo.IsMinInclusive())
			require.Equal(t, tt.maxInc, versionInfo.IsMaxInclusive())
		})
	}
}

func TestParseVersionRangeWithNullThrows(t *testing.T) {
	versionInfo, err := ParseRange("")
	require.Nil(t, versionInfo)
	require.Equal(t, fmt.Errorf("'' is not a valid version string"), err)
}

func TestParseVersionRangeWithBadVersionThrows(t *testing.T) {
	tests := []struct {
		versionRange string
	}{
		{
			versionRange: "",
		},
		{
			versionRange: "      ",
		},
		{
			versionRange: "-1",
		},
		{
			versionRange: "+1",
		},
		{
			versionRange: "1.",
		},
		{
			versionRange: ".1",
		},
		{
			versionRange: "1,",
		},
		{
			versionRange: ",1",
		},
		{
			versionRange: ",",
		},
		{
			versionRange: "-",
		},
		{
			versionRange: "+",
		},
		{
			versionRange: "a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.versionRange, func(t *testing.T) {
			// Act & Assert
			versionInfo, err := ParseRange(tt.versionRange)
			require.Nil(t, versionInfo)
			require.Equal(t, fmt.Errorf("'%s' is not a valid version string", tt.versionRange), err)
		})
	}
}

func TestParse_Illogical_VersionRange_Throws(t *testing.T) {
	tests := []struct {
		versionRange string
	}{
		{
			versionRange: "[1.1.4, 1.1.2)",
		},
		{
			versionRange: "[1.1.4, 1.1.2]",
		},
		{
			versionRange: "(1.1.4, 1.1.2)",
		},
		{
			versionRange: "(1.1.4, 1.1.2]",
		},
		{
			versionRange: "[1.0.0, 1.0.0)",
		},
		{
			versionRange: "(1.0.0, 1.0.0]",
		},
		{
			versionRange: "(1.0, 1.0.0]",
		},
		{
			versionRange: "(*, *]",
		},
		{
			versionRange: "[1.0.0-beta, 1.0.0-beta+900)",
		},
		{
			versionRange: "(1.0.0-beta+600, 1.0.0-beta]",
		},
		{
			versionRange: "(1.0)",
		},
		{
			versionRange: "(1.0.0)",
		},
		{
			versionRange: "[2.0.0)",
		},
		{
			versionRange: "(2.0.0]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.versionRange, func(t *testing.T) {
			// Act & Assert
			versionInfo, err := ParseRange(tt.versionRange)
			require.Nil(t, versionInfo)
			require.Equal(t, fmt.Errorf("'%s' is not a valid version string", tt.versionRange), err)
		})
	}

}

func TestParseVersionRangeIntegerRanges(t *testing.T) {
	// Act
	versionInfo, err := ParseRange("   [-1, 2]  ")
	require.Nil(t, versionInfo)
	require.Equal(t, fmt.Errorf("'   [-1, 2]  ' is not a valid version string"), err)
}

func TestParseVersionRangeNegativeIntegerRanges(t *testing.T) {
	// Act
	versionInfo, parsed := TryParseRange("   [-1, 2]  ", true)
	require.False(t, parsed)
	require.Nil(t, versionInfo)
}

func TestParseVersionThrowsIfExclusiveMinAndMaxVersionRangeContainsNoValues(t *testing.T) {
	// Arrange
	versionInfo, err := ParseRange("(,)")
	require.Nil(t, versionInfo)
	// Assert
	require.Equal(t, fmt.Errorf("'(,)' is not a valid version string"), err)
}
func TestParseVersionThrowsIfInclusiveMinAndMaxVersionRangeContainsNoValues(t *testing.T) {
	// Arrange
	versionInfo, err := ParseRange("[,]")
	require.Nil(t, versionInfo)
	// Assert
	require.Equal(t, fmt.Errorf("'[,]' is not a valid version string"), err)
}
func TestParseVersionThrowsIfInclusiveMinAndExclusiveMaxVersionRangeContainsNoValues(t *testing.T) {
	// Arrange
	versionInfo, err := ParseRange("[,)")
	require.Nil(t, versionInfo)
	// Assert
	require.Equal(t, fmt.Errorf("'[,)' is not a valid version string"), err)
}
func TestParseVersionThrowsIfExclusiveMinAndInclusiveMaxVersionRangeContainsNoValues(t *testing.T) {
	// Arrange
	versionInfo, err := ParseRange("(,]")
	require.Nil(t, versionInfo)
	// Assert
	require.Equal(t, fmt.Errorf("'(,]' is not a valid version string"), err)
}
func TestParseVersionThrowsIfVersionRangeIsMissingVersionComponent(t *testing.T) {
	// Arrange
	versionInfo, err := ParseRange("(,1.3..2]")
	require.Nil(t, versionInfo)
	// Assert
	require.Equal(t, fmt.Errorf("'(,1.3..2]' is not a valid version string"), err)
}
func TestParseVersionThrowsIfVersionRangeContainsMoreThen4VersionComponents(t *testing.T) {
	// Arrange
	versionInfo, err := ParseRange("(1.2.3.4.5,1.2]")
	require.Nil(t, versionInfo)
	// Assert
	require.Equal(t, fmt.Errorf("'(1.2.3.4.5,1.2]' is not a valid version string"), err)
}
