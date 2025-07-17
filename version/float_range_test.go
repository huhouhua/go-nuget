// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package version

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFloatRange_ParseBasic(t *testing.T) {
	// Act
	versionRange, err := ParseFloatRange("1.0.0")
	require.NoError(t, err)
	require.Equal(t, versionRange.MinVersion, versionRange.MinVersion)
	require.Equal(t, versionRange.FloatBehavior, None)
}

func TestFloatingRange_FloatNone(t *testing.T) {
	versionRange, err := ParseFloatRange("1.0.0")
	require.NoError(t, err)

	require.Equal(t, "1.0.0", versionRange.MinVersion.Semver.String())
	require.Equal(t, None, versionRange.FloatBehavior)
}

func TestFloatingRange_FloatPre(t *testing.T) {
	versionRange, err := ParseFloatRange("1.0.0-*")
	require.NoError(t, err)

	require.Equal(t, "1.0.0-0", versionRange.MinVersion.Semver.String())
	require.Equal(t, Prerelease, versionRange.FloatBehavior)
}

func TestFloatingRange_FloatPrePrefix(t *testing.T) {
	versionRange, err := ParseFloatRange("1.0.0-alpha-*")
	require.NoError(t, err)

	require.Equal(t, "1.0.0-alpha-", versionRange.MinVersion.Semver.String())
	require.Equal(t, Prerelease, versionRange.FloatBehavior)
}
func TestFloatingRange_FloatRev(t *testing.T) {
	versionRange, err := ParseFloatRange("1.0.0.*")
	require.NoError(t, err)

	require.Equal(t, "1.0.0", versionRange.MinVersion.Semver.String())
	require.Equal(t, Revision, versionRange.FloatBehavior)
}

func TestFloatingRange_FloatPatch(t *testing.T) {
	versionRange, err := ParseFloatRange("1.0.*")
	require.NoError(t, err)

	require.Equal(t, "1.0.0", versionRange.MinVersion.Semver.String())
	require.Equal(t, Patch, versionRange.FloatBehavior)
}
func TestFloatingRange_FloatMinor(t *testing.T) {
	versionRange, err := ParseFloatRange("1.*")
	require.NoError(t, err)

	require.Equal(t, "1.0.0", versionRange.MinVersion.Semver.String())
	require.Equal(t, Minor, versionRange.FloatBehavior)
}

func TestFloatingRange_FloatMajor(t *testing.T) {
	versionRange, err := ParseFloatRange("*")
	require.NoError(t, err)

	require.Equal(t, "0.0.0", versionRange.MinVersion.Semver.String())
	require.Equal(t, Major, versionRange.FloatBehavior)
}

func TestFloatingRange_ToStringPre(t *testing.T) {
	// Act
	versionRange, err := ParseRange("1.0.0-*")
	require.NoError(t, err)

	normalized, err := versionRange.ToNormalizedString()
	require.NoError(t, err)

	require.Equal(t, "[1.0.0-*, )", normalized)
}
func TestFloatingRange_ToStringPrePrefix(t *testing.T) {
	// Act
	versionRange, err := ParseRange("1.0.0-alpha.*")
	require.NoError(t, err)

	normalized, err := versionRange.ToNormalizedString()
	require.NoError(t, err)

	require.Equal(t, "[1.0.0-alpha.*, )", normalized)
}

func TestFloatingRange_ToStringRev(t *testing.T) {
	// Act
	versionRange, err := ParseRange("1.0.0.*")
	require.NoError(t, err)

	normalized, err := versionRange.ToNormalizedString()
	require.NoError(t, err)

	require.Equal(t, "[1.0.0.*, )", normalized)
}

func TestFloatingRange_ToStringPatch(t *testing.T) {
	// Act
	versionRange, err := ParseRange("1.0.*")
	require.NoError(t, err)

	normalized, err := versionRange.ToNormalizedString()
	require.NoError(t, err)

	require.Equal(t, "[1.0.*, )", normalized)
}
func TestFloatingRange_ToStringMinor(t *testing.T) {
	// Act
	versionRange, err := ParseRange("1.*")
	require.NoError(t, err)

	normalized, err := versionRange.ToNormalizedString()
	require.NoError(t, err)

	require.Equal(t, "[1.*, )", normalized)
}
