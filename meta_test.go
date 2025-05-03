// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"

	"github.com/stretchr/testify/require"
)

func TestPackageMetadataResource_ListMetadata(t *testing.T) {
	mux, client := setup(t, index_Baget)

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
	p, resp, err := client.MetadataResource.ListMetadata("gitlabapiclient", &ListMetadataOptions{
		IncludePrerelease: true,
		IncludeUnlisted:   false,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, p)
}

func TestPackageMetadataResource_GetMetadata(t *testing.T) {
	mux, client := setup(t, index_Baget)

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
	t.Run("compare results", func(t *testing.T) {
		b, resp, err := client.MetadataResource.GetMetadata("gitlabapiclient", "1.8.1-beta.5")
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, want, b)
	})
	t.Run("invalid version", func(t *testing.T) {
		wantErr := errors.New("Invalid Semantic Version")
		_, _, err = client.MetadataResource.GetMetadata("json", "x.0.0")
		require.Equal(t, wantErr, err)
	})
}

func TestPackageSearchMetadataRegistration(t *testing.T) {
	t.Run("parse identity", func(t *testing.T) {
		input := &PackageSearchMetadataRegistration{
			SearchMetadata: &SearchMetadata{
				PackageId: "json",
				Version:   "1.0.0-beta",
			},
			Owners: "Kevin Berger,test2,test3",
		}
		wantIdentity := &PackageIdentity{
			Id: input.SearchMetadata.PackageId,
			Version: &NuGetVersion{
				Version: semver.New(1, 0, 0, "beta", ""),
			},
		}
		identity, err := input.Identity()
		require.NoError(t, err)
		require.Equal(t, wantIdentity, identity)

		wantOwners := []string{"Kevin Berger", "test2", "test3"}
		require.Equal(t, wantOwners, input.OwnersList())
	})
	t.Run("invalid version", func(t *testing.T) {
		inputErr := &PackageSearchMetadataRegistration{
			SearchMetadata: &SearchMetadata{
				PackageId: "json",
				Version:   "^0.0.1",
			},
		}
		wantErr := errors.New("Invalid Semantic Version")
		_, err := inputErr.Identity()
		require.Equal(t, wantErr, err)
	})
}

func TestParseAndReplaceUrl(t *testing.T) {
	invalidUrlTemplate := createUrl(t, "https://example.com/packages/{id}/{version}")
	invalidUrlTemplate.Path = invalidUrlTemplate.Path + "%%details"

	unescapeUrlTemplate := createUrl(t, "")
	unescapeUrlTemplate.Scheme = "%eth0"

	tests := []struct {
		name         string
		urlTemplate  *url.URL
		replacements map[string]string
		want         *url.URL
		error        error
	}{
		{
			name:        "valid replacements",
			urlTemplate: createUrl(t, "https://example.com/packages/{id}/{version}/details"),
			replacements: map[string]string{
				"{id}":      "testpackage",
				"{version}": "1.0.0",
			},
			want: createUrl(t, "https://example.com/packages/testpackage/1.0.0/details"),
		},
		{
			name:        "nil template",
			urlTemplate: nil,
			replacements: map[string]string{
				"{id}":      "testpackage",
				"{version}": "1.0.0",
			},
			want: nil,
		},
		{
			name:        "invalid url parsing",
			urlTemplate: invalidUrlTemplate,
			replacements: map[string]string{
				"{id}":      "testpackage",
				"{version}": "1.0.0",
			},
			want: nil,
			error: &url.Error{
				Op:  "parse",
				URL: "https://example.com/packages/testpackage/1.0.0%%details",
				Err: url.EscapeError("%%d"),
			},
		},
		{
			name:         "unescape url error",
			urlTemplate:  unescapeUrlTemplate,
			replacements: nil,
			want:         nil,
			error:        url.EscapeError("%et"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseAndReplaceUrl(*tt.urlTemplate, tt.replacements)
			require.Equal(t, tt.error, err)
			require.Equal(t, tt.want, actual)
		})
	}
}

func TestWithReportAbuseUrl(t *testing.T) {
	unescapeUrlTemplate := createUrl(t, "")
	unescapeUrlTemplate.Scheme = "%eth0"
	tests := []struct {
		name        string
		urlTemplate *url.URL
		metadata    *PackageSearchMetadataRegistration
		want        *url.URL
		error       error
	}{
		{
			name:        "valid url template",
			urlTemplate: createUrl(t, "https://example.com/packages/{id}/{version}/ReportAbuse"),
			metadata: &PackageSearchMetadataRegistration{
				SearchMetadata: &SearchMetadata{
					PackageId: "TestPackage",
					Version:   "1.0.0",
				},
			},
			want: createUrl(t, "https://example.com/packages/testpackage/1.0.0/ReportAbuse"),
		},
		{
			name:        "nil url template",
			urlTemplate: nil,
			metadata: &PackageSearchMetadataRegistration{
				SearchMetadata: &SearchMetadata{
					PackageId: "TestPackage",
					Version:   "1.0.0",
				},
			},
			want: nil,
		},
		{
			name:        "unescape url error",
			urlTemplate: unescapeUrlTemplate,
			metadata: &PackageSearchMetadataRegistration{
				SearchMetadata: &SearchMetadata{
					PackageId: "TestPackage",
					Version:   "1.0.0",
				},
			},
			want:  nil,
			error: url.EscapeError("%et"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WithReportAbuseUrl(*tt.urlTemplate)(tt.metadata)
			require.Equal(t, tt.error, err)
			require.Equal(t, tt.want, tt.metadata.ReportAbuseUrl)
		})
	}
}

func TestWithPackageDetailsUrl(t *testing.T) {
	unescapeUrlTemplate := createUrl(t, "")
	unescapeUrlTemplate.Scheme = "%eth0"
	tests := []struct {
		name        string
		urlTemplate *url.URL
		metadata    *PackageSearchMetadataRegistration
		want        *url.URL
		error       error
	}{
		{
			name:        "valid url template",
			urlTemplate: createUrl(t, "https://example.com/packages/{id}/{version}?_src=template"),
			metadata: &PackageSearchMetadataRegistration{
				SearchMetadata: &SearchMetadata{
					PackageId: "TestPackage",
					Version:   "1.0.0",
				},
			},
			want: createUrl(t, "https://example.com/packages/testpackage/1.0.0?_src=template"),
		},
		{
			name:        "nil url template",
			urlTemplate: nil,
			metadata: &PackageSearchMetadataRegistration{
				SearchMetadata: &SearchMetadata{
					PackageId: "TestPackage",
					Version:   "1.0.0",
				},
			},
			want: nil,
		},
		{
			name:        "unescape url error",
			urlTemplate: unescapeUrlTemplate,
			metadata: &PackageSearchMetadataRegistration{
				SearchMetadata: &SearchMetadata{
					PackageId: "TestPackage",
					Version:   "1.0.0",
				},
			},
			want:  nil,
			error: url.EscapeError("%et"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WithPackageDetailsUrl(*tt.urlTemplate)(tt.metadata)
			require.Equal(t, tt.error, err)
			require.Equal(t, tt.want, tt.metadata.PackageDetailsUrl)
		})
	}
}

func TestWithReadmeFileUrl(t *testing.T) {
	unescapeUrlTemplate := createUrl(t, "")
	unescapeUrlTemplate.Scheme = "%eth0"
	tests := []struct {
		name        string
		urlTemplate *url.URL
		metadata    *PackageSearchMetadataRegistration
		want        *url.URL
		error       error
	}{
		{
			name:        "valid url template",
			urlTemplate: createUrl(t, "https://example.com/v3-flatcontainer/{lower_id}/{lower_version}/readme"),
			metadata: &PackageSearchMetadataRegistration{
				SearchMetadata: &SearchMetadata{
					PackageId: "TestPackage",
					Version:   "1.0.0",
				},
			},
			want: createUrl(t, "https://example.com/v3-flatcontainer/testpackage/1.0.0/readme"),
		},
		{
			name:        "nil url template",
			urlTemplate: nil,
			metadata: &PackageSearchMetadataRegistration{
				SearchMetadata: &SearchMetadata{
					PackageId: "TestPackage",
					Version:   "1.0.0",
				},
			},
			want: nil,
		}, {
			name:        "unescape url error",
			urlTemplate: unescapeUrlTemplate,
			metadata: &PackageSearchMetadataRegistration{
				SearchMetadata: &SearchMetadata{
					PackageId: "TestPackage",
					Version:   "1.0.0",
				},
			},
			want:  nil,
			error: url.EscapeError("%et"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WithReadmeFileUrl(*tt.urlTemplate)(tt.metadata)
			require.Equal(t, tt.error, err)
			require.Equal(t, tt.want, tt.metadata.ReadmeFileUrl)
		})
	}
}

func TestApplyMetadataRegistration(t *testing.T) {
	t.Run("Apply All Metadata Functions", func(t *testing.T) {
		reportAbuseUrlTemplate := createUrl(t, "https://example.com/packages/{id}/{version}/ReportAbuse")
		detailsUrlTemplate := createUrl(t, "https://example.com/packages/{id}/{version}?_src=template")
		readmeUrlTemplate := createUrl(t, "https://example.com/v3-flatcontainer/{lower_id}/{lower_version}/readme")

		metadata := &PackageSearchMetadataRegistration{
			SearchMetadata: &SearchMetadata{
				PackageId: "TestPackage",
				Version:   "1.0.0",
			},
		}

		err := ApplyMetadataRegistration(metadata,
			WithReportAbuseUrl(*reportAbuseUrlTemplate),
			WithPackageDetailsUrl(*detailsUrlTemplate),
			WithReadmeFileUrl(*readmeUrlTemplate),
		)

		require.NoError(t, err)
		require.Equal(t, "https://example.com/packages/testpackage/1.0.0/ReportAbuse", metadata.ReportAbuseUrl.String())
		require.Equal(
			t,
			"https://example.com/packages/testpackage/1.0.0?_src=template",
			metadata.PackageDetailsUrl.String(),
		)
		require.Equal(
			t,
			"https://example.com/v3-flatcontainer/testpackage/1.0.0/readme",
			metadata.ReadmeFileUrl.String(),
		)
	})

	t.Run("Handle Errors in Metadata Functions", func(t *testing.T) {
		invalidUrlTemplate := createUrl(t, "https://example.com/packages/{id}/{version}")
		invalidUrlTemplate.Path = invalidUrlTemplate.Path + "%%ReportAbuse"

		metadata := &PackageSearchMetadataRegistration{
			SearchMetadata: &SearchMetadata{
				PackageId: "TestPackage",
				Version:   "1.0.0",
			},
		}

		err := ApplyMetadataRegistration(metadata,
			WithReportAbuseUrl(*invalidUrlTemplate),
		)
		wantErr := &url.Error{
			Op:  "parse",
			URL: "https://example.com/packages/testpackage/1.0.0%%ReportAbuse",
			Err: url.EscapeError("%%R"),
		}
		require.Equal(t, wantErr, err)
	})
}

func TestAddMetadataToPackages(t *testing.T) {
	versionRange, err := ParseVersionRange("[1.5.0, )")
	require.NoError(t, err)

	_, client := setup(t, index_V3)
	require.NotNil(t, client)

	emptyPkg := make([]*PackageSearchMetadataRegistration, 0)

	var (
		tests = []struct {
			name    string
			page    *registrationPage
			options *ListMetadataOptions
			error   error
			wantPkg []*PackageSearchMetadataRegistration
		}{
			{
				name: "Valid package in range",
				page: &registrationPage{
					Lower: "1.0.0",
					Upper: "2.0.0",
					Items: []*registrationLeafItem{
						{
							CatalogEntry: &PackageSearchMetadataRegistration{
								SearchMetadata: &SearchMetadata{
									PackageId: "TestPackage",
									Version:   "1.5.0",
									IsListed:  true,
								},
							},
						},
					},
				},
				options: &ListMetadataOptions{
					IncludePrerelease: true,
					IncludeUnlisted:   false,
				},
				error: nil,
				wantPkg: []*PackageSearchMetadataRegistration{
					{
						SearchMetadata: &SearchMetadata{
							PackageId: "TestPackage",
							Version:   "1.5.0",
							IsListed:  true,
						},
						PackageDetailsUrl: createUrl(t, fmt.Sprintf("%s/packages/testpackage/1.5.0?_src=template", client.baseURL.String())),
						ReadmeFileUrl:     createUrl(t, fmt.Sprintf("%s/v3-flatcontainer/testpackage/1.5.0/readme", client.baseURL.String())),
						ReportAbuseUrl:    createUrl(t, fmt.Sprintf("%s/packages/testpackage/1.5.0/ReportAbuse", client.baseURL.String())),
					},
				},
			},
			{
				name: "Invalid lower version in page",
				page: &registrationPage{
					Lower: "invalid-lower-version",
					Upper: "2.0.0",
				},
				options: &ListMetadataOptions{
					IncludePrerelease: true,
					IncludeUnlisted:   true,
				},
				wantPkg: emptyPkg,
				error:   errors.New("Invalid Semantic Version"),
			},
			{
				name: "Invalid upper version in page",
				page: &registrationPage{
					Lower: "1.0.0",
					Upper: "invalid-upper-version",
				},
				options: &ListMetadataOptions{
					IncludePrerelease: true,
					IncludeUnlisted:   true,
				},
				wantPkg: emptyPkg,
				error:   errors.New("Invalid Semantic Version"),
			},
			{
				name: "Version out of range",
				page: &registrationPage{
					Lower: "3.0.0",
					Upper: "4.0.0",
				},
				options: &ListMetadataOptions{
					IncludePrerelease: true,
					IncludeUnlisted:   true,
				},
				wantPkg: emptyPkg,
				error:   nil,
			},
			{
				name: "IncludeUnlisted is false and package is unlisted",
				page: &registrationPage{
					Lower: "1.0.0",
					Upper: "2.0.0",
					Items: []*registrationLeafItem{
						{
							CatalogEntry: &PackageSearchMetadataRegistration{
								SearchMetadata: &SearchMetadata{
									PackageId: "TestPackage",
									Version:   "1.5.0",
									IsListed:  false,
								},
							},
						},
					},
				},
				options: &ListMetadataOptions{
					IncludePrerelease: true,
					IncludeUnlisted:   false,
				},
				wantPkg: emptyPkg,
				error:   nil,
			},
			{
				name: "parse dependencySets",
				page: &registrationPage{
					Lower: "1.5.0",
					Upper: "1.5.0",
					Items: []*registrationLeafItem{
						{
							CatalogEntry: &PackageSearchMetadataRegistration{
								SearchMetadata: &SearchMetadata{
									PackageId: "TestPackage",
									Version:   "1.5.0",
									IsListed:  true,
									DependencySets: []*PackageDependencyGroup{
										{
											TargetFramework: "net48",
											Packages: []*Dependency{
												{
													Id:              "Newtonsoft.Json",
													VersionRangeRaw: "[1.5.0, )",
													VersionRange:    versionRange,
												},
											},
										},
									},
								},
							},
						},
					},
				},
				options: &ListMetadataOptions{
					IncludePrerelease: true,
					IncludeUnlisted:   true,
				},
				error: nil,
				wantPkg: []*PackageSearchMetadataRegistration{
					{
						SearchMetadata: &SearchMetadata{
							PackageId: "TestPackage",
							Version:   "1.5.0",
							DependencySets: []*PackageDependencyGroup{
								{
									TargetFramework: "net48",
									Packages: []*Dependency{
										{
											Id:              "Newtonsoft.Json",
											VersionRangeRaw: "[1.5.0, )",
											VersionRange:    versionRange,
										},
									},
								},
							},
							IsListed: true,
						},
						PackageDetailsUrl: createUrl(t, fmt.Sprintf("%s/packages/testpackage/1.5.0?_src=template", client.baseURL.String())),
						ReadmeFileUrl:     createUrl(t, fmt.Sprintf("%s/v3-flatcontainer/testpackage/1.5.0/readme", client.baseURL.String())),
						ReportAbuseUrl:    createUrl(t, fmt.Sprintf("%s/packages/testpackage/1.5.0/ReportAbuse", client.baseURL.String())),
					},
				},
			},
			{
				name: "parse dependencySets fail",
				page: &registrationPage{
					Lower: "1.5.0",
					Upper: "1.5.0",
					Items: []*registrationLeafItem{
						{
							CatalogEntry: &PackageSearchMetadataRegistration{
								SearchMetadata: &SearchMetadata{
									PackageId: "TestPackage",
									Version:   "1.5.0",
									IsListed:  true,
									DependencySets: []*PackageDependencyGroup{
										{
											TargetFramework: "net48",
											Packages: []*Dependency{
												{
													Id:              "Newtonsoft.Json",
													VersionRangeRaw: "[1.5.0, )",
													VersionRange:    versionRange,
													VersionRaw:      "1.0.0%",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				options: &ListMetadataOptions{
					IncludePrerelease: true,
					IncludeUnlisted:   true,
				},
				error:   errors.New("invalid version: Invalid Semantic Version"),
				wantPkg: emptyPkg,
			},
		}
	)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packages := make([]*PackageSearchMetadataRegistration, 0)

			err = client.MetadataResource.addMetadataToPackages(&packages, tt.page, tt.options, versionRange)
			require.Equal(t, tt.error, err)

			require.Equal(t, tt.wantPkg, packages)
		})
	}
}
