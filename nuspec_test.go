// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReaderNupkg(t *testing.T) {
	nupkgPath := "testdata/test.1.0.0.nupkg"

	file, err := os.Open(nupkgPath)
	require.NoError(t, err, "open file %s failed %v", nupkgPath, err)
	t.Cleanup(func() {
		_ = file.Close()
	})

	var dest io.Writer = &bytes.Buffer{}
	_, err = io.Copy(dest, file)
	require.NoError(t, err)

	reader, err := NewPackageArchiveReader(dest)
	require.NoError(t, err, "Failed to parse nuget package archive")

	spec, err := reader.Nuspec()
	require.NoErrorf(t, err, "Failed Get nuspec file content: %v", err)

	want := &Nuspec{
		XMLName: xml.Name{
			Space: "http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd",
			Local: "package",
		},
		Metadata: &Metadata{
			PackageInfo: PackageInfo{
				ID:                       "MyTestLibrary",
				Version:                  "1.0.0",
				Authors:                  "Kevin Berger",
				Owners:                   "Kevin Berger",
				RequireLicenseAcceptance: false,
				License:                  "MIT",
				ProjectURL:               "https://github.com/huhouhua/go-nuget",
				IconUrl:                  "https://raw.githubusercontent.com/huhouhua/go-nuget/main/icon.png",
				Description:              "A fantastic library that solves all your problems.",
				Summary:                  "Lightweight helper for building apps",
				ReleaseNotes:             "Initial stable release",
				Copyright:                "Copyright © 2025 Kevin Berger",
				Tags:                     "utility helper tools awesome",
				Language:                 "en-US",
				Repository: &Repository{
					Type:   "git",
					URL:    "https://github.com/huhouhua/go-nuget.git",
					Branch: "main",
					Commit: "abc123",
				},
			},
			Dependencies: &Dependencies{
				Groups: []*DependenciesGroup{
					{
						TargetFramework: ".NETFramework4.8",
						Dependencies: []*Dependency{
							{
								Id:         "Newtonsoft.Json",
								VersionRaw: "12.0.3",
								ExcludeRaw: "Build,Analyzers",
							},
							{
								Id:         "Microsoft.Extensions.Logging",
								VersionRaw: "5.0.0",
							},
						},
					},
					{
						TargetFramework: ".NETCoreApp3.1",
						Dependencies: []*Dependency{
							{
								Id:         "Newtonsoft.Json",
								VersionRaw: "12.0.3",
								ExcludeRaw: "Build,Analyzers",
							},
						},
					},
					{
						TargetFramework: "net5.0",
						Dependencies: []*Dependency{
							{
								Id:         "Newtonsoft.Json",
								VersionRaw: "12.0.3",
								ExcludeRaw: "Build,Analyzers",
							},
						},
					},
					{
						TargetFramework: ".NETStandard2.0",
						Dependencies: []*Dependency{
							{
								Id:         "Newtonsoft.Json",
								VersionRaw: "12.0.3",
								ExcludeRaw: "Build,Analyzers",
							},
						},
					},
				},
				Dependency: []*Dependency{
					{
						Id:         "Castle.Core.AsyncInterceptor",
						VersionRaw: "2.1.0",
						ExcludeRaw: "Build,Analyzers",
					},
					{
						Id:         "JetBrains.Annotations",
						VersionRaw: "2024.3.0",
						ExcludeRaw: "Build,Analyzers",
					},
				},
			},
			FrameworkAssemblies: &FrameworkAssemblies{
				FrameworkAssembly: []*FrameworkAssembly{
					{
						AssemblyName: []string{
							"System.Net.Http",
						},
						TargetFramework: ".NETFramework4.8",
					},
				},
			},
			References: &References{
				Groups: []*ReferenceGroup{
					{
						TargetFramework: ".NETStandard2.0",
						References: []*Reference{
							{
								File: "MyTestLibrary.dll",
							},
						},
					},
				},
			},
		},
	}

	require.Equal(t, want, spec)
}
