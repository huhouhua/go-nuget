// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"fmt"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/require"
)

func TestNuGetVersion(t *testing.T) {
	v := &NuGetVersion{
		semver.New(1, 0, 0, "beta", ""),
	}
	require.True(t, true, v.IsSemVer2())
	require.True(t, true, v.IsPrerelease())
}

func TestParseVersionRange(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *VersionRange
		wantErr bool
	}{
		{
			name:  "exact version",
			input: "1.0.0",
			want: &VersionRange{
				MinVersion: &NuGetVersion{semver.MustParse("1.0.0")},
				MaxVersion: &NuGetVersion{semver.MustParse("1.0.0")},
				IncludeMin: true,
				IncludeMax: true,
			},
		},
		{
			name:  "exact version with prerelease",
			input: "1.0.0-beta.1",
			want: &VersionRange{
				MinVersion: &NuGetVersion{semver.MustParse("1.0.0-beta.1")},
				MaxVersion: &NuGetVersion{semver.MustParse("1.0.0-beta.1")},
				IncludeMin: true,
				IncludeMax: true,
			},
		},
		{
			name:  "exact version with build metadata",
			input: "1.0.0+build.1",
			want: &VersionRange{
				MinVersion: &NuGetVersion{semver.MustParse("1.0.0+build.1")},
				MaxVersion: &NuGetVersion{semver.MustParse("1.0.0+build.1")},
				IncludeMin: true,
				IncludeMax: true,
			},
		},
		{
			name:  "inclusive range",
			input: "[1.0.0, 2.0.0]",
			want: &VersionRange{
				MinVersion: &NuGetVersion{semver.MustParse("1.0.0")},
				MaxVersion: &NuGetVersion{semver.MustParse("2.0.0")},
				IncludeMin: true,
				IncludeMax: true,
			},
		},
		{
			name:  "exclusive range",
			input: "(1.0.0, 2.0.0)",
			want: &VersionRange{
				MinVersion: &NuGetVersion{semver.MustParse("1.0.0")},
				MaxVersion: &NuGetVersion{semver.MustParse("2.0.0")},
				IncludeMin: false,
				IncludeMax: false,
			},
		},
		{
			name:  "mixed range",
			input: "(1.0.0, 2.0.0]",
			want: &VersionRange{
				MinVersion: &NuGetVersion{semver.MustParse("1.0.0")},
				MaxVersion: &NuGetVersion{semver.MustParse("2.0.0")},
				IncludeMin: false,
				IncludeMax: true,
			},
		},
		{
			name:  "min only",
			input: "[1.0.0,)",
			want: &VersionRange{
				MinVersion: &NuGetVersion{semver.MustParse("1.0.0")},
				MaxVersion: nil,
				IncludeMin: true,
				IncludeMax: false,
			},
		},
		{
			name:  "max only",
			input: "(,2.0.0]",
			want: &VersionRange{
				MinVersion: nil,
				MaxVersion: &NuGetVersion{semver.MustParse("2.0.0")},
				IncludeMin: false,
				IncludeMax: true,
			},
		},
		{
			name:  "wildcard",
			input: "*",
			want: &VersionRange{
				Float: Major,
			},
		},
		{
			name:  "tilde range",
			input: "~1.2.3",
			want: &VersionRange{
				MinVersion: &NuGetVersion{semver.MustParse("1.2.3")},
				MaxVersion: &NuGetVersion{semver.MustParse("1.3.0")},
				IncludeMin: true,
				IncludeMax: false,
				Float:      Patch,
			},
		},
		{
			name:  "tilde range with prerelease",
			input: "~1.2.3-beta.1",
			want: &VersionRange{
				MinVersion: &NuGetVersion{semver.MustParse("1.2.3-beta.1")},
				MaxVersion: &NuGetVersion{semver.MustParse("1.3.0")},
				IncludeMin: true,
				IncludeMax: false,
				Float:      Patch,
			},
		},
		{
			name:  "caret range",
			input: "^1.2.3",
			want: &VersionRange{
				MinVersion: &NuGetVersion{semver.MustParse("1.2.3")},
				MaxVersion: &NuGetVersion{semver.MustParse("2.0.0")},
				IncludeMin: true,
				IncludeMax: false,
				Float:      Minor,
			},
		},
		{
			name:  "caret range with prerelease",
			input: "^1.2.3-beta.1",
			want: &VersionRange{
				MinVersion: &NuGetVersion{semver.MustParse("1.2.3-beta.1")},
				MaxVersion: &NuGetVersion{semver.MustParse("2.0.0")},
				IncludeMin: true,
				IncludeMax: false,
				Float:      Minor,
			},
		},
		{
			name:  "caret range pre-1.0",
			input: "^0.2.3",
			want: &VersionRange{
				MinVersion: &NuGetVersion{semver.MustParse("0.2.3")},
				MaxVersion: &NuGetVersion{semver.MustParse("0.3.0")},
				IncludeMin: true,
				IncludeMax: false,
				Float:      Minor,
			},
		},
		{
			name:  "symbol range",
			input: "*-",
			want: &VersionRange{
				Float: Prerelease,
			},
		},
		{
			name:    "parse symbol range error",
			input:   "-*[1.0.0]",
			wantErr: true,
		},
		{
			name:    "parse prefix symbol error",
			input:   "~1.0.0*",
			wantErr: true,
		},
		{
			name:    "unsupported prefix symbol error",
			input:   "*1.0.0",
			wantErr: true,
		},
		{
			name:    "invalid version",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "invalid range format",
			input:   "[1.0.0]",
			wantErr: true,
		},
		{
			name:    "invalid range format with extra comma",
			input:   "[1.0.0,,2.0.0]",
			wantErr: true,
		},
		{
			name:    "empty range",
			input:   "",
			wantErr: true,
		},
		{
			name:    "parse min version error",
			input:   "[~1.0.0, 2.0.0]",
			wantErr: true,
		},
		{
			name:    "parse max version error",
			input:   "[1.0.0, ~2.0.0]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersionRange(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseFloatingRange(t *testing.T) {
	_, actualErr := parseFloatingRange("*")
	expectedErr := fmt.Errorf("invalid floating range format: *")
	require.Equal(t, expectedErr, actualErr)

	_, actualErr = parseFloatingRange("^*")
	expectedErr = fmt.Errorf("invalid version in ^ range: Invalid Semantic Version")
	require.Equal(t, expectedErr, actualErr)
}

func TestNewVersionRange(t *testing.T) {
	minVersion := semver.New(1, 0, 0, "", "")
	minVersionPre := semver.New(1, 0, 0, "beta", "")

	maxVersion := semver.New(2, 0, 0, "", "")
	maxVersionPre := semver.New(2, 0, 0, "beta", "")

	minMinorVersion := semver.New(1, 2, 3, "", "")
	maxMinorVersion := semver.New(1, 5, 1, "", "")

	minMinorBetaVersion := semver.New(1, 2, 3, "beta", "")
	maxMinorBetaVersion := semver.New(1, 5, 1, "beta", "")

	rangeVersion := NewVersionRange(minVersion, maxVersion, true, true)
	require.NotNil(t, rangeVersion)

	require.False(t, rangeVersion.satisfiesFloat(maxVersion))
	require.Empty(t, rangeVersion.stringFloat())

	rangeVersion.Float = Prerelease
	require.Equal(t, "*-", rangeVersion.stringFloat())
	require.False(t, rangeVersion.IsBetter(minVersion, minVersion))
	require.False(t, rangeVersion.IsBetter(minVersion, nil))
	require.False(t, rangeVersion.IsBetter(minVersion, maxVersionPre))

	rangeVersion.Float = None
	require.False(t, rangeVersion.IsBetter(nil, minVersionPre))

	rangeMinorVersion := NewVersionRange(maxMinorVersion, minMinorVersion, true, true)
	require.NotNil(t, rangeMinorVersion)

	rangeMinorVersion.Float = Prerelease
	require.True(t, rangeMinorVersion.IsBetter(minMinorVersion, maxVersionPre))

	rangeMinorBetaVersion := NewVersionRange(maxMinorBetaVersion, minMinorBetaVersion, true, true)
	require.NotNil(t, rangeMinorBetaVersion)

	rangeMinorBetaVersion.Float = Prerelease
	require.False(t, rangeMinorBetaVersion.IsBetter(maxVersion, minMinorBetaVersion))

	rangeMinorBetaVersion.Float = Patch
	actualRange := rangeMinorBetaVersion.ToNonSnapshotRange()
	wantRange := NewVersionRange(maxMinorBetaVersion, semver.New(1, 6, 0, "", ""), true, false)
	require.Equal(t, wantRange, actualRange)

	betaRange := NewVersionRange(minMinorBetaVersion, maxMinorBetaVersion, true, false)
	betaRange.Float = Prerelease
	require.Equal(t, "Latest prerelease version >= 1.2.3-beta", betaRange.PrettyPrint())

	betaRange.MinVersion = nil
	betaRange.MaxVersion = nil
	require.Equal(t, "Latest prerelease version", betaRange.PrettyPrint())
	betaRange.Float = Minor
	require.Equal(t, "Latest minor version", betaRange.PrettyPrint())
	betaRange.Float = Patch
	require.Equal(t, "Latest patch version", betaRange.PrettyPrint())
	betaRange.Float = None
	require.Equal(t, "Any version", betaRange.PrettyPrint())
}

func TestVersionRange_Satisfies(t *testing.T) {
	tests := []struct {
		name         string
		rangeVersion string
		version      string
		want         bool
	}{
		{
			name:         "exact version match",
			rangeVersion: "1.0.0",
			version:      "1.0.0",
			want:         true,
		},
		{
			name:         "exact version with prerelease match",
			rangeVersion: "1.0.0-beta.1",
			version:      "1.0.0-beta.1",
			want:         true,
		},
		{
			name:         "exact version with build metadata match",
			rangeVersion: "1.0.0+build.1",
			version:      "1.0.0+build.2",
			want:         true,
		},
		{
			name:         "exact version mismatch",
			rangeVersion: "1.0.0",
			version:      "1.0.1",
			want:         false,
		},
		{
			name:         "inclusive range within",
			rangeVersion: "[1.0.0, 2.0.0]",
			version:      "1.5.0",
			want:         true,
		},
		{
			name:         "inclusive range at min",
			rangeVersion: "[1.0.0, 2.0.0]",
			version:      "1.0.0",
			want:         true,
		},
		{
			name:         "inclusive range at max",
			rangeVersion: "[1.0.0, 2.0.0]",
			version:      "2.0.0",
			want:         true,
		},
		{
			name:         "exclusive range within",
			rangeVersion: "(1.0.0, 2.0.0)",
			version:      "1.5.0",
			want:         true,
		},
		{
			name:         "exclusive range at min",
			rangeVersion: "(1.0.0, 2.0.0)",
			version:      "1.0.0",
			want:         false,
		},
		{
			name:         "exclusive range at max",
			rangeVersion: "(1.0.0, 2.0.0)",
			version:      "2.0.0",
			want:         false,
		},
		{
			name:         "wildcard any version",
			rangeVersion: "*",
			version:      "1.0.0",
			want:         true,
		},
		{
			name:         "wildcard with prerelease",
			rangeVersion: "*",
			version:      "1.0.0-beta.1",
			want:         true,
		},
		{
			name:         "tilde range within",
			rangeVersion: "~1.2.3",
			version:      "1.2.5",
			want:         true,
		},
		{
			name:         "tilde range at min",
			rangeVersion: "~1.2.3",
			version:      "1.2.3",
			want:         true,
		},
		{
			name:         "tilde range at max",
			rangeVersion: "~1.2.3",
			version:      "1.3.0",
			want:         false,
		},
		{
			name:         "tilde range with prerelease within",
			rangeVersion: "~1.2.3-beta.1",
			version:      "1.2.5",
			want:         true,
		},
		{
			name:         "caret range within",
			rangeVersion: "^1.2.3",
			version:      "1.5.0",
			want:         true,
		},
		{
			name:         "caret range at min",
			rangeVersion: "^1.2.3",
			version:      "1.2.3",
			want:         true,
		},
		{
			name:         "caret range at max",
			rangeVersion: "^1.2.3",
			version:      "2.0.0",
			want:         false,
		},
		{
			name:         "caret range pre-1.0 within",
			rangeVersion: "^0.2.3",
			version:      "0.2.5",
			want:         true,
		},
		{
			name:         "caret range pre-1.0 at max",
			rangeVersion: "^0.2.3",
			version:      "0.3.0",
			want:         false,
		},
		{
			name:         "caret range with prerelease within",
			rangeVersion: "^1.2.3-beta.1",
			version:      "1.5.0",
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr, err := ParseVersionRange(tt.rangeVersion)
			require.NoError(t, err)
			v := semver.MustParse(tt.version)
			got := vr.Satisfies(v)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestVersionRange_String(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "exact version",
			input: "1.0.0",
			want:  "1.0.0",
		},
		{
			name:  "exact version with prerelease",
			input: "1.0.0-beta.1",
			want:  "1.0.0-beta.1",
		},
		{
			name:  "exact version with build metadata",
			input: "1.0.0+build.1",
			want:  "1.0.0+build.1",
		},
		{
			name:  "inclusive range",
			input: "[1.0.0,2.0.0]",
			want:  "[1.0.0,2.0.0]",
		},
		{
			name:  "exclusive range",
			input: "(1.0.0,2.0.0)",
			want:  "(1.0.0,2.0.0)",
		},
		{
			name:  "mixed range",
			input: "(1.0.0,2.0.0]",
			want:  "(1.0.0,2.0.0]",
		},
		{
			name:  "min only",
			input: "[1.0.0,)",
			want:  "[1.0.0,)",
		},
		{
			name:  "max only",
			input: "(,2.0.0]",
			want:  "(,2.0.0]",
		},
		{
			name:  "wildcard",
			input: "*",
			want:  "*",
		},
		{
			name:  "tilde range",
			input: "~1.2.3",
			want:  "~1.2.3",
		},
		{
			name:  "tilde range with prerelease",
			input: "~1.2.3-beta.1",
			want:  "~1.2.3-beta.1",
		},
		{
			name:  "caret range",
			input: "^1.2.3",
			want:  "^1.2.3",
		},
		{
			name:  "caret range with prerelease",
			input: "^1.2.3-beta.1",
			want:  "^1.2.3-beta.1",
		},
		{
			name:  "caret range pre-1.0",
			input: "^0.2.3",
			want:  "^0.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr, err := ParseVersionRange(tt.input)
			require.NoError(t, err)
			got := vr.String()
			require.Equal(t, tt.want, got)
		})
	}
}
func TestVersionRange_DoesRangeSatisfy(t *testing.T) {
	tests := []struct {
		name     string
		range1   string
		range2   string
		want     bool
		wantErr1 bool
		wantErr2 bool
	}{
		{
			name:   "overlapping ranges",
			range1: "[1.0.0, 2.0.0]",
			range2: "[1.5.0, 2.5.0]",
			want:   true,
		},
		{
			name:   "non-overlapping ranges",
			range1: "[1.0.0, 2.0.0]",
			range2: "[2.1.0, 3.0.0]",
			want:   false,
		},
		{
			name:   "exact version in range",
			range1: "[1.0.0, 2.0.0]",
			range2: "1.5.0",
			want:   true,
		},
		{
			name:   "exact version not in range",
			range1: "[1.0.0, 2.0.0]",
			range2: "2.1.0",
			want:   false,
		},
		{
			name:   "wildcard range",
			range1: "*",
			range2: "[1.0.0, 2.0.0]",
			want:   true,
		},
		{
			name:   "tilde range with overlapping",
			range1: "~1.2.3",
			range2: "[1.2.0, 1.3.0]",
			want:   true,
		},
		{
			name:   "caret range with overlapping",
			range1: "^1.2.3",
			range2: "[1.0.0, 2.0.0]",
			want:   true,
		},
		{
			name:   "pre-1.0 caret range",
			range1: "^0.2.3",
			range2: "[0.2.0, 0.3.0]",
			want:   true,
		},
		{
			name:     "invalid range1",
			range1:   "invalid",
			range2:   "[1.0.0, 2.0.0]",
			wantErr1: true,
		},
		{
			name:     "invalid range2",
			range1:   "[1.0.0, 2.0.0]",
			range2:   "invalid",
			wantErr2: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr1, err1 := ParseVersionRange(tt.range1)
			if tt.wantErr1 {
				require.Error(t, err1)
				return
			}
			require.NoError(t, err1)

			vr2, err2 := ParseVersionRange(tt.range2)
			if tt.wantErr2 {
				require.Error(t, err2)
				return
			}
			require.NoError(t, err2)

			got := vr1.DoesRangeSatisfy(vr2)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestVersionRange_FindBestMatch(t *testing.T) {
	tests := []struct {
		name     string
		range_   string
		versions []string
		want     string
	}{
		{
			name:     "exact version match",
			range_:   "1.0.0",
			versions: []string{"1.0.0", "1.0.1", "1.1.0"},
			want:     "1.0.0",
		},
		{
			name:     "latest patch version",
			range_:   "~1.0.0",
			versions: []string{"1.0.0", "1.0.1", "1.0.2", "1.1.0"},
			want:     "1.0.2",
		},
		{
			name:     "latest minor version",
			range_:   "^1.0.0",
			versions: []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"},
			want:     "1.2.0",
		},
		{
			name:     "prerelease version",
			range_:   "1.0.0-*",
			versions: []string{"1.0.0-alpha", "1.0.0-beta", "1.0.0"},
			want:     "1.0.0-beta",
		},
		{
			name:     "no matching versions",
			range_:   "2.0.0",
			versions: []string{"1.0.0", "1.1.0", "1.2.0"},
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr, err := ParseVersionRange(tt.range_)
			require.NoError(t, err)

			var versions []*semver.Version
			for _, v := range tt.versions {
				version, err := semver.NewVersion(v)
				require.NoError(t, err)
				versions = append(versions, version)
			}

			got := vr.FindBestMatch(versions)
			if tt.want == "" {
				require.Nil(t, got)
			} else {
				require.NotNil(t, got)
				require.Equal(t, tt.want, got.String())
			}
		})
	}
}

func TestVersionRange_ToNonSnapshotRange(t *testing.T) {
	tests := []struct {
		name   string
		range_ string
		want   string
	}{
		{
			name:   "exact version",
			range_: "1.0.0",
			want:   "1.0.0",
		},
		{
			name:   "prerelease with dash",
			range_: "1.0.0-beta-",
			want:   "1.0.0-beta",
		},
		{
			name:   "prerelease with zero",
			range_: "1.0.0-0",
			want:   "1.0.0",
		},
		{
			name:   "floating major",
			range_: "*",
			want:   "*",
		},
		{
			name:   "floating minor",
			range_: "^1.0.0",
			want:   "[1.0.0,2.0.0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr, err := ParseVersionRange(tt.range_)
			require.NoError(t, err)

			got := vr.ToNonSnapshotRange()
			require.Equal(t, tt.want, got.String())
		})
	}
}

func TestVersionRange_PrettyPrint(t *testing.T) {
	tests := []struct {
		name   string
		range_ string
		want   string
	}{
		{
			name:   "exact version",
			range_: "1.0.0",
			want:   "Version 1.0.0 exactly",
		},
		{
			name:   "latest version",
			range_: "*",
			want:   "Latest version",
		},
		{
			name:   "latest minor version",
			range_: "^1.0.0",
			want:   "Latest minor version >= 1.0.0",
		},
		{
			name:   "latest patch version",
			range_: "~1.0.0",
			want:   "Latest patch version >= 1.0.0",
		},
		{
			name:   "inclusive range",
			range_: "[1.0.0,2.0.0]",
			want:   ">= 1.0.0 and <= 2.0.0",
		},
		{
			name:   "exclusive range",
			range_: "(1.0.0,2.0.0)",
			want:   "> 1.0.0 and < 2.0.0",
		},
		{
			name:   "min only",
			range_: "[1.0.0,)",
			want:   ">= 1.0.0",
		},
		{
			name:   "max only",
			range_: "(,2.0.0]",
			want:   "<= 2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr, err := ParseVersionRange(tt.range_)
			require.NoError(t, err)

			got := vr.PrettyPrint()
			require.Equal(t, tt.want, got)
		})
	}
}
