// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package version

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionRange_PrettyPrint(t *testing.T) {
	tests := []struct {
		version  string
		expected string
	}{
		{
			version:  "1.0.0",
			expected: "(>= 1.0.0)",
		},
		{
			version:  "[1.0.0]",
			expected: "(= 1.0.0)",
		},
		{
			version:  "[1.0.0, ]",
			expected: "(>= 1.0.0)",
		},
		{
			version:  "[1.0.0, )",
			expected: "(>= 1.0.0)",
		},
		{
			version:  "(1.0.0, )",
			expected: "(> 1.0.0)",
		},
		{
			version:  "(1.0.0, ]",
			expected: "(> 1.0.0)",
		},
		{
			version:  "(1.0.0, 2.0.0)",
			expected: "(> 1.0.0 && < 2.0.0)",
		},
		{
			version:  "[1.0.0, 2.0.0]",
			expected: "(>= 1.0.0 && <= 2.0.0)",
		},
		{
			version:  "[1.0.0, 2.0.0)",
			expected: "(>= 1.0.0 && < 2.0.0)",
		},
		{
			version:  "(1.0.0, 2.0.0]",
			expected: "(> 1.0.0 && <= 2.0.0)",
		},
		{
			version:  "(, 2.0.0]",
			expected: "(<= 2.0.0)",
		},
		{
			version:  "(, 2.0.0)",
			expected: "(< 2.0.0)",
		},
		{
			version:  "[, 2.0.0)",
			expected: "(< 2.0.0)",
		},
		{
			version:  "[, 2.0.0]",
			expected: "(<= 2.0.0)",
		},
		{
			version:  "1.0.0-beta*",
			expected: "(>= 1.0.0-beta)",
		},
		{
			version:  "[1.0.0-beta*, 2.0.0)",
			expected: "(>= 1.0.0-beta && < 2.0.0)",
		},
		{
			version:  "[1.0.0-beta.1, 2.0.0-alpha.2]",
			expected: "(>= 1.0.0-beta.1 && <= 2.0.0-alpha.2)",
		},
		{
			version:  "[1.0.0+beta.1, 2.0.0+alpha.2]",
			expected: "(>= 1.0.0 && <= 2.0.0)",
		},
		{
			version:  "[1.0, 2.0]",
			expected: "(>= 1.0.0 && <= 2.0.0)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			// Arrange
			versionRange, err := ParseRange(tt.version)
			require.NoError(t, err)

			// Act
			s, err := Format("P", *versionRange)
			require.NoError(t, err)
			s2, err := versionRange.PrettyPrint()
			require.NoError(t, err)

			require.Equal(t, tt.expected, s)
			require.Equal(t, tt.expected, s2)
		})
	}
}

func TestVersionRange_NormalizationRoundTrips(t *testing.T) {
	tests := []struct {
		version                    string
		isOriginalStringNormalized bool
	}{
		{
			version:                    "1.0.0",
			isOriginalStringNormalized: false,
		},
		{
			version:                    "1.*",
			isOriginalStringNormalized: false,
		},
		{
			version:                    "*",
			isOriginalStringNormalized: false,
		},
		{
			version:                    "[*, )",
			isOriginalStringNormalized: true,
		},
		{
			version:                    "[1.*, ]",
			isOriginalStringNormalized: false,
		},
		{
			version:                    "[1.*, 2.0.0)",
			isOriginalStringNormalized: true,
		},
		{
			version:                    "(, )",
			isOriginalStringNormalized: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			// Arrange
			originalParsedRange, err := ParseRange(tt.version)
			require.NoError(t, err)

			// Act
			normalizedRangeRepresentation, err := originalParsedRange.ToNormalizedString()
			require.NoError(t, err)
			roundTrippedRange, err := ParseRange(normalizedRangeRepresentation)
			require.NoError(t, err)
			// Assert
			orgStr, err := originalParsedRange.String()
			require.NoError(t, err)
			roundStr, err := roundTrippedRange.String()
			require.NoError(t, err)
			require.Equal(t, orgStr, roundStr)

			roundNormalizedStr, err := roundTrippedRange.ToNormalizedString()
			require.NoError(t, err)
			require.Equal(t, normalizedRangeRepresentation, roundNormalizedStr)
			if tt.isOriginalStringNormalized {
				require.Equal(t, normalizedRangeRepresentation, tt.version)
			} else {
				require.NotEqual(t, normalizedRangeRepresentation, tt.version)
			}
		})
	}
}

func TestVersionRange_PrettyPrintAllRange(t *testing.T) {
	// Arrange
	rangeAll := All

	// Act
	s, err := Format("P", *rangeAll)
	require.NoError(t, err)
	s2, err := rangeAll.PrettyPrint()
	require.NoError(t, err)

	require.Equal(t, "", s)
	require.Equal(t, "", s2)
}

func TestVersionRange_MetadataIsIgnored_Satisfy(t *testing.T) {
	// Arrange
	noMetadata, err := ParseRange("[1.0.0, 2.0.0]")
	require.NoError(t, err)
	lowerMetadata, err := ParseRange("[1.0.0+A, 2.0.0]")
	require.NoError(t, err)
	upperMetadata, err := ParseRange("[1.0.0, 2.0.0+A]")
	require.NoError(t, err)
	bothMetadata, err := ParseRange("[1.0.0+A, 2.0.0+A]")
	require.NoError(t, err)

	versionNoMetadata, err := Parse("1.0.0")
	require.NoError(t, err)
	versionMetadata, err := Parse("1.0.0+B")
	require.NoError(t, err)

	// Act & Assert
	require.True(t, noMetadata.Satisfies(versionNoMetadata))
	require.True(t, noMetadata.Satisfies(versionMetadata))
	require.True(t, lowerMetadata.Satisfies(versionNoMetadata))
	require.True(t, lowerMetadata.Satisfies(versionMetadata))
	require.True(t, upperMetadata.Satisfies(versionNoMetadata))
	require.True(t, upperMetadata.Satisfies(versionMetadata))
	require.True(t, bothMetadata.Satisfies(versionNoMetadata))
	require.True(t, bothMetadata.Satisfies(versionMetadata))
}

func TestVersionRange_MetadataIsIgnored_String(t *testing.T) {
	// Arrange
	noMetadata, err := ParseRange("[1.0.0, 2.0.0]")
	require.NoError(t, err)
	lowerMetadata, err := ParseRange("[1.0.0+A, 2.0.0]")
	require.NoError(t, err)
	upperMetadata, err := ParseRange("[1.0.0, 2.0.0+A]")
	require.NoError(t, err)
	bothMetadata, err := ParseRange("[1.0.0+A, 2.0.0+A]")
	require.NoError(t, err)

	noMetadataStr, err := noMetadata.String()
	require.NoError(t, err)
	lowerMetadataStr, err := lowerMetadata.String()
	require.NoError(t, err)
	upperMetadataStr, err := upperMetadata.String()
	require.NoError(t, err)
	bothMetadataStr, err := bothMetadata.String()
	require.NoError(t, err)

	// Act & Assert
	require.Equal(t, noMetadataStr, lowerMetadataStr)
	require.Equal(t, lowerMetadataStr, upperMetadataStr)
	require.Equal(t, upperMetadataStr, bothMetadataStr)
	require.Equal(t, bothMetadataStr, noMetadataStr)
}

func TestVersionRange_AllSpecialCases_NormalizeSame(t *testing.T) {
	normalizedStr, err := All.ToNormalizedString()
	require.NoError(t, err)
	require.Equal(t, "(, )", normalizedStr)
}

func TestVersionRange_Exact(t *testing.T) {
	// Act
	versionInfo, err := NewVersionRange(
		NewVersionFrom(4, 3, 0, "", ""),
		NewVersionFrom(4, 3, 0, "", ""),
		true,
		true,
		nil,
		"")
	require.NoError(t, err)
	// Assert
	v, err := Parse("4.3.0")
	require.NoError(t, err)
	require.True(t, versionInfo.Satisfies(v))
}

func TestParseVersionRangeDoesNotSatisfy(t *testing.T) {
	tests := []struct {
		spec    string
		version string
	}{
		{
			spec:    "1.0.0",
			version: "0.0.0",
		},
		{
			spec:    "[1.0.0, 2.0.0]",
			version: "2.0.1",
		},
		{
			spec:    "[1.0.0, 2.0.0]",
			version: "0.0.0",
		},
		{
			spec:    "[1.0.0, 2.0.0]",
			version: "3.0.0",
		},
		{
			spec:    "[1.0.0-beta+meta, 2.0.0-beta+meta]",
			version: "1.0.0-alpha",
		},
		{
			spec:    "[1.0.0-beta+meta, 2.0.0-beta+meta]",
			version: "1.0.0-alpha+meta",
		},
		{
			spec:    "[1.0.0-beta+meta, 2.0.0-beta+meta]",
			version: "2.0.0-rc",
		},
		{
			spec:    "[1.0.0-beta+meta, 2.0.0-beta+meta]",
			version: "2.0.0+meta",
		},
		{
			spec:    "(1.0.0-beta+meta, 2.0.0-beta+meta)",
			version: "2.0.0-beta+meta",
		},
		{
			spec:    "(, 2.0.0-beta+meta)",
			version: "2.0.0-beta+meta",
		},
	}
	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			// Act
			versionInfo, err := ParseRange(tt.spec)
			require.NoError(t, err)
			middleVersion, err := Parse(tt.version)
			require.NoError(t, err)

			// Assert
			require.False(t, versionInfo.Satisfies(middleVersion))
		})
	}
}

func TestParseVersionRangeSatisfies(t *testing.T) {
	tests := []struct {
		spec    string
		version string
	}{
		{
			spec:    "1.0.0",
			version: "2.0.0",
		},
		{
			spec:    "[1.0.0, 2.0.0]",
			version: "2.0.0",
		},
		{
			spec:    "(2.0.0,)",
			version: "2.1.0",
		},
		{
			spec:    "[2.0.0]",
			version: "2.0.0",
		},
		{
			spec:    "(,2.0.0]",
			version: "2.0.0",
		},
		{
			spec:    "(,2.0.0]",
			version: "1.0.0",
		},
		{
			spec:    "[2.0.0, )",
			version: "2.0.0",
		},
		{
			spec:    "1.0.0",
			version: "1.0.0",
		},
		{
			spec:    "[1.0.0]",
			version: "1.0.0",
		},
		{
			spec:    "[1.0.0, 1.0.0]",
			version: "1.0.0",
		},
		{
			spec:    "[1.0.0, 2.0.0]",
			version: "1.0.0",
		},
		{
			spec:    "[1.0.0-beta+meta, 2.0.0-beta+meta]",
			version: "1.0.0",
		},
		{
			spec:    "[1.0.0-beta+meta, 2.0.0-beta+meta]",
			version: "1.0.0-beta+meta",
		},
		{
			spec:    "[1.0.0-beta+meta, 2.0.0-beta+meta]",
			version: "2.0.0-beta",
		},
		{
			spec:    "[1.0.0-beta+meta, 2.0.0-beta+meta]",
			version: "1.0.0+meta",
		},
		{
			spec:    "(1.0.0-beta+meta, 2.0.0-beta+meta)",
			version: "1.0.0",
		},
		{
			spec:    "(1.0.0-beta+meta, 2.0.0-beta+meta)",
			version: "2.0.0-alpha+meta",
		},
		{
			spec:    "(1.0.0-beta+meta, 2.0.0-beta+meta)",
			version: "2.0.0-alpha",
		},
		{
			spec:    "(, 2.0.0-beta+meta)",
			version: "2.0.0-alpha",
		},
	}
	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			// Act
			versionInfo, err := ParseRange(tt.spec)
			require.NoError(t, err)
			middleVersion, err := Parse(tt.version)
			require.NoError(t, err)

			// Assert
			require.True(t, versionInfo.Satisfies(middleVersion))
		})
	}
}

func TestParseVersionRangeToString(t *testing.T) {
	tests := []struct {
		version  string
		expected string
	}{
		{
			version:  "1.2.0",
			expected: "[1.2.0, )",
		},
		{
			version:  "1.2.3-beta.2.4.55.X+900",
			expected: "[1.2.3-beta.2.4.55.X, )",
		},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			// Act
			versionInfo, err := ParseRange(tt.version)
			require.NoError(t, err)
			versionStr, err := versionInfo.String()
			require.NoError(t, err)

			// Assert
			require.Equal(t, tt.expected, versionStr)
		})
	}
}

func TestParseVersionRangeToStringShortHand(t *testing.T) {
	tests := []struct {
		version  string
		expected string
	}{
		{
			version:  "1.2.0",
			expected: "1.2.0",
		},
		{
			version:  "1.2.3",
			expected: "1.2.3",
		},
		{
			version:  "1.2.3-beta",
			expected: "1.2.3-beta",
		},
		{
			version:  "1.2.3-beta+900",
			expected: "1.2.3-beta",
		},
		{
			version:  "1.2.3-beta.2.4.55.X+900",
			expected: "1.2.3-beta.2.4.55.X",
		},
		{
			version:  "1.2.3-0+900",
			expected: "1.2.3-0",
		},
		{
			version:  "[1.2.0]",
			expected: "[1.2.0]",
		},
		{
			version:  "[1.2.3]",
			expected: "[1.2.3]",
		},
		{
			version:  "[1.2.3-beta]",
			expected: "[1.2.3-beta]",
		},
		{
			version:  "[1.2.3-beta+900]",
			expected: "[1.2.3-beta]",
		},
		{
			version:  "[1.2.3-beta.2.4.55.X+900]",
			expected: "[1.2.3-beta.2.4.55.X]",
		},
		{
			version:  "[1.2.3-0+90]",
			expected: "[1.2.3-0]",
		},
		{
			version:  "(, 1.2.0]",
			expected: "(, 1.2.0]",
		},
		{
			version:  "(, 1.2.3]",
			expected: "(, 1.2.3]",
		},
		{
			version:  "(, 1.2.3-beta]",
			expected: "(, 1.2.3-beta]",
		},
		{
			version:  "(, 1.2.3-beta+900]",
			expected: "(, 1.2.3-beta]",
		},
		{
			version:  "(, 1.2.3-beta.2.4.55.X+900]",
			expected: "(, 1.2.3-beta.2.4.55.X]",
		},
		{
			version:  "(, 1.2.3-0+900]",
			expected: "(, 1.2.3-0]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			// Act
			versionInfo, err := ParseRange(tt.version)
			require.NoError(t, err)

			actual, err := Format("S", *versionInfo)
			require.NoError(t, err)

			// Assert
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestStringFormatNullProvider(t *testing.T) {
	tests := []struct {
		versionRange string
	}{
		{
			versionRange: "1.2.0",
		},
		{
			versionRange: "1.2.3",
		},
		{
			versionRange: "1.2.3-beta",
		},
		{
			versionRange: "1.2.3-beta+900",
		},
		{
			versionRange: "1.2.3-beta.2.4.55.X+900",
		},
		{
			versionRange: "1.2.3-0+900",
		},
		{
			versionRange: "[1.2.0]",
		},
		{
			versionRange: "[1.2.3]",
		},
		{
			versionRange: "[1.2.3-beta]",
		},
		{
			versionRange: "[1.2.3-beta+900]",
		},
		{
			versionRange: "[1.2.3-beta.2.4.55.X+900]",
		},
		{
			versionRange: "[1.2.3-0+900]",
		},
		{
			versionRange: "(, 1.2.0)",
		},
		{
			versionRange: "(, 1.2.3)",
		},
		{
			versionRange: "(, 1.2.3-beta)",
		},
		{
			versionRange: "(, 1.2.3-beta+900)",
		},
		{
			versionRange: "(, 1.2.3-beta.2.4.55.X+900)",
		},
		{
			versionRange: "(, 1.2.3-0+900)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.versionRange, func(t *testing.T) {
			// Arrange
			versionInfo, err := ParseRange(tt.versionRange)
			require.NoError(t, err)

			actual, err := Format("", *versionInfo)
			require.NoError(t, err)

			expected, err := versionInfo.String()
			require.NoError(t, err)

			// Assert
			require.Equal(t, expected, actual)
		})
	}
}

func TestParseVersionRangeSimpleVersionNoBrackets(t *testing.T) {
	// Act
	versionInfo, err := ParseRange("1.2")
	require.NoError(t, err)
	maxVersionStr := versionInfo.MinVersion.Semver.String()

	// Assert
	require.Equal(t, "1.2.0", maxVersionStr)
	require.True(t, versionInfo.IsMinInclusive())
	require.Nil(t, versionInfo.MaxVersion)
	require.False(t, versionInfo.IsMaxInclusive())
}

func TestParseVersionRangeSimpleVersionNoBracketsExtraSpaces(t *testing.T) {
	// Act
	versionInfo, err := ParseRange("  1  .   2  ")
	require.NoError(t, err)
	minVersionStr := versionInfo.MinVersion.Semver.String()

	// Assert
	require.Equal(t, "1.2.0", minVersionStr)
	require.True(t, versionInfo.IsMinInclusive())
	require.Nil(t, versionInfo.MaxVersion)
	require.False(t, versionInfo.IsMaxInclusive())
}

func TestParseVersionRangeMaxOnlyInclusive(t *testing.T) {
	// Act
	versionInfo, err := ParseRange("(,1.2]")
	require.NoError(t, err)
	maxVersionStr := versionInfo.MaxVersion.Semver.String()

	// Assert
	require.Nil(t, versionInfo.MinVersion)
	require.False(t, versionInfo.IsMinInclusive())
	require.Equal(t, "1.2.0", maxVersionStr)
	require.True(t, versionInfo.IsMaxInclusive())
}
func TestParseVersionRangeMaxOnlyExclusive(t *testing.T) {
	// Act
	versionInfo, err := ParseRange("(,1.2)")
	require.NoError(t, err)
	maxVersionStr := versionInfo.MaxVersion.Semver.String()

	// Assert
	require.Nil(t, versionInfo.MinVersion)
	require.False(t, versionInfo.IsMinInclusive())
	require.Equal(t, "1.2.0", maxVersionStr)
	require.False(t, versionInfo.IsMaxInclusive())
}
func TestParseVersionRangeExactVersion(t *testing.T) {
	// Act
	versionInfo, err := ParseRange("[1.2]")
	require.NoError(t, err)
	maxVersionStr := versionInfo.MaxVersion.Semver.String()
	minVersionStr := versionInfo.MinVersion.Semver.String()

	// Assert
	require.Equal(t, "1.2.0", maxVersionStr)
	require.True(t, versionInfo.IsMinInclusive())
	require.Equal(t, "1.2.0", minVersionStr)
	require.True(t, versionInfo.IsMaxInclusive())
}

func TestParseVersionRangeMinOnlyExclusive(t *testing.T) {
	// Act
	versionInfo, err := ParseRange("(1.2,)")
	require.NoError(t, err)
	minVersionStr := versionInfo.MinVersion.Semver.String()

	// Assert
	require.Equal(t, "1.2.0", minVersionStr)
	require.False(t, versionInfo.IsMinInclusive())
	require.Nil(t, versionInfo.MaxVersion)
	require.False(t, versionInfo.IsMaxInclusive())
}
func TestParseVersionRangeExclusiveExclusive(t *testing.T) {
	// Act
	versionInfo, err := ParseRange("(1.2,2.3)")
	require.NoError(t, err)
	maxVersionStr := versionInfo.MaxVersion.Semver.String()
	minVersionStr := versionInfo.MinVersion.Semver.String()

	// Assert
	require.Equal(t, "1.2.0", minVersionStr)
	require.False(t, versionInfo.IsMinInclusive())
	require.Equal(t, "2.3.0", maxVersionStr)
	require.False(t, versionInfo.IsMaxInclusive())
}
func TestParseVersionRangeExclusiveInclusive(t *testing.T) {
	// Act
	versionInfo, err := ParseRange("(1.2,2.3]")
	require.NoError(t, err)
	maxVersionStr := versionInfo.MaxVersion.Semver.String()
	minVersionStr := versionInfo.MinVersion.Semver.String()

	// Assert
	require.Equal(t, "1.2.0", minVersionStr)
	require.False(t, versionInfo.IsMinInclusive())
	require.Equal(t, "2.3.0", maxVersionStr)
	require.True(t, versionInfo.IsMaxInclusive())
}
func TestParseVersionRangeInclusiveExclusive(t *testing.T) {
	// Act
	versionInfo, err := ParseRange("[1.2,2.3)")
	require.NoError(t, err)
	maxVersionStr := versionInfo.MaxVersion.Semver.String()
	minVersionStr := versionInfo.MinVersion.Semver.String()

	// Assert
	require.Equal(t, "1.2.0", minVersionStr)
	require.True(t, versionInfo.IsMinInclusive())
	require.Equal(t, "2.3.0", maxVersionStr)
	require.False(t, versionInfo.IsMaxInclusive())
}
func TestParseVersionRangeInclusiveInclusive(t *testing.T) {
	// Act
	versionInfo, err := ParseRange("[1.2,2.3]")
	require.NoError(t, err)
	maxVersionStr := versionInfo.MaxVersion.Semver.String()
	minVersionStr := versionInfo.MinVersion.Semver.String()

	// Assert
	require.Equal(t, "1.2.0", minVersionStr)
	require.True(t, versionInfo.IsMinInclusive())
	require.Equal(t, "2.3.0", maxVersionStr)
	require.True(t, versionInfo.IsMaxInclusive())
}
func TestParseVersionRangeInclusiveInclusiveExtraSpaces(t *testing.T) {
	// Act
	versionInfo, err := ParseRange("   [  1 .2   , 2  .3   ]  ")
	require.NoError(t, err)
	maxVersionStr := versionInfo.MaxVersion.Semver.String()
	minVersionStr := versionInfo.MinVersion.Semver.String()

	// Assert
	require.Equal(t, "1.2.0", minVersionStr)
	require.True(t, versionInfo.IsMinInclusive())
	require.Equal(t, "2.3.0", maxVersionStr)
	require.True(t, versionInfo.IsMaxInclusive())
}

func TestParsedVersionRangeHasOriginalString(t *testing.T) {
	tests := []struct {
		versionRange string
	}{
		{
			versionRange: "*",
		},
		{
			versionRange: "1.*",
		},
		{
			versionRange: "1.0.0",
		},
		{
			versionRange: " 1.0.0",
		},
		{
			versionRange: "[1.0.0]",
		},
		{
			versionRange: "[1.0.0] ",
		},
		{
			versionRange: "[1.0.0, 2.0.0)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.versionRange, func(t *testing.T) {
			// Act
			versionInfo, err := ParseRange(tt.versionRange)
			require.NoError(t, err)

			// Assert
			require.Equal(t, tt.versionRange, versionInfo.OriginalString)
		})
	}
}

func TestParseVersionToNormalizedVersion(t *testing.T) {
	// Act
	versionInfo, err := ParseRange("(1.0,1.2]")
	require.NoError(t, err)
	versionStr, err := versionInfo.String()
	require.NoError(t, err)

	// Assert
	require.Equal(t, "(1.0.0, 1.2.0]", versionStr)
}

func TestParseVersionParsesTokensVersionsCorrectly(t *testing.T) {
	tests := []struct {
		version      string
		minVersion   string
		minInclusive bool
		maxVersion   string
		maxInclusive bool
	}{
		{
			version:      "(1.2.3.4, 3.2)",
			minVersion:   "1.2.3.4",
			minInclusive: false,
			maxVersion:   "3.2",
			maxInclusive: false,
		},
		{
			version:      "(1.2.3.4, 3.2]",
			minVersion:   "1.2.3.4",
			minInclusive: false,
			maxVersion:   "3.2",
			maxInclusive: true,
		},
		{
			version:      "[1.2, 3.2.5)",
			minVersion:   "1.2",
			minInclusive: true,
			maxVersion:   "3.2.5",
			maxInclusive: false,
		},
		{
			version:      "[2.3.7, 3.2.4.5]",
			minVersion:   "2.3.7",
			minInclusive: true,
			maxVersion:   "3.2.4.5",
			maxInclusive: true,
		},
		{
			version:      "(, 3.2.4.5]",
			minVersion:   "",
			minInclusive: false,
			maxVersion:   "3.2.4.5",
			maxInclusive: true,
		},
		{
			version:      "(1.6, ]",
			minVersion:   "1.6",
			minInclusive: false,
			maxVersion:   "",
			maxInclusive: true,
		},
		{
			version:      "[2.7]",
			minVersion:   "2.7",
			minInclusive: true,
			maxVersion:   "2.7",
			maxInclusive: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			// Arrange
			var (
				minVersion *Version
				maxVersion *Version
				err        error
			)
			if strings.TrimSpace(tt.minVersion) != "" {
				minVersion, err = Parse(tt.minVersion)
				require.NoError(t, err)
			}
			if strings.TrimSpace(tt.maxVersion) != "" {
				maxVersion, err = Parse(tt.maxVersion)
				require.NoError(t, err)
			}
			versionRange, err := NewVersionRange(minVersion, maxVersion, tt.minInclusive, tt.maxInclusive, nil, "")
			require.NoError(t, err)

			// Act
			actual, err := ParseRange(tt.version)
			require.NoError(t, err)

			// Assert
			require.Equal(t, versionRange.IsMinInclusive(), actual.IsMinInclusive())
			require.Equal(t, versionRange.IsMaxInclusive(), actual.IsMaxInclusive())
			if versionRange.MinVersion != nil && actual.MinVersion != nil {
				if versionRange.MinVersion.Revision != actual.MinVersion.Revision ||
					versionRange.MinVersion.OriginalVersion != actual.MinVersion.OriginalVersion ||
					!versionRange.MinVersion.Semver.Equal(actual.MinVersion.Semver) {
					t.Errorf("min version Expected %+v, got %+v", versionRange.MinVersion, actual.MinVersion)
				}
			} else if versionRange.MinVersion != actual.MinVersion {
				t.Errorf("min version Expected %+v, got %+v", versionRange.MinVersion, actual.MinVersion)
			}

			if versionRange.MaxVersion != nil && actual.MaxVersion != nil {
				if versionRange.MaxVersion.Revision != actual.MaxVersion.Revision ||
					versionRange.MaxVersion.OriginalVersion != actual.MaxVersion.OriginalVersion ||
					!versionRange.MaxVersion.Semver.Equal(actual.MaxVersion.Semver) {
					t.Errorf("max version Expected %+v, got %+v", versionRange.MaxVersion, actual.MaxVersion)
				}
			} else if versionRange.MaxVersion != actual.MaxVersion {
				t.Errorf("max version Expected %+v, got %+v", versionRange.MaxVersion, actual.MaxVersion)
			}
		})
	}
}
func TestVersionRange_Equals(t *testing.T) {
	tests := []struct {
		versionString1 string
		versionString2 string
		expected       bool
	}{
		{
			versionString1: "1.0.0",
			versionString2: "1.0.*",
			expected:       false,
		},
		{
			versionString1: "[1.0.0,)",
			versionString2: "[1.0.*, )",
			expected:       false,
		},
		{
			versionString1: "1.1.*",
			versionString2: "1.0.*",
			expected:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.versionString1, func(t *testing.T) {
			// Act
			versionRange1, err := ParseRange(tt.versionString1)
			require.NoError(t, err)
			versionRange2, err := ParseRange(tt.versionString2)
			require.NoError(t, err)

			v1, err := versionRange1.String()
			require.NoError(t, err)
			v2, err := versionRange2.String()
			require.NoError(t, err)

			// Assert
			require.Equal(t, tt.expected, strings.EqualFold(v1, v2))
		})
	}
}
func TestVersionRange_ToStringRevPrefix(t *testing.T) {
	// Act
	versionRange, err := ParseRange("1.1.1.*-*")
	require.NoError(t, err)
	normalized, err := versionRange.ToNormalizedString()
	require.NoError(t, err)

	// Assert
	require.Equal(t, "[1.1.1.*-*, )", normalized)
}
func TestVersionRange_ToStringPatchPrefix(t *testing.T) {
	// Act
	versionRange, err := ParseRange("1.1.*-*")
	require.NoError(t, err)
	normalized, err := versionRange.ToNormalizedString()
	require.NoError(t, err)

	// Assert
	require.Equal(t, "[1.1.*-*, )", normalized)
}
func TestVersionRange_ToStringMinorPrefix(t *testing.T) {
	// Act
	versionRange, err := ParseRange("1.*-*")
	require.NoError(t, err)
	normalized, err := versionRange.ToNormalizedString()
	require.NoError(t, err)

	// Assert
	require.Equal(t, "[1.*-*, )", normalized)
}
func TestVersionRange_ToStringAbsoluteLatest(t *testing.T) {
	// Act
	versionRange, err := ParseRange("*-*")
	require.NoError(t, err)
	normalized, err := versionRange.ToNormalizedString()
	require.NoError(t, err)

	// Assert
	require.Equal(t, "[*-*, )", normalized)
	require.Equal(t, "0.0.0-0", versionRange.MinVersion.Semver.String())
	require.Equal(t, "0.0.0-0", versionRange.Float.MinVersion.Semver.String())
	require.Equal(t, AbsoluteLatest, versionRange.Float.FloatBehavior)
}
func TestVersionRange_ToStringPrereleaseMajor(t *testing.T) {
	// Act
	versionRange, err := ParseRange("*-rc.*")
	require.NoError(t, err)
	normalized, err := versionRange.ToNormalizedString()
	require.NoError(t, err)

	// Assert
	require.Equal(t, "[*-rc.*, )", normalized)
	require.Equal(t, "0.0.0-rc.0", versionRange.MinVersion.Semver.String())
	require.Equal(t, "0.0.0-rc.0", versionRange.Float.MinVersion.Semver.String())
	require.Equal(t, PrereleaseMajor, versionRange.Float.FloatBehavior)
}
