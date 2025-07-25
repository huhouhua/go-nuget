// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	meta1 "github.com/huhouhua/go-nuget/internal/meta"

	nugetVersion "github.com/huhouhua/go-nuget/version"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
)

func TestPackageMetadataResource_ListMetadata(t *testing.T) {
	mux, client := setup(t, index_Baget)

	baseURL := client.getResourceURL(RegistrationsBaseURL)
	u := fmt.Sprintf("%s/gitlabapiclient/index.json", baseURL.Path)

	mux.HandleFunc(u, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/medata.json")
	})

	versionrange1203, err := nugetVersion.ParseRange("[12.0.3, )")
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
			ReportAbuseURL: reportUrl,
			Authors:        "nmklotas",
			SearchMetadata: &SearchMetadata{
				PackageId: "GitLabApiClient",
				Version:   "1.8.1-beta.5",
				DependencySets: []*meta1.PackageDependencyGroup{
					{
						TargetFramework: "net48",
						Packages: []*meta1.Dependency{
							{
								Id:              "Newtonsoft.Json",
								VersionRangeRaw: "[12.0.3, )",
								VersionRange:    versionrange1203,
							},
						},
					},
					{
						TargetFramework: "netcoreapp3.1",
						Packages: []*meta1.Dependency{
							{
								Id:              "Newtonsoft.Json",
								VersionRangeRaw: "[12.0.3, )",
								VersionRange:    versionrange1203,
							},
						},
					},
					{
						TargetFramework: "net5.0",
						Packages: []*meta1.Dependency{
							{
								Id:              "Newtonsoft.Json",
								VersionRangeRaw: "[12.0.3, )",
								VersionRange:    versionrange1203,
							},
						},
					},
					{
						TargetFramework: "netstandard2.0",
						Packages: []*meta1.Dependency{
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
				LicenseURL:               "https://licenses.nuget.org/MIT",
				ProjectURL:               "https://github.com/nmklotas/GitLabApiClient",
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
	tests := []struct {
		name        string
		id          string
		version     string
		optionsFunc []RequestOptionFunc
		handleFunc  func(http.ResponseWriter, *http.Request)
		wantFunc    func(client *Client, err error, meta *PackageSearchMetadataRegistration)
	}{
		{
			name:    "invalid version",
			id:      "json",
			version: "x.0.0",
			wantFunc: func(client *Client, err error, meta *PackageSearchMetadataRegistration) {
				wantErr := errors.New("invalid semantic version")
				require.Equal(t, wantErr, err)
			},
		},
		{
			name:    "parse id return error",
			version: "1.8.1-beta.5",
			wantFunc: func(client *Client, err error, meta *PackageSearchMetadataRegistration) {
				wantErr := errors.New("id is empty")
				require.Equal(t, wantErr, err)
			},
		},
		{
			name:    "new request return error",
			id:      "gitlabapiclient",
			version: "1.8.1-beta.5",
			optionsFunc: []RequestOptionFunc{
				func(request *retryablehttp.Request) error {
					return errors.New("new request fail")
				},
			},
			wantFunc: func(client *Client, err error, meta *PackageSearchMetadataRegistration) {
				wantErr := errors.New("new request fail")
				require.Equal(t, wantErr, err)
			},
		},
		{
			name:    "status bad gateway api return error",
			id:      "newtonsoft.json",
			version: "1.8.5",
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				_, err := fmt.Fprint(w, `{"msg":"gateway fail"}`)
				require.NoError(t, err)
			},
			wantFunc: func(client *Client, err error, meta *PackageSearchMetadataRegistration) {
				var errResp *ErrorResponse
				require.True(t, errors.As(err, &errResp), "TestPackageMetadataResource_GetMetadata returns an error")
				wantErr := `{msg: gateway fail}`
				require.Equal(t, wantErr, errResp.Message)
			},
		},
		{
			name:    "not find return error",
			id:      "gitlabapiclient",
			version: "1.8.1-beta.5",
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodGet)

				tmpDir := t.TempDir()
				index := registrationIndex{
					Items: []*registrationPage{},
				}
				testData, err := json.Marshal(index)
				require.NoError(t, err)

				fileUrl := filepath.Join(tmpDir, "medata.json")
				createFile(t, fileUrl, string(testData))
				mustWriteHTTPResponse(t, w, fileUrl)
			},
			wantFunc: func(client *Client, err error, meta *PackageSearchMetadataRegistration) {
				wantErr := errors.New("gitlabapiclient 1.8.1-beta.5 not find")
				require.Equal(t, wantErr, err)
			},
		},
		{
			name:    "registration items has nil return error",
			id:      "gitlabapiclient",
			version: "1.8.1-beta.5",
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodGet)

				tmpDir := t.TempDir()
				index := registrationIndex{
					Items: []*registrationPage{
						nil,
					},
				}
				testData, err := json.Marshal(index)
				require.NoError(t, err)

				fileUrl := filepath.Join(tmpDir, "medata.json")
				createFile(t, fileUrl, string(testData))
				mustWriteHTTPResponse(t, w, fileUrl)
			},
			wantFunc: func(client *Client, err error, meta *PackageSearchMetadataRegistration) {
				if !strings.HasPrefix(err.Error(), "invalid") {
					t.Fatalf("TestPackageMetadataResource_GetMetadata return error %s want prefix invalid", err.Error())
				}
			},
		},
		{
			name:    "add metadata to packages return error",
			id:      "gitlabapiclient",
			version: "1.8.1-beta.5",
			handleFunc: func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodGet)

				tmpDir := t.TempDir()
				index := registrationIndex{
					Items: []*registrationPage{
						{
							Lower: "^1.0.0",
						},
					},
				}
				testData, err := json.Marshal(index)
				require.NoError(t, err)

				fileUrl := filepath.Join(tmpDir, "medata.json")
				createFile(t, fileUrl, string(testData))
				mustWriteHTTPResponse(t, w, fileUrl)
			},
			wantFunc: func(client *Client, err error, meta *PackageSearchMetadataRegistration) {
				wantErr := errors.New("invalid semantic version")
				require.Equal(t, wantErr, err)
			},
		},
		{
			name:    "query return success",
			id:      "gitlabapiclient",
			version: "1.8.1-beta.5",
			wantFunc: func(client *Client, _ error, meta *PackageSearchMetadataRegistration) {
				versionrange1203, err := nugetVersion.ParseRange("[12.0.3, )")
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
					ReportAbuseURL: reportUrl,
					Authors:        "nmklotas",
					SearchMetadata: &SearchMetadata{
						PackageId: "GitLabApiClient",
						Version:   "1.8.1-beta.5",
						DependencySets: []*meta1.PackageDependencyGroup{
							{
								TargetFramework: "net48",
								Packages: []*meta1.Dependency{
									{
										Id:              "Newtonsoft.Json",
										VersionRangeRaw: "[12.0.3, )",
										VersionRange:    versionrange1203,
									},
								},
							},
							{
								TargetFramework: "netcoreapp3.1",
								Packages: []*meta1.Dependency{
									{
										Id:              "Newtonsoft.Json",
										VersionRangeRaw: "[12.0.3, )",
										VersionRange:    versionrange1203,
									},
								},
							},
							{
								TargetFramework: "net5.0",
								Packages: []*meta1.Dependency{
									{
										Id:              "Newtonsoft.Json",
										VersionRangeRaw: "[12.0.3, )",
										VersionRange:    versionrange1203,
									},
								},
							},
							{
								TargetFramework: "netstandard2.0",
								Packages: []*meta1.Dependency{
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
						LicenseURL:               "https://licenses.nuget.org/MIT",
						ProjectURL:               "https://github.com/nmklotas/GitLabApiClient",
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
				require.Equal(t, want, meta)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, client := setup(t, index_Baget)

			baseURL := client.getResourceURL(RegistrationsBaseURL)
			u := fmt.Sprintf("%s/%s/index.json", baseURL.Path, tt.id)
			handler := func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodGet)
				mustWriteHTTPResponse(t, w, "testdata/medata.json")
			}
			if tt.handleFunc != nil {
				handler = tt.handleFunc
			}
			mux.HandleFunc(u, handler)
			b, _, err := client.MetadataResource.GetMetadata(tt.id, tt.version, tt.optionsFunc...)
			if tt.wantFunc != nil {
				tt.wantFunc(client, err, b)
			}
		})
	}
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
		wantIdentity := &meta1.PackageIdentity{
			Id:      input.SearchMetadata.PackageId,
			Version: nugetVersion.NewVersionFrom(1, 0, 0, "beta", ""),
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
		wantErr := errors.New("invalid semantic version")
		_, err := inputErr.Identity()
		require.Equal(t, wantErr, err)
	})
}

func TestParseAndReplaceURL(t *testing.T) {
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
			actual, err := parseAndReplaceURL(tt.urlTemplate, tt.replacements)
			require.Equal(t, tt.error, err)
			require.Equal(t, tt.want, actual)
		})
	}
}

func TestWithReportAbuseURL(t *testing.T) {
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
			err := WithReportAbuseURL(tt.urlTemplate)(tt.metadata)
			require.Equal(t, tt.error, err)
			require.Equal(t, tt.want, tt.metadata.ReportAbuseURL)
		})
	}
}

func TestWithPackageDetailsURL(t *testing.T) {
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
			err := WithPackageDetailsURL(tt.urlTemplate)(tt.metadata)
			require.Equal(t, tt.error, err)
			require.Equal(t, tt.want, tt.metadata.PackageDetailsURL)
		})
	}
}

func TestWithReadmeFileURL(t *testing.T) {
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
			err := WithReadmeFileURL(tt.urlTemplate)(tt.metadata)
			require.Equal(t, tt.error, err)
			require.Equal(t, tt.want, tt.metadata.ReadmeFileURL)
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
			WithReportAbuseURL(reportAbuseUrlTemplate),
			WithPackageDetailsURL(detailsUrlTemplate),
			WithReadmeFileURL(readmeUrlTemplate),
		)

		require.NoError(t, err)
		require.Equal(t, "https://example.com/packages/testpackage/1.0.0/ReportAbuse", metadata.ReportAbuseURL.String())
		require.Equal(
			t,
			"https://example.com/packages/testpackage/1.0.0?_src=template",
			metadata.PackageDetailsURL.String(),
		)
		require.Equal(
			t,
			"https://example.com/v3-flatcontainer/testpackage/1.0.0/readme",
			metadata.ReadmeFileURL.String(),
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
			WithReportAbuseURL(invalidUrlTemplate),
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
	versionRange, err := nugetVersion.ParseRange("[1.5.0, )")
	require.NoError(t, err)

	emptyPkg := make([]*PackageSearchMetadataRegistration, 0)

	tests := []struct {
		name             string
		page             *registrationPage
		options          *ListMetadataOptions
		configClientFunc func(client *Client)
		wantPkgFunc      func(baseURL string) []*PackageSearchMetadataRegistration
		error            error
	}{
		{
			name: "valid package in range return success",
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
			wantPkgFunc: func(baseURL string) []*PackageSearchMetadataRegistration {
				return []*PackageSearchMetadataRegistration{
					{
						SearchMetadata: &SearchMetadata{
							PackageId: "TestPackage",
							Version:   "1.5.0",
							IsListed:  true,
						},
						PackageDetailsURL: createUrl(
							t,
							fmt.Sprintf("%s/packages/testpackage/1.5.0?_src=template", baseURL),
						),
						ReadmeFileURL: createUrl(
							t,
							fmt.Sprintf("%s/v3-flatcontainer/testpackage/1.5.0/readme", baseURL),
						),
						ReportAbuseURL: createUrl(
							t,
							fmt.Sprintf("%s/packages/testpackage/1.5.0/ReportAbuse", baseURL),
						),
					},
				}
			},
			error: nil,
		},
		{
			name: "invalid lower version in page return error",
			page: &registrationPage{
				Lower: "invalid-lower-version",
				Upper: "2.0.0",
			},
			options: &ListMetadataOptions{
				IncludePrerelease: true,
				IncludeUnlisted:   true,
			},
			wantPkgFunc: func(_ string) []*PackageSearchMetadataRegistration {
				return emptyPkg
			},
			error: errors.New("invalid semantic version"),
		},
		{
			name: "invalid upper version in page return error",
			page: &registrationPage{
				Lower: "1.0.0",
				Upper: "invalid-upper-version",
			},
			options: &ListMetadataOptions{
				IncludePrerelease: true,
				IncludeUnlisted:   true,
			},
			wantPkgFunc: func(_ string) []*PackageSearchMetadataRegistration {
				return emptyPkg
			},
			error: errors.New("invalid semantic version"),
		},
		{
			name: "version out of range",
			page: &registrationPage{
				Lower: "3.0.0",
				Upper: "4.0.0",
			},
			options: &ListMetadataOptions{
				IncludePrerelease: true,
				IncludeUnlisted:   true,
			},
			wantPkgFunc: func(_ string) []*PackageSearchMetadataRegistration {
				return emptyPkg
			},
			error: nil,
		},
		{
			name: "parse template url error",
			configClientFunc: func(client *Client) {
				invalidUrlTemplate := createUrl(t, "https://example.com/packages/{id}/{version}")
				invalidUrlTemplate.Path = invalidUrlTemplate.Path + "%%ReportAbuse"
				client.serviceURLs[ReportAbuseURLTemplate] = invalidUrlTemplate
			},
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
								DependencySets: []*meta1.PackageDependencyGroup{
									{
										TargetFramework: "net48",
										Packages: []*meta1.Dependency{
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
			wantPkgFunc: func(_ string) []*PackageSearchMetadataRegistration {
				return emptyPkg
			},
			error: &url.Error{
				Op:  "parse",
				URL: "https://example.com/packages/testpackage/1.5.0%%ReportAbuse",
				Err: url.EscapeError("%%R"),
			},
		},
		{
			name: "invalid version in catalog entry return error",
			page: &registrationPage{
				Lower: "3.0.0",
				Upper: "4.0.0",
				Items: []*registrationLeafItem{
					{
						CatalogEntry: &PackageSearchMetadataRegistration{
							SearchMetadata: &SearchMetadata{
								Version: "^4.0.0",
							},
						},
					},
				},
			},
			options: &ListMetadataOptions{
				IncludePrerelease: true,
				IncludeUnlisted:   true,
			},
			wantPkgFunc: func(_ string) []*PackageSearchMetadataRegistration {
				return emptyPkg
			},
			error: errors.New("invalid semantic version"),
		},
		{
			name: "includeUnlisted is false and package is unlisted return success",
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
			wantPkgFunc: func(_ string) []*PackageSearchMetadataRegistration {
				return emptyPkg
			},
			error: nil,
		},
		{
			name: "parse dependencySets return success",
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
								DependencySets: []*meta1.PackageDependencyGroup{
									{
										TargetFramework: "net48",
										Packages: []*meta1.Dependency{
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
			wantPkgFunc: func(baseURL string) []*PackageSearchMetadataRegistration {
				return []*PackageSearchMetadataRegistration{
					{
						SearchMetadata: &SearchMetadata{
							PackageId: "TestPackage",
							Version:   "1.5.0",
							DependencySets: []*meta1.PackageDependencyGroup{
								{
									TargetFramework: "net48",
									Packages: []*meta1.Dependency{
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
						PackageDetailsURL: createUrl(
							t,
							fmt.Sprintf("%s/packages/testpackage/1.5.0?_src=template", baseURL),
						),
						ReadmeFileURL: createUrl(
							t,
							fmt.Sprintf("%s/v3-flatcontainer/testpackage/1.5.0/readme", baseURL),
						),
						ReportAbuseURL: createUrl(
							t,
							fmt.Sprintf("%s/packages/testpackage/1.5.0/ReportAbuse", baseURL),
						),
					},
				}
			},
		},
		{
			name: "parse dependencySets return error",
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
								DependencySets: []*meta1.PackageDependencyGroup{
									{
										TargetFramework: "net48",
										Packages: []*meta1.Dependency{
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
			error: errors.New("'1.0.0%' is not a valid version string"),
			wantPkgFunc: func(_ string) []*PackageSearchMetadataRegistration {
				return emptyPkg
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packages := make([]*PackageSearchMetadataRegistration, 0)
			_, client := setup(t, index_V3)
			if tt.configClientFunc != nil {
				tt.configClientFunc(client)
			}
			require.NotNil(t, client)
			err = client.MetadataResource.addMetadataToPackages(&packages, tt.page, tt.options, versionRange)
			require.Equal(t, tt.error, err)

			require.Equal(t, tt.wantPkgFunc(client.baseURL.String()), packages)
		})
	}
}
