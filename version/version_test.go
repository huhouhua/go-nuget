// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package version

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionLength(t *testing.T) {
	tests := []struct {
		input   string
		wantErr error
	}{
		{
			input: "2",
		},
		{
			input: "2.0",
		},
		{
			input: "2.0.0",
		},
		{
			input: "2.0.0.0",
		},
		{
			input:   "",
			wantErr: errors.New("argument cannot be null or empty"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			semVer, err := Parse(tt.input)
			require.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}
			require.Equal(t, "2.0.0", semVer.Semver.String())
		})
	}
}

func TestFullVersionParsing(t *testing.T) {
	tests := []struct {
		input string
	}{
		{
			input: "1.0.0-Beta",
		},
		{
			input: "1.0.0-Beta.2",
		},
		{
			input: "1.0.0+MetaOnly",
		},
		{
			input: "1.0.0",
		},
		{
			input: "1.0.0-Beta+Meta",
		},
		{
			input: "1.0.0-RC.X+MetaAA",
		},
		{
			input: "1.0.0-RC.X.35.A.3455+Meta-A-B-C",
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Arrange & Act
			versions := parse(t, tt.input)

			// Assert
			for _, v := range versions {
				require.Equal(t, tt.input, v.Semver.String())
			}
		})
	}
}

func TestSpecialVersionParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "1.0.0-Beta",
			expected: "Beta",
		},
		{
			input:    "1.0.0-Beta+Meta",
			expected: "Beta",
		},
		{
			input:    "1.0.0-RC.X+Meta",
			expected: "RC.X",
		},
		{
			input:    "1.0.0-RC.X.35.A.3455+Meta",
			expected: "RC.X.35.A.3455",
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Arrange & Act
			versions := parse(t, tt.input)

			// Assert
			for _, v := range versions {
				require.Equal(t, tt.expected, v.Semver.Prerelease())
			}
		})
	}
}

func TestIsPrereleaseParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{
			input:    "1.0.0+Metadata",
			expected: false,
		},
		{
			input:    "1.0.0",
			expected: false,
		},
		{
			input:    "1.0.0-Beta",
			expected: true,
		},
		{
			input:    "1.0.0-Beta+Meta",
			expected: true,
		},
		{
			input:    "1.0.0-RC.X+Meta",
			expected: true,
		},
		{
			input:    "1.0.0-RC.X.35.A.3455+Meta",
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Arrange & Act
			versions := parse(t, tt.input)

			// Assert
			for _, v := range versions {
				if tt.expected {
					require.NotEmpty(t, v.Semver.Prerelease())
				} else {
					require.Empty(t, v.Semver.Prerelease())
				}
			}
		})
	}
}

func TestMetadataParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "1.0.0-Beta",
			expected: "",
		},
		{
			input:    "1.0.0-Beta+Meta",
			expected: "Meta",
		},
		{
			input:    "1.0.0-RC.X+MetaAA",
			expected: "MetaAA",
		},
		{
			input:    "1.0.0-RC.X.35.A.3455+Meta-A-B-C",
			expected: "Meta-A-B-C",
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Arrange & Act
			versions := parse(t, tt.input)

			// Assert
			for _, v := range versions {
				require.Equal(t, tt.expected, v.Semver.Metadata())
			}
		})
	}
}

func TestIsRevision(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{
			input:    "2018.4.8.256",
			expected: 256,
		},
		{
			input:    "2018-beta.256",
			expected: 0,
		},
		{
			input:    "1.0.0-Beta+Meta",
			expected: 0,
		},
		{
			input:    "1.0.0-RC.X+MetaAA",
			expected: 0,
		},
		{
			input:    "1.0.0-RC.X.35.A.3455+Meta-A-B-C",
			expected: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Arrange & Act
			versions := parse(t, tt.input)

			// Assert
			for _, v := range versions {
				require.Equal(t, tt.expected, v.Revision)
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		major   uint64
		minor   uint64
		patch   uint64
		version string
	}{
		{
			major:   0,
			minor:   0,
			patch:   0,
			version: "0.0.0",
		},
		{
			major:   1,
			minor:   0,
			patch:   0,
			version: "1.0.0",
		},
		{
			major:   3,
			minor:   5,
			patch:   1,
			version: "3.5.1",
		},
		{
			major:   234,
			minor:   234234,
			patch:   1111,
			version: "234.234234.1111",
		},
		{
			major:   3,
			minor:   5,
			patch:   1,
			version: "3.5.1+Meta",
		},
		{
			major:   3,
			minor:   5,
			patch:   1,
			version: "3.5.1-x.y.z+AA",
		},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			// Arrange & Act
			versions := parse(t, tt.version)

			// Assert
			for _, v := range versions {
				require.Equal(t, tt.major, v.Semver.Major())
				require.Equal(t, tt.minor, v.Semver.Minor())
				require.Equal(t, tt.patch, v.Semver.Patch())
			}
		})
	}
}

func TestTryGetNormalizedVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected *SemanticVersion
		ok       bool
	}{
		{"1.2.3.4", &SemanticVersion{1, 2, 3, 4}, true},
		{"1.2.3", &SemanticVersion{1, 2, 3, 0}, true},
		{"1.2", &SemanticVersion{1, 2, 0, 0}, true},
		{"1", &SemanticVersion{1, 0, 0, 0}, true},
		{"", nil, false},
		{" 1.2.3.4 ", &SemanticVersion{1, 2, 3, 4}, true},
		{"1.2.3.4.5", nil, false},
		{"a.b.c.d", nil, false},
		{"1.2.3.x", nil, false},
		{"1. 2.3.4", &SemanticVersion{1, 2, 3, 4}, true},
		{"0.0.0.0", &SemanticVersion{0, 0, 0, 0}, true},
		{"2147483647.0.0.0", &SemanticVersion{2147483647, 0, 0, 0}, true},
		{"1.-2.3.4", nil, false},
		{"1..2.3.4", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ver, ok := tryGetNormalizedVersion(tt.input)
			if tt.ok {
				require.True(t, ok)
				require.Equal(t, tt.expected, ver)
			} else {
				require.False(t, ok)
			}
		})
	}
}

func TestParseSection(t *testing.T) {
	tests := []struct {
		input   string
		start   int
		wantEnd int
		wantNum int
		wantOk  bool
	}{
		{"123.456", 0, 4, 123, true},
		{"  42.7", 0, 5, 42, true},
		{"0.1", 0, 2, 0, true},
		{"9999999999", 0, 9, 0, false},
		{"abc", 0, 0, 0, false},
		{"", 0, 0, 0, true},
		{"  ", 0, 2, 0, false},
		{"1.2.3", 2, 4, 2, true},
		{"1.2.3", 4, 5, 3, true},
		{"123.", 0, 4, 0, false},
		{"12a.3", 0, 2, 0, false},
		{"1..2", 0, 2, 1, true},
		{".", 0, 0, 0, false},
		{"-1.2", 0, 0, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			end, num, ok := parseSection(tt.input, tt.start)
			require.Equal(t, tt.wantEnd, end)
			require.Equal(t, tt.wantNum, num)
			require.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestConvertVersion(t *testing.T) {
	ver := &SemanticVersion{1, 2, 3, 4}
	v, err := ConvertVersion(ver, "1.2.3.4", "meta", []string{"beta", "rc"})
	require.NoError(t, err)
	require.Equal(t, uint64(1), v.Semver.Major())
	require.Equal(t, uint64(2), v.Semver.Minor())
	require.Equal(t, uint64(3), v.Semver.Patch())
	require.Equal(t, "beta.rc", v.Semver.Prerelease())
	require.Equal(t, "meta", v.Semver.Metadata())
	require.Equal(t, 4, v.Revision)

	_, err = ConvertVersion(nil, "1.2.3.4", "meta", []string{"beta"})
	require.Error(t, err)

	verNeg := &SemanticVersion{1, 2, -3, -4}
	v2, err := ConvertVersion(verNeg, "1.2.-3.-4", "", nil)
	require.NoError(t, err)
	require.Equal(t, uint64(0), v2.Semver.Patch())
	require.Equal(t, 0, v2.Revision)

	v3, err := ConvertVersion(&SemanticVersion{1, 2, 3, 4}, "1.2.3.4", "", nil)
	require.NoError(t, err)
	require.Equal(t, "", v3.Semver.Metadata())
	require.Equal(t, "", v3.Semver.Prerelease())

	v4, err := ConvertVersion(&SemanticVersion{1, 2, 3, 4}, "1.2.3.4", "meta-!@#", []string{"beta-!@#"})
	require.NoError(t, err)
	require.Contains(t, v4.Semver.Metadata(), "meta-!@#")
	require.Contains(t, v4.Semver.Prerelease(), "beta-!@#")
}

func TestParseSections(t *testing.T) {
	tests := []struct {
		input        string
		wantVersion  string
		wantLabels   []string
		wantMetadata string
	}{
		{"1.2.3", "1.2.3", nil, ""},
		{"1.2.3-beta", "1.2.3", []string{"beta"}, ""},
		{"1.2.3-beta.1", "1.2.3", []string{"beta", "1"}, ""},
		{"1.2.3+meta", "1.2.3", nil, "meta"},
		{"1.2.3-beta+meta", "1.2.3", []string{"beta"}, "meta"},
		{"1.2.3-beta.1+meta.2", "1.2.3", []string{"beta", "1"}, "meta.2"},
		{"1.2.3-rc.x.35.a.3455+meta-a-b-c", "1.2.3", []string{"rc", "x", "35", "a", "3455"}, "meta-a-b-c"},
		{"1.2.3+", "1.2.3+", nil, ""},
		{"1.2.3-", "1.2.3-", nil, ""},
		{"1.2.3-+meta", "1.2.3", []string{""}, "meta"},
		{"1.2.3-beta+", "1.2.3", []string{"beta+"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ver, labels, meta := parseSections(tt.input)
			require.Equal(t, tt.wantVersion, ver)
			require.Equal(t, tt.wantLabels, labels)
			require.Equal(t, tt.wantMetadata, meta)
		})
	}
}

func TestIsValidPart(t *testing.T) {
	tests := []struct {
		input             string
		allowLeadingZeros bool
		want              bool
	}{
		{"0", false, true},
		{"01", false, false},
		{"01", true, true},
		{"0A", false, true},
		{"A0", false, true},
		{"A-0", false, true},
		{"", false, false},
		{"abc", false, true},
		{"aBc-123", false, true},
		{"0123", false, false},
		{"0123", true, true},
		{"123", false, true},
		{"-abc", false, true},
		{"abc!", false, false},
		{"---", false, true},
		{"1-2-3", false, true},
		{"ABCDEFGHIJKLMNOPQRSTUVWXYZ", false, true},
		{"abcdefghijklmnopqrstuvwxyz", false, true},
		{"", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isValidPart(tt.input, tt.allowLeadingZeros)
			require.Equal(t, tt.want, got)
		})
	}
}

// parse is All possible ways to parse a version from a string
func parse(t *testing.T, version string) []*Version {
	// Parse
	versions := make([]*Version, 0)
	v, err := Parse(version)
	require.NoError(t, err)
	versions = append(versions, v)

	vCache, err := Parse(version)
	require.NoError(t, err)
	versions = append(versions, vCache)

	// TryParse
	ok, semVer, err := TryParse(version)
	require.True(t, ok)
	require.NoError(t, err)
	versions = append(versions, semVer)

	ok, semVerCache, err := TryParse(version)
	require.True(t, ok)
	require.NoError(t, err)
	versions = append(versions, semVerCache)

	// Parse Strict
	versions = append(versions, NewVersion(semVerCache.Semver, semVerCache.Revision, semVerCache.OriginalVersion))
	return versions
}
