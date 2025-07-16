// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	nugetVersion "github.com/huhouhua/go-nuget/version"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
)

func TestPackageResource_ListAllVersions(t *testing.T) {
	mux, client := setup(t, index_V3)

	baseURL := client.getResourceURL(PackageBaseAddress)
	u := fmt.Sprintf("%s/newtonsoft.json/index.json", baseURL.Path)

	mux.HandleFunc(u, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/list_all_versions.json")
	})

	want := []*nugetVersion.Version{
		nugetVersion.NewVersionFrom(6, 0, 1, "beta1", ""),
		nugetVersion.NewVersionFrom(6, 0, 1, "", ""),
	}

	b, resp, err := client.FindPackageResource.ListAllVersions("newtonsoft.json", nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, b)
}

func TestPackageResource_ListAllVersions_ErrorScenarios(t *testing.T) {
	tmpDir := t.TempDir()
	tests := []struct {
		name        string
		id          string
		handleFunc  func(client *Client, mux *http.ServeMux)
		optionsFunc []RequestOptionFunc
		error       error
	}{
		{
			name:  "parse id return error",
			error: errors.New("id is empty"),
		},
		{
			name: "new request return error",
			id:   "newtonsoft.json",
			optionsFunc: []RequestOptionFunc{
				func(request *retryablehttp.Request) error {
					return errors.New("new request fail")
				},
			},
			error: errors.New("new request fail"),
		},
		{
			name: "api interface does not exist return error",
			id:   "newtonsoft.json",
			handleFunc: func(client *Client, mux *http.ServeMux) {
			},
			error: errors.New("404 Not Found"),
		},
		{
			name: "version parse return error",
			id:   "newtonsoft.json",
			handleFunc: func(client *Client, mux *http.ServeMux) {
				baseURL := client.getResourceURL(PackageBaseAddress)
				u := fmt.Sprintf("%s/newtonsoft.json/index.json", baseURL.Path)

				mux.HandleFunc(u, func(w http.ResponseWriter, r *http.Request) {
					testMethod(t, r, http.MethodGet)

					testDataUrl := "testdata/list_all_versions.json"
					data, err := os.ReadFile(testDataUrl)
					require.NoError(t, err)

					var version struct {
						Versions []string `json:"versions"`
					}
					err = json.Unmarshal(data, &version)
					require.NoError(t, err)

					for i := 0; i < len(version.Versions); i++ {
						version.Versions[i] = "^0.0.1"
					}
					testData, err := json.Marshal(version)
					require.NoError(t, err)

					fileUrl := filepath.Join(tmpDir, "list_all_versions.json")
					createFile(t, fileUrl, string(testData))
					mustWriteHTTPResponse(t, w, fileUrl)
				})
			},
			error: errors.New("invalid semantic version"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, client := setup(t, index_V3)
			if tt.handleFunc != nil {
				tt.handleFunc(client, mux)
			} else {
				baseURL := client.getResourceURL(PackageBaseAddress)
				u := fmt.Sprintf("%s/%s/index.json", baseURL.Path, PathEscape(tt.id))

				mux.HandleFunc(u, func(w http.ResponseWriter, r *http.Request) {
					testMethod(t, r, http.MethodGet)
					mustWriteHTTPResponse(t, w, "testdata/list_all_versions.json")
				})
			}
			_, _, err := client.FindPackageResource.ListAllVersions(tt.id, tt.optionsFunc...)
			require.Equal(t, tt.error, err)
		})
	}
}

func TestPackageResource_GetDependencyInfo(t *testing.T) {
	mux, client := setup(t, index_V3)

	baseURL := client.getResourceURL(PackageBaseAddress)
	u := fmt.Sprintf("%s/testdependency/%s/testdependency.nuspec", baseURL.Path, PathEscape("1.0.0"))

	mux.HandleFunc(u, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/testDependency.nuspec")
	})

	versionRange1203, err := nugetVersion.ParseRange("12.0.3")
	require.NoError(t, err)

	versionRange500, err := nugetVersion.ParseRange("5.0.0")
	require.NoError(t, err)

	want := &PackageDependencyInfo{
		PackageIdentity: &PackageIdentity{
			Id:      "TestDependency",
			Version: nugetVersion.NewVersionFrom(1, 0, 0, "", ""),
		},
		DependencyGroups: []*PackageDependencyGroup{
			{
				TargetFramework: ".NETFramework4.8",
				Packages: []*Dependency{
					{
						Id:           "Newtonsoft.Json",
						VersionRaw:   "12.0.3",
						ExcludeRaw:   "Build,Analyzers",
						VersionRange: versionRange1203,
						//Version:    &NuGetVersion{semver.New(12, 0, 3, "", "")},
						Exclude: []string{"Build", "Analyzers"},
					},
					{
						Id:           "Microsoft.Extensions.Logging",
						VersionRaw:   "5.0.0",
						VersionRange: versionRange500,
						//Version:    &NuGetVersion{semver.New(5, 0, 0, "", "")},
					},
				},
			},
			{
				TargetFramework: ".NETStandard2.0",
				Packages: []*Dependency{
					{
						Id:           "Newtonsoft.Json",
						VersionRaw:   "12.0.3",
						ExcludeRaw:   "Build,Analyzers",
						VersionRange: versionRange1203,
						//Version:    &NuGetVersion{semver.New(12, 0, 3, "", "")},
						Exclude: []string{"Build", "Analyzers"},
					},
				},
			},
		},
		FrameworkReferenceGroups: []*FrameworkSpecificGroup{
			{
				Items:           []string{"System.Net.Http"},
				HasEmptyFolder:  false,
				TargetFramework: ".NETFramework4.8",
			},
		},
	}
	b, resp, err := client.FindPackageResource.GetDependencyInfo("testdependency", "1.0.0")
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, b)
}

func TestPackageResource_GetDependencyInfo_ErrorScenarios(t *testing.T) {
	tmpDir := t.TempDir()
	tests := []struct {
		name        string
		id          string
		version     string
		handleFunc  func(client *Client, mux *http.ServeMux)
		optionsFunc []RequestOptionFunc
		error       error
	}{
		{
			name:  "parse id return error",
			error: errors.New("id is empty"),
		},
		{
			name: "new request return error",
			id:   "newtonsoft.json",
			optionsFunc: []RequestOptionFunc{
				func(request *retryablehttp.Request) error {
					return errors.New("new request fail")
				},
			},
			error: errors.New("new request fail"),
		},
		{
			name: "api interface does not exist return error",
			id:   "newtonsoft.json",
			handleFunc: func(client *Client, mux *http.ServeMux) {
			},
			error: errors.New("404 Not Found"),
		},
		{
			name:    "version parse return error",
			id:      "testdependency",
			version: "1.0.0",
			handleFunc: func(client *Client, mux *http.ServeMux) {
				baseURL := client.getResourceURL(PackageBaseAddress)
				u := fmt.Sprintf("%s/testdependency/%s/testdependency.nuspec", baseURL.Path, PathEscape("1.0.0"))

				mux.HandleFunc(u, func(w http.ResponseWriter, r *http.Request) {
					testMethod(t, r, http.MethodGet)

					testDataUrl := "testdata/testDependency.nuspec"
					file, err := os.Open(testDataUrl)
					require.NoError(t, err)

					t.Cleanup(func() {
						_ = file.Close()
					})
					var nuspec Nuspec
					err = xml.NewDecoder(file).Decode(&nuspec)
					require.NoError(t, err)

					for _, assembly := range nuspec.Metadata.FrameworkAssemblies.FrameworkAssembly {
						assembly.AssemblyName = nil
					}
					testData, err := xml.Marshal(nuspec)
					require.NoError(t, err)

					fileUrl := filepath.Join(tmpDir, "testDependency.nuspec")
					createFile(t, fileUrl, string(testData))
					mustWriteHTTPResponse(t, w, fileUrl)
				})
			},
			error: errors.New("items cannot be nil"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, client := setup(t, index_V3)
			if tt.handleFunc != nil {
				tt.handleFunc(client, mux)
			} else {
				baseURL := client.getResourceURL(PackageBaseAddress)
				u := fmt.Sprintf("%s/testdependency/%s/testdependency.nuspec", baseURL.Path, PathEscape(tt.version))

				mux.HandleFunc(u, func(w http.ResponseWriter, r *http.Request) {
					testMethod(t, r, http.MethodGet)
					mustWriteHTTPResponse(t, w, "testdata/testDependency.nuspec")
				})
			}
			_, _, err := client.FindPackageResource.GetDependencyInfo(tt.id, tt.version, tt.optionsFunc...)
			require.Equal(t, tt.error, err)
		})
	}
}

func TestPackageResource_CopyNupkgToStream(t *testing.T) {
	mux, client := setup(t, index_V3)
	opt := &CopyNupkgOptions{
		Version: "6.0.1-beta1",
		Writer:  &bytes.Buffer{},
	}
	id := "newtonsoft.json"
	packageId, version := PathEscape(id), PathEscape(opt.Version)
	baseURL := client.getResourceURL(PackageBaseAddress)
	url := fmt.Sprintf("%s/%s/%s/%s.%s.nupkg",
		baseURL.Path,
		packageId,
		version,
		packageId,
		version)

	mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/newtonsoft.json.6.0.1-beta1.nupkg")
	})

	resp, err := client.FindPackageResource.CopyNupkgToStream(id, opt)
	require.NoError(t, err)
	require.NotNil(t, resp)

	reader, err := NewPackageArchiveReader(opt.Writer)
	require.NoError(t, err)

	spec, err := reader.Nuspec()
	require.NoError(t, err)
	require.NotNil(t, spec)
}

func TestPackageResource_CopyNupkgToStream_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		opt         *CopyNupkgOptions
		handleFunc  func(client *Client, mux *http.ServeMux)
		optionsFunc []RequestOptionFunc
		error       error
	}{
		{
			name:  "parse id return error",
			error: errors.New("id is empty"),
		},
		{
			name: "new request return error",
			id:   "newtonsoft.json",
			opt: &CopyNupkgOptions{
				Version: "1.0.0",
			},
			optionsFunc: []RequestOptionFunc{
				func(request *retryablehttp.Request) error {
					return errors.New("new request fail")
				},
			},
			error: errors.New("new request fail"),
		},
		{
			name: "api interface does not exist return error",
			id:   "newtonsoft.json",
			opt: &CopyNupkgOptions{
				Version: "1.0.0",
			},
			handleFunc: func(client *Client, mux *http.ServeMux) {
			},
			error: errors.New("404 Not Found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, client := setup(t, index_V3)
			if tt.handleFunc != nil {
				tt.handleFunc(client, mux)
			} else if tt.opt != nil {
				packageId, version := PathEscape(tt.id), PathEscape(tt.opt.Version)
				baseURL := client.getResourceURL(PackageBaseAddress)
				url := fmt.Sprintf("%s/%s/%s/%s.%s.nupkg",
					baseURL.Path,
					packageId,
					version,
					packageId,
					version)

				mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
					testMethod(t, r, http.MethodGet)
					mustWriteHTTPResponse(t, w, "testdata/newtonsoft.json.6.0.1-beta1.nupkg")
				})
			}
			_, err := client.FindPackageResource.CopyNupkgToStream(tt.id, tt.opt, tt.optionsFunc...)
			require.Equal(t, tt.error, err)
		})
	}
}
