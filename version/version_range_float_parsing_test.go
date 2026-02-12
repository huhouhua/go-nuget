// Copyright (c) 2025 Kevin Berger <huhouhuam@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package version

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionRangeFloatParsing_Prerelease(t *testing.T) {
	// Act
	versionRange, err := ParseRange("1.0.0-*")
	require.NoError(t, err)
	require.NotEmpty(t, versionRange.MinVersion.Semver.Prerelease())
}

func TestVersionRangeFloatParsing_PrereleaseWithNumericOnlyLabelVerifyMinVersion(t *testing.T) {
	tests := []struct {
		rangeString string
		expected    string
	}{
		{
			rangeString: "1.0.0-*",
			expected:    "1.0.0-0",
		},
		{
			rangeString: "1.0.0-0*",
			expected:    "1.0.0-0",
		},
		{
			rangeString: "1.0.0--*",
			expected:    "1.0.0--",
		},
		{
			rangeString: "1.0.0-a-*",
			expected:    "1.0.0-a-",
		},
		{
			rangeString: "1.0.0-a.*",
			expected:    "1.0.0-a.0",
		},
		{
			rangeString: "1.*-*",
			expected:    "1.0.0-0",
		},
		{
			rangeString: "1.0.*-0*",
			expected:    "1.0.0-0",
		},
		{
			rangeString: "1.0.*--*",
			expected:    "1.0.0--",
		},
		{
			rangeString: "1.0.*-a-*",
			expected:    "1.0.0-a-",
		},
		{
			rangeString: "1.0.*-a.*",
			expected:    "1.0.0-a.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.rangeString, func(t *testing.T) {
			// Act
			versionRange, err := ParseRange(tt.rangeString)
			require.NoError(t, err)
			require.Equal(t, tt.expected, versionRange.MinVersion.Semver.String())
		})
	}
}

func TestVersionRangeFloatParsing_PrereleaseWithNumericOnlyLabelVerifySatisfies(t *testing.T) {
	tests := []struct {
		version string
	}{
		{
			version: "1.0.0-0",
		},
		{
			version: "1.0.0-100",
		},
		{
			version: "1.0.0-0.0.0.0",
		},
		{
			version: "1.0.0-0+0-0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			// Act
			versionRange, err := ParseRange("1.0.0-*")
			require.NoError(t, err)

			v, err := Parse(tt.version)
			require.NoError(t, err)

			require.True(t, versionRange.Satisfies(v))
		})
	}
}

func TestVersionRangeFloatParsing_VerifySatisfiesForFloatingRange(t *testing.T) {
	tests := []struct {
		rangeString string
		version     string
	}{
		{
			rangeString: "1.0.0-a*",
			version:     "1.0.0-a.0",
		},
		{
			rangeString: "1.0.0-a*",
			version:     "1.0.0-a-0",
		},
		{
			rangeString: "1.0.0-a*",
			version:     "1.0.0-a",
		},
		{
			rangeString: "1.0.*-a*",
			version:     "1.0.0-a",
		},
		{
			rangeString: "1.*-a*",
			version:     "1.0.0-a",
		},
		{
			rangeString: "*-a*",
			version:     "1.0.0-a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.rangeString, func(t *testing.T) {
			// Act
			versionRange, err := ParseRange(tt.rangeString)
			require.NoError(t, err)

			v, err := Parse(tt.version)
			require.NoError(t, err)

			require.True(t, versionRange.Satisfies(v))
		})
	}
}

func TestVersionRangeFloatParsing_VerifyReleaseLabels(t *testing.T) {
	tests := []struct {
		rangeString   string
		versionLabel  string
		originalLabel string
	}{
		{
			rangeString:   "1.0.0-*",
			versionLabel:  "0",
			originalLabel: "",
		},
		{
			rangeString:   "1.0.0-a*",
			versionLabel:  "a",
			originalLabel: "a",
		},
		{
			rangeString:   "1.0.0-a-*",
			versionLabel:  "a-",
			originalLabel: "a-",
		},
		{
			rangeString:   "1.0.0-a.*",
			versionLabel:  "a.0",
			originalLabel: "a.",
		},
		{
			rangeString:   "1.0.0-0*",
			versionLabel:  "0",
			originalLabel: "0",
		},
		{
			rangeString:   "1.0.*-0*",
			versionLabel:  "0",
			originalLabel: "0",
		},
		{
			rangeString:   "1.*-0*",
			versionLabel:  "0",
			originalLabel: "0",
		},
		{
			rangeString:   "*-0*",
			versionLabel:  "0",
			originalLabel: "0",
		},
		{
			rangeString:   "1.0.*-*",
			versionLabel:  "0",
			originalLabel: "",
		},
		{
			rangeString:   "1.*-*",
			versionLabel:  "0",
			originalLabel: "",
		},
		{
			rangeString:   "*-*",
			versionLabel:  "0",
			originalLabel: "",
		},
		{
			rangeString:   "1.0.*-a*",
			versionLabel:  "a",
			originalLabel: "a",
		},
		{
			rangeString:   "1.*-a*",
			versionLabel:  "a",
			originalLabel: "a",
		},
		{
			rangeString:   "*-a*",
			versionLabel:  "a",
			originalLabel: "a",
		},
		{
			rangeString:   "1.0.*-a-*",
			versionLabel:  "a-",
			originalLabel: "a-",
		},
		{
			rangeString:   "1.*-a-*",
			versionLabel:  "a-",
			originalLabel: "a-",
		},
		{
			rangeString:   "*-a-*",
			versionLabel:  "a-",
			originalLabel: "a-",
		},
		{
			rangeString:   "1.0.*-a.*",
			versionLabel:  "a.0",
			originalLabel: "a.",
		},
		{
			rangeString:   "1.*-a.*",
			versionLabel:  "a.0",
			originalLabel: "a.",
		},
		{
			rangeString:   "*-a.*",
			versionLabel:  "a.0",
			originalLabel: "a.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.rangeString, func(t *testing.T) {
			// Act
			versionRange, err := ParseRange(tt.rangeString)
			require.NoError(t, err)

			require.Equal(t, tt.versionLabel, versionRange.Float.MinVersion.Semver.Prerelease())
			require.Equal(t, tt.originalLabel, versionRange.Float.OriginalReleasePrefix)
		})
	}
}

func TestVersionRangeFloatParsing_Invalid(t *testing.T) {
	tests := []struct {
		rangeString string
	}{
		{rangeString: "[]"},
		{rangeString: "[*]"},
		{rangeString: "[1.0.0, 1.1.*)"},
		{rangeString: "[1.0.0, 2.0.*)"},
		{rangeString: "(, 2.*.*)"},
		{rangeString: "<1.0.*"},
		{rangeString: "<=1.0.*"},
		{rangeString: "1.0.0<"},
		{rangeString: "1.0.0~"},
		{rangeString: "~1.*.*"},
		{rangeString: "~*"},
		{rangeString: "~"},
		{rangeString: "^"},
		{rangeString: "^*"},
		{rangeString: ">=*"},
		{rangeString: "1.*.0"},
		{rangeString: "1.*.0-beta-*"},
		{rangeString: "1.*.0-beta"},
		{rangeString: "1.0.0.0.*"},
		{rangeString: "=1.0.*"},
		{rangeString: "1.0.0+*"},
		{rangeString: "1.0.**"},
		{rangeString: "1.0.*-*bla"},
		{rangeString: "1.0.*-*bla+*"},
		{rangeString: "**"},
		{rangeString: "1.0.0-preview.*+blabla"},
		{rangeString: "1.0.*--"},
		{rangeString: "1.0.*-alpha*+"},
		{rangeString: "1.0.*-"},
	}
	for _, tt := range tests {
		t.Run(tt.rangeString, func(t *testing.T) {
			// Act
			versionRange, ok := TryParseRange(tt.rangeString, true)
			require.False(t, ok)
			require.Nil(t, versionRange)
		})
	}
}

func TestVersionRangeFloatParsing_Valid(t *testing.T) {
	tests := []struct {
		rangeString string
	}{
		{rangeString: "*"},
		{rangeString: "1.*"},
		{rangeString: "1.0.*"},
		{rangeString: "1.0.0.*"},
		{rangeString: "1.0.0.0-beta"},
		{rangeString: "1.0.0.0-beta*"},
		{rangeString: "1.0.0"},
		{rangeString: "1.0"},
		{rangeString: "[1.0.*, )"},
		{rangeString: "[1.0.0-beta.*, 2.0.0)"},
		{rangeString: "1.0.0-beta.*"},
		{rangeString: "1.0.0-beta-*"},
		{rangeString: "1.0.*-bla*"},
		{rangeString: "1.0.*-*"},
		{rangeString: "1.0.*-preview.1.*"},
		{rangeString: "1.0.*-preview.1*"},
		{rangeString: "1.0.0--"},
		{rangeString: "1.0.0-bla*"},
		{rangeString: "1.0.*--*"},
		{rangeString: "1.0.0--*"},
	}
	for _, tt := range tests {
		t.Run(tt.rangeString, func(t *testing.T) {
			// Act
			versionRange, ok := TryParseRange(tt.rangeString, true)
			require.True(t, ok)
			require.NotNil(t, versionRange)
		})
	}
}

func TestVersionRangeFloatParsing_LegacyEquivalent(t *testing.T) {
	tests := []struct {
		rangeString  string
		legacyString string
	}{
		{rangeString: "1.0.0", legacyString: "[1.0.0, )"},
		{rangeString: "1.0.*", legacyString: "[1.0.0, )"},
		{rangeString: "[1.0.*, )", legacyString: "[1.0.0, )"},
		{rangeString: "[1.*, )", legacyString: "[1.0.0, )"},
		{rangeString: "[1.*, 2.0)", legacyString: "[1.0.0, 2.0.0)"},
		{rangeString: "*", legacyString: "[0.0.0, )"},
	}

	for _, tt := range tests {
		t.Run(tt.rangeString, func(t *testing.T) {
			// Act
			versionRange, ok := TryParseRange(tt.rangeString, true)
			require.True(t, ok)

			legacyStr, err := versionRange.ToLegacyString()
			require.NoError(t, err)

			require.Equal(t, tt.legacyString, legacyStr)
		})
	}
}

func TestVersionRangeFloatParsing_CorrectFloatRange(t *testing.T) {
	tests := []struct {
		rangeString string
	}{
		{rangeString: "1.0.0-beta*"},
		{rangeString: "1.0.0-beta.*"},
		{rangeString: "1.0.0-beta-*"},
	}
	for _, tt := range tests {
		t.Run(tt.rangeString, func(t *testing.T) {
			// Act
			versionRange, ok := TryParseRange(tt.rangeString, true)
			require.True(t, ok)
			require.Equal(t, tt.rangeString, versionRange.Float.String())
		})
	}
}
