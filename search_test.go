// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"errors"
	"github.com/Masterminds/semver/v3"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPackageSearchResource_Search(t *testing.T) {
	mux, client := setup(t, "testdata/index.json")

	baseURL := client.getResourceUrl(SearchQueryService)
	mux.HandleFunc(baseURL.Path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/search.json")
	})

	publishedTime, err := time.Parse(time.RFC3339, "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	want := []*PackageSearchMetadata{
		{
			SearchMetadata: &SearchMetadata{
				PackageId:                "Newtonsoft.Json",
				Version:                  "13.0.3",
				Description:              "Json.NET is a popular high-performance JSON framework for .NET",
				DownloadCount:            6111703093,
				IconUrl:                  "https://api.nuget.org/v3-flatcontainer/newtonsoft.json/13.0.3/icon",
				LicenseUrl:               "https://www.nuget.org/packages/Newtonsoft.Json/13.0.3/license",
				ProjectUrl:               "https://www.newtonsoft.com/json",
				Published:                publishedTime,
				RequireLicenseAcceptance: false,
				Tags:                     []string{"json"},
				Title:                    "Json.NET",
				IsListed:                 false,
				Vulnerabilities:          []*PackageVulnerabilityMetadata{},
				PrefixReserved:           true,
			},
			Versions: []*VersionInfo{
				{
					Url:           "https://api.nuget.org/v3/registration5-gz-semver2/newtonsoft.json/3.5.8.json",
					Version:       "3.5.8",
					DownloadCount: 4342578,
				},
			},
			Authors: []string{"James Newton-King"},
			Owners:  []string{"dotnetfoundation"},
		},
	}

	opt := &SearchOptions{
		SearchTerm:        "json",
		IncludePrerelease: true,
		Skip:              0,
		Take:              10,
	}
	b, resp, err := client.SearchResource.Search(opt)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, b)
}

func TestParseVersion(t *testing.T) {
	symbolVersion, err := semver.NewVersion("1-0-0")
	require.NoError(t, err)
	tests := []struct {
		name    string
		version *VersionInfo
		want    *NuGetVersion
		error   error
	}{
		{
			name: "beta",
			version: &VersionInfo{
				Version: "1.0.0-beta",
			},
			want: &NuGetVersion{
				Version: semver.New(1, 0, 0, "beta", ""),
			},
			error: nil,
		},
		{
			name: "beta with last number",
			version: &VersionInfo{
				Version: "1.0.0-beta.1",
			},
			want: &NuGetVersion{
				Version: semver.New(1, 0, 0, "beta.1", ""),
			},
			error: nil,
		},
		{
			name: "pre-release with last number",
			version: &VersionInfo{
				Version: "1.0.0-preview.1",
			},
			want: &NuGetVersion{
				Version: semver.New(1, 0, 0, "preview.1", ""),
			},
			error: nil,
		},
		{
			name: "alpha with last number",
			version: &VersionInfo{
				Version: "1.0.0-alpha.1",
			},
			want: &NuGetVersion{
				Version: semver.New(1, 0, 0, "alpha.1", ""),
			},
			error: nil,
		},
		{
			name: "rc with build sha",
			version: &VersionInfo{
				Version: "1.0.0-rc.22997fbc939e55215eb5162aa4ad6edafe4e7b65",
			},
			want: &NuGetVersion{
				Version: semver.New(1, 0, 0, "rc.22997fbc939e55215eb5162aa4ad6edafe4e7b65", ""),
			},
			error: nil,
		},
		{
			name: "with symbol",
			version: &VersionInfo{
				Version: "1-0-0",
			},
			want: &NuGetVersion{
				Version: symbolVersion,
			},
			error: nil,
		},
		{
			name: "all zeroes",
			version: &VersionInfo{
				Version: "00000.0000.0",
			},
			want:  nil,
			error: errors.New("Invalid Semantic Version"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := tt.version.ParseVersion()
			require.Equal(t, err, tt.error)
			require.Equal(t, v, tt.want)
		})
	}
}
