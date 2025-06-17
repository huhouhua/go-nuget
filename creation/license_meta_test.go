// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/require"

	"github.com/huhouhua/go-nuget"
)

func TestNewLicense(t *testing.T) {
	wantVersion := semver.New(1, 0, 0, "", "")
	t.Run("with File", func(t *testing.T) {
		license := NewLicense(nuget.File, "/docs/LICENSE", wantVersion)
		require.Equal(t, "/docs/LICENSE", license.GetLicense())
		require.Equal(t, nuget.File, license.GetLicenseType())
		require.Equal(t, wantVersion, license.GetVersion())
		actual, err := license.GetLicenseURL()
		require.NoError(t, err)
		require.Equal(t, &LicenseFileDeprecationURL, actual)
	})
	t.Run("with Expression", func(t *testing.T) {
		license := NewLicense(nuget.Expression, "MIT", wantVersion)
		require.Equal(t, "MIT", license.GetLicense())
		require.Equal(t, nuget.Expression, license.GetLicenseType())
		require.Equal(t, wantVersion, license.GetVersion())
		actual, err := license.GetLicenseURL()
		require.NoError(t, err)
		expected := &url.URL{Scheme: "https", Host: "licenses.nuget.org", Path: "/MIT"}
		require.Equal(t, actual, expected)
	})
	t.Run("no supported", func(t *testing.T) {
		license := NewLicense("unsupported", "Apache", wantVersion)
		require.Equal(t, "Apache", license.GetLicense())
		require.Equal(t, nuget.LicenseType("unsupported"), license.GetLicenseType())
		require.Equal(t, wantVersion, license.GetVersion())
		actual, err := license.GetLicenseURL()
		require.Nil(t, actual)
		require.Equal(t, err, fmt.Errorf("unsupported license type: unsupported"))
	})
}
