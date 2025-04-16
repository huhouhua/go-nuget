// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

// https://api.nuget.org/v3-flatcontainer/newtonsoft.json/index.json
func TestPackageResource_ListAllVersions(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/v3-flatcontainer/newtonsoft.json/index.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/list_all_versions.json")
	})

	want := []*NuGetVersion{{
		Version: &Version{
			Major:    6,
			Minor:    0,
			Build:    1,
			Revision: 0,
		},
		SemanticVersion: &SemanticVersion{
			releaseLabels: []string{"beta1"},
			metadata:      "",
			Major:         6,
			Minor:         0,
			Patch:         1,
		},
		Revision:        0,
		OriginalVersion: "6.0.1-beta1",
	}, {
		Version: &Version{
			Major:    6,
			Minor:    0,
			Build:    1,
			Revision: 0,
		},
		SemanticVersion: &SemanticVersion{
			releaseLabels: nil,
			metadata:      "",
			Major:         6,
			Minor:         0,
			Patch:         1,
		},
		Revision:        0,
		OriginalVersion: "6.0.1",
	}}

	b, resp, err := client.FindPackageResource.ListAllVersions("newtonsoft.json", nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, b)
}

func TestPackageResource_GetDependencyInfo(t *testing.T) {
	mux, client := setup(t)
	url := fmt.Sprintf("/v3-flatcontainer/testdependency/%s/testdependency.nuspec", PathEscape("1.0.0"))
	mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/testDependency.nuspec")
	})

	want := &PackageDependencyInfo{
		PackageIdentity: &PackageIdentity{
			Id: "TestDependency",
			Version: &NuGetVersion{
				SemanticVersion: &SemanticVersion{
					Major: 1,
					Minor: 0,
					Patch: 0,
				},
				Version: &Version{
					Major:    1,
					Minor:    0,
					Build:    0,
					Revision: 0,
				},
				Revision:        0,
				OriginalVersion: "1.0.0",
			},
		},
		DependencyGroups: []*PackageDependencyGroup{
			{
				TargetFramework: ".NETFramework4.8",
				Packages: []*Dependency{
					{
						Id:         "Newtonsoft.Json",
						VersionRaw: "12.0.3",
						ExcludeRaw: "Build,Analyzers",
						Version: &NuGetVersion{
							SemanticVersion: &SemanticVersion{
								Major: 12,
								Minor: 0,
								Patch: 3,
							},
							Version: &Version{
								Major:    12,
								Minor:    0,
								Build:    3,
								Revision: 0,
							},
							Revision:        0,
							OriginalVersion: "12.0.3",
						},
						Exclude: []string{"Build", "Analyzers"},
					},
					{
						Id:         "Microsoft.Extensions.Logging",
						VersionRaw: "5.0.0",
						Version: &NuGetVersion{
							SemanticVersion: &SemanticVersion{
								Major: 5,
								Minor: 0,
								Patch: 0,
							},
							Version: &Version{
								Major:    5,
								Minor:    0,
								Build:    0,
								Revision: 0,
							},
							Revision:        0,
							OriginalVersion: "5.0.0",
						},
					},
				},
			},
			{
				TargetFramework: ".NETStandard2.0",
				Packages: []*Dependency{
					{
						Id:         "Newtonsoft.Json",
						VersionRaw: "12.0.3",
						ExcludeRaw: "Build,Analyzers",
						Version: &NuGetVersion{
							SemanticVersion: &SemanticVersion{
								Major: 12,
								Minor: 0,
								Patch: 3,
							},
							Version: &Version{
								Major:    12,
								Minor:    0,
								Build:    3,
								Revision: 0,
							},
							Revision:        0,
							OriginalVersion: "12.0.3",
						},
						Exclude: []string{"Build", "Analyzers"},
					},
				},
			},
		},
		FrameworkReferenceGroups: []*FrameworkSpecificGroup{
			{
				Items:           []string{"", "System.Net.Http"},
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

func TestPackageResource_CopyNupkgToStream(t *testing.T) {
	mux, client := setup(t)
	opt := &CopyNupkgOptions{
		Version: "6.0.1-beta1",
		Writer:  &bytes.Buffer{},
	}
	id := "newtonsoft.json"
	packageId, version := PathEscape(id), PathEscape(opt.Version)

	url := fmt.Sprintf("/v3-flatcontainer/%s/%s/%s.%s.nupkg",
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
