// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPackageMetadataResource_ListMetadata(t *testing.T) {
	mux, client := setup(t, "testdata/index_2.json")

	baseURL := client.getResourceUrl(RegistrationsBaseUrl)
	u := fmt.Sprintf("%s/gitlabapiclient/index.json", baseURL.Path)

	mux.HandleFunc(u, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/medata.json")
	})

	versionrange1203, err := ParseVersionRange("[12.0.3, )")
	require.NoError(t, err)

	publishedTime, err := time.Parse(time.RFC3339, "2025-04-18T09:41:56.5124797Z")
	require.NoError(t, err)

	rawURL := fmt.Sprintf(
		"%s://%s/packages/gitlabapiclient/1.8.1-beta.5/ReportAbuse",
		client.baseURL.Scheme,
		client.baseURL.Host,
	)
	reportUrl, err := url.Parse(rawURL)
	require.NoError(t, err)

	want := []*PackageSearchMetadataRegistration{
		{
			ReportAbuseUrl: reportUrl,
			Authors:        "nmklotas",
			SearchMetadata: &SearchMetadata{
				PackageId: "GitLabApiClient",
				Version:   "1.8.1-beta.5",
				DependencySets: []*PackageDependencyGroup{
					{
						TargetFramework: "net48",
						Packages: []*Dependency{
							{
								Id:              "Newtonsoft.Json",
								VersionRangeRaw: "[12.0.3, )",
								VersionRange:    versionrange1203,
							},
						},
					},
					{
						TargetFramework: "netcoreapp3.1",
						Packages: []*Dependency{
							{
								Id:              "Newtonsoft.Json",
								VersionRangeRaw: "[12.0.3, )",
								VersionRange:    versionrange1203,
							},
						},
					},
					{
						TargetFramework: "net5.0",
						Packages: []*Dependency{
							{
								Id:              "Newtonsoft.Json",
								VersionRangeRaw: "[12.0.3, )",
								VersionRange:    versionrange1203,
							},
						},
					},
					{
						TargetFramework: "netstandard2.0",
						Packages: []*Dependency{
							{
								Id:              "Newtonsoft.Json",
								VersionRangeRaw: "[12.0.3, )",
								VersionRange:    versionrange1203,
							},
						},
					},
				},
				Description:              "GitLabApiClient is a .NET rest client for GitLab API v4.",
				DownloadCount:            0,
				LicenseUrl:               "https://licenses.nuget.org/MIT",
				ProjectUrl:               "https://github.com/nmklotas/GitLabApiClient",
				Published:                publishedTime,
				RequireLicenseAcceptance: false,
				Tags: []string{
					"GitLab",
					"REST",
					"API",
					"CI",
					"Client",
				},
				IsListed:       true,
				PrefixReserved: false,
			},
		},
	}
	b, resp, err := client.MetadataResource.ListMetadata("gitlabapiclient", &ListMetadataOptions{
		IncludePrerelease: true,
		IncludeUnlisted:   false,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, b)
}

func TestPackageMetadataResource_GetMetadata(t *testing.T) {
	mux, client := setup(t, "testdata/index_2.json")

	baseURL := client.getResourceUrl(RegistrationsBaseUrl)
	u := fmt.Sprintf("%s/gitlabapiclient/index.json", baseURL.Path)

	mux.HandleFunc(u, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/medata.json")
	})

	versionrange1203, err := ParseVersionRange("[12.0.3, )")
	require.NoError(t, err)

	publishedTime, err := time.Parse(time.RFC3339, "2025-04-18T09:41:56.5124797Z")
	require.NoError(t, err)

	rawURL := fmt.Sprintf(
		"%s://%s/packages/gitlabapiclient/1.8.1-beta.5/ReportAbuse",
		client.baseURL.Scheme,
		client.baseURL.Host,
	)
	reportUrl, err := url.Parse(rawURL)
	require.NoError(t, err)

	want := &PackageSearchMetadataRegistration{
		ReportAbuseUrl: reportUrl,
		Authors:        "nmklotas",
		SearchMetadata: &SearchMetadata{
			PackageId: "GitLabApiClient",
			Version:   "1.8.1-beta.5",
			DependencySets: []*PackageDependencyGroup{
				{
					TargetFramework: "net48",
					Packages: []*Dependency{
						{
							Id:              "Newtonsoft.Json",
							VersionRangeRaw: "[12.0.3, )",
							VersionRange:    versionrange1203,
						},
					},
				},
				{
					TargetFramework: "netcoreapp3.1",
					Packages: []*Dependency{
						{
							Id:              "Newtonsoft.Json",
							VersionRangeRaw: "[12.0.3, )",
							VersionRange:    versionrange1203,
						},
					},
				},
				{
					TargetFramework: "net5.0",
					Packages: []*Dependency{
						{
							Id:              "Newtonsoft.Json",
							VersionRangeRaw: "[12.0.3, )",
							VersionRange:    versionrange1203,
						},
					},
				},
				{
					TargetFramework: "netstandard2.0",
					Packages: []*Dependency{
						{
							Id:              "Newtonsoft.Json",
							VersionRangeRaw: "[12.0.3, )",
							VersionRange:    versionrange1203,
						},
					},
				},
			},
			Description:              "GitLabApiClient is a .NET rest client for GitLab API v4.",
			DownloadCount:            0,
			LicenseUrl:               "https://licenses.nuget.org/MIT",
			ProjectUrl:               "https://github.com/nmklotas/GitLabApiClient",
			Published:                publishedTime,
			RequireLicenseAcceptance: false,
			Tags: []string{
				"GitLab",
				"REST",
				"API",
				"CI",
				"Client",
			},
			IsListed:       true,
			PrefixReserved: false,
		},
	}
	b, resp, err := client.MetadataResource.GetMetadata("gitlabapiclient", "1.8.1-beta.5")
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, b)
}
