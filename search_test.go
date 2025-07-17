// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	version2 "github.com/huhouhua/go-nuget/version"

	"github.com/Masterminds/semver/v3"

	"github.com/stretchr/testify/require"
)

func TestPackageSearchResource_Search(t *testing.T) {
	mux, client := setup(t, index_V3)

	baseURL := client.getResourceURL(SearchQueryService)
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
				IconURL:                  "https://api.nuget.org/v3-flatcontainer/newtonsoft.json/13.0.3/icon",
				LicenseURL:               "https://www.nuget.org/packages/Newtonsoft.Json/13.0.3/license",
				ProjectURL:               "https://www.newtonsoft.com/json",
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

func TestSearchPackageUrl(t *testing.T) {
	wantError := url.EscapeError("%qu")

	_, client := setup(t, index_V3)
	require.NotNil(t, client)

	u, err := url.Parse("")
	u.Path = "%query"
	require.NoError(t, err)

	client.serviceURLs[SearchQueryService] = u

	_, _, err = client.SearchResource.Search(&SearchOptions{})
	require.Equal(t, wantError, err)
}

func TestSearchOptions(t *testing.T) {
	mux, client := setup(t, index_V3)
	require.NotNil(t, client)
	baseURL := client.getResourceURL(SearchQueryService)
	mux.HandleFunc(baseURL.Path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if r.URL.Query().Get("q") == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("searchOptions request 'q' parameter is missing;"))
			require.NoError(t, err)
			return
		}
		mustWriteHTTPResponse(t, w, "testdata/search.json")
	})

	tests := []struct {
		name   string
		opt    *SearchOptions
		errMsg string
	}{
		{
			name:   "nil",
			opt:    nil,
			errMsg: "failed to parse unknown error format: searchOptions request 'q' parameter is missing; 400 Bad Request",
		},
		{
			name:   "query parameter has not been set",
			opt:    &SearchOptions{},
			errMsg: "failed to parse unknown error format: searchOptions request 'q' parameter is missing; 400 Bad Request",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			search, resp, err := client.SearchResource.Search(tt.opt)
			var errResp *ErrorResponse
			if err != nil && errors.As(err, &errResp) {
				require.Equal(t, tt.errMsg, errResp.Message)
				return
			}
			require.NotNil(t, resp)
			require.NotNil(t, search)
		})
	}
}

func TestAddSemVer(t *testing.T) {
	tests := []struct {
		name string
		url  *url.URL
		want *url.URL
	}{
		{
			name: "add semVerLevel",
			url:  createUrl(t, "https://127.0.0.1/query"),
			want: createUrl(t, "https://127.0.0.1/query?semVerLevel=2.0.0"),
		},
		{
			name: "already existed semVerLevel",
			url:  createUrl(t, "https://127.0.0.1/query?semVerLevel=2.0.0"),
			want: createUrl(t, "https://127.0.0.1/query?semVerLevel=2.0.0"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addSemVer(tt.url)
			require.Equal(t, tt.want, tt.url)
		})
	}
}

func createUrl(t *testing.T, u string) *url.URL {
	rawUrl, err := url.Parse(u)
	require.NoError(t, err)
	return rawUrl
}

func TestParseVersionInfo(t *testing.T) {
	symbolVersion, err := semver.NewVersion("1-0-0")
	require.NoError(t, err)

	vZeroes, err := semver.NewVersion("00000.0000.0")
	require.NoError(t, err)

	tests := []struct {
		name    string
		version *VersionInfo
		want    *version2.Version
		error   error
	}{
		{
			name: "beta",
			version: &VersionInfo{
				Version: "1.0.0-beta",
			},
			want:  version2.NewVersionFrom(1, 0, 0, "beta", ""),
			error: nil,
		},
		{
			name: "beta with last number",
			version: &VersionInfo{
				Version: "1.0.0-beta.1",
			},
			want:  version2.NewVersionFrom(1, 0, 0, "beta.1", ""),
			error: nil,
		},
		{
			name: "pre-release with last number",
			version: &VersionInfo{
				Version: "1.0.0-preview.1",
			},
			want:  version2.NewVersionFrom(1, 0, 0, "preview.1", ""),
			error: nil,
		},
		{
			name: "alpha with last number",
			version: &VersionInfo{
				Version: "1.0.0-alpha.1",
			},
			want:  version2.NewVersionFrom(1, 0, 0, "alpha.1", ""),
			error: nil,
		},
		{
			name: "rc with build sha",
			version: &VersionInfo{
				Version: "1.0.0-rc.22997fbc939e55215eb5162aa4ad6edafe4e7b65",
			},
			want:  version2.NewVersionFrom(1, 0, 0, "rc.22997fbc939e55215eb5162aa4ad6edafe4e7b65", ""),
			error: nil,
		},
		{
			name: "with symbol",
			version: &VersionInfo{
				Version: "1-0-0",
			},
			want:  version2.NewVersion(symbolVersion, 0, symbolVersion.Original()),
			error: nil,
		},
		{
			name: "all zeroes",
			version: &VersionInfo{
				Version: "00000.0000.0",
			},
			want:  version2.NewVersion(vZeroes, 0, "00000.0000.0"),
			error: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := tt.version.ParseVersion()
			require.Equal(t, tt.error, err)
			require.Equal(t, tt.want, actual)
		})
	}
}
