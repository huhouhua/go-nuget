// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/huhouhua/go-nuget/version"

	"github.com/stretchr/testify/require"
)

func TestNewLicense(t *testing.T) {
	wantVersion := version.NewVersionFrom(1, 0, 0, "", "")
	t.Run("with File", func(t *testing.T) {
		license := NewLicense(File, "/docs/LICENSE", wantVersion)
		require.Equal(t, "/docs/LICENSE", license.GetLicense())
		require.Equal(t, File, license.GetLicenseType())
		require.Equal(t, wantVersion, license.GetVersion())
		actual, err := license.GetLicenseURL()
		require.NoError(t, err)
		require.Equal(t, &LicenseFileDeprecationURL, actual)
	})
	t.Run("with Expression", func(t *testing.T) {
		license := NewLicense(Expression, "MIT", wantVersion)
		require.Equal(t, "MIT", license.GetLicense())
		require.Equal(t, Expression, license.GetLicenseType())
		require.Equal(t, wantVersion, license.GetVersion())
		actual, err := license.GetLicenseURL()
		require.NoError(t, err)
		expected := &url.URL{Scheme: "https", Host: "licenses.org", Path: "/MIT"}
		require.Equal(t, expected, actual)
	})
	t.Run("no supported", func(t *testing.T) {
		license := NewLicense("unsupported", "Apache", wantVersion)
		require.Equal(t, "Apache", license.GetLicense())
		require.Equal(t, LicenseType("unsupported"), license.GetLicenseType())
		require.Equal(t, wantVersion, license.GetVersion())
		actual, err := license.GetLicenseURL()
		require.Nil(t, actual)
		require.Equal(t, err, fmt.Errorf("unsupported license type: unsupported"))
	})
}
