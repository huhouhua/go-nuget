// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"io/fs"
	"os"
	"syscall"
	"testing"

	"github.com/huhouhua/go-nuget/internal/meta"

	"github.com/stretchr/testify/require"
)

func TestReaderNupkg(t *testing.T) {
	nupkgPath := "testdata/test.1.0.0.nupkg"

	file, err := os.Open(nupkgPath)
	require.NoError(t, err, "open file %s failed %v", nupkgPath, err)
	t.Cleanup(func() {
		_ = file.Close()
	})

	reader, err := NewPackageArchiveReader(file)
	require.NoError(t, err, "Failed to parse nuget package archive")
	_, err = reader.Nuspec()
	require.NoErrorf(t, err, "Failed Get nuspec file content: %v", err)
	spec, _ := reader.Nuspec()

	want := &meta.Nuspec{
		Xmlns: "http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd",
		XMLName: xml.Name{
			Space: "http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd",
			Local: "package",
		},
		Metadata: &meta.Metadata{
			PackageInfo: meta.PackageInfo{
				ID:                       "MyTestLibrary",
				Version:                  "1.0.0",
				Authors:                  "Kevin Berger",
				Owners:                   "Kevin Berger",
				RequireLicenseAcceptance: false,
				License: &meta.LicenseMetadata{
					Type:  "expression",
					Value: "MIT",
				},
				ProjectURL:   "https://github.com/huhouhua/go-nuget",
				IconURL:      "https://raw.githubusercontent.com/huhouhua/go-nuget/main/icon.png",
				Description:  "A fantastic library that solves all your problems.",
				Summary:      "Lightweight helper for building apps",
				ReleaseNotes: "Initial stable release",
				Copyright:    "Copyright © 2025 Kevin Berger",
				Tags:         "utility helper tools awesome",
				Language:     "en-US",
				Repository: &meta.RepositoryMetadata{
					Type:   "git",
					URL:    "https://github.com/huhouhua/go-nuget.git",
					Branch: "main",
					Commit: "abc123",
				},
			},
			Dependencies: &meta.Dependencies{
				Groups: []*meta.DependenciesGroup{
					{
						TargetFramework: ".NETFramework4.8",
						Dependencies: []*meta.Dependency{
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
						Dependencies: []*meta.Dependency{
							{
								Id:         "Newtonsoft.Json",
								VersionRaw: "12.0.3",
								ExcludeRaw: "Build,Analyzers",
							},
						},
					},
					{
						TargetFramework: "net5.0",
						Dependencies: []*meta.Dependency{
							{
								Id:         "Newtonsoft.Json",
								VersionRaw: "12.0.3",
								ExcludeRaw: "Build,Analyzers",
							},
						},
					},
					{
						TargetFramework: ".NETStandard2.0",
						Dependencies: []*meta.Dependency{
							{
								Id:         "Newtonsoft.Json",
								VersionRaw: "12.0.3",
								ExcludeRaw: "Build,Analyzers",
							},
						},
					},
				},
				Dependency: []*meta.Dependency{
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
			FrameworkAssemblies: &meta.FrameworkAssemblies{
				FrameworkAssembly: []*meta.FrameworkAssembly{
					{
						AssemblyName: []string{
							"System.Net.Http",
						},
						TargetFramework: ".NETFramework4.8",
					},
				},
			},
			References: &meta.References{
				Groups: []*meta.ReferenceGroup{
					{
						TargetFramework: ".NETStandard2.0",
						References: []*meta.Reference{
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
	for _, f := range reader.GetFilesFromDir("testLibrary") {
		require.NotEmpty(t, f.Name)
	}
}

func TestReaderFullPackage(t *testing.T) {
	nupkgPath := "testdata/my_package.nupkg"

	file, err := os.Open(nupkgPath)
	require.NoError(t, err, "open file %s failed %v", nupkgPath, err)
	t.Cleanup(func() {
		_ = file.Close()
	})

	reader, err := NewPackageArchiveReader(file)
	require.NoError(t, err, "Failed to parse nuget package archive")
	_, err = reader.Nuspec()
	require.NoErrorf(t, err, "Failed Get nuspec file content: %v", err)
	spec, _ := reader.Nuspec()

	want := &meta.Nuspec{
		Xmlns: "http://schemas.microsoft.com/packaging/2013/01/nuspec.xsd",
		XMLName: xml.Name{
			Space: "http://schemas.microsoft.com/packaging/2013/01/nuspec.xsd",
			Local: "package",
		},
		Metadata: &meta.Metadata{
			MinClientVersion: "1.0.0",
			PackageInfo: meta.PackageInfo{
				ID:                       "MyPackage",
				Version:                  "1.0.0-beta",
				Title:                    "My Full Sample Package",
				Authors:                  "Sample author,Sample author2",
				Owners:                   "Sample author,Sample author2",
				DevelopmentDependency:    true,
				RequireLicenseAcceptance: false,
				License: &meta.LicenseMetadata{
					Type:  "expression",
					Value: "MIT",
				},
				LicenseURL:   "https://licenses.nuget.org/MIT",
				Icon:         "images/test-nuget.png",
				Readme:       "docs/README.md",
				ProjectURL:   "https://example.com/mypackage",
				IconURL:      "https://example.com/images/icon.png",
				Description:  "My test package created from the API.",
				Summary:      "This is a summary for MyPackage.",
				ReleaseNotes: "Initial beta release.",
				Copyright:    "Copyright 2025 by Sample author",
				Tags:         "utility sample",
				Language:     "en-US",
				Serviceable:  true,
				PackageTypes: &meta.PackageTypes{
					PackageTypes: []*meta.PackageType{
						{
							Name:    "DotnetTool",
							Version: "1.0.0",
						},
					},
				},
				Repository: &meta.RepositoryMetadata{
					Type:   "git",
					URL:    "https://github.com/huhouhua/go-nuget",
					Branch: "main",
					Commit: "4a5eec0ec02cbc120f8fa85b3c37327c5c451640",
				},
			},
			Dependencies: &meta.Dependencies{
				Groups: []*meta.DependenciesGroup{
					{
						TargetFramework: ".NETStandard1.4",
						Dependencies: []*meta.Dependency{
							{
								Id:         "Newtonsoft.Json",
								VersionRaw: "10.0.1",
							},
						},
					},
				},
			},
			References: &meta.References{
				Groups: []*meta.ReferenceGroup{
					{
						TargetFramework: "net8.0",
						References: []*meta.Reference{
							{
								File: "System.Text.Json.dll",
							},
							{
								File: "System.Xml.dll",
							},
						},
					},
					{
						TargetFramework: ".NETStandard1.4",
						References: []*meta.Reference{
							{
								File: "System.Xml.Linq.dll",
							},
							{
								File: "System.Xml.Linq.dll",
							},
							{
								File: "System.Xml.Linq.dll",
							},
							{
								File: "System.Xml.Linq.dll",
							},
						},
					},
				},
			},
			FrameworkReferences: &meta.FrameworkReferences{
				Groups: []*meta.FrameworkReferenceGroup{
					{
						TargetFramework: "net5.0",
						FrameworkReferences: []*meta.FrameworkReference{
							{
								Name: "Microsoft.NETCore.App",
							},
						},
					},
				},
			},
			FrameworkAssemblies: &meta.FrameworkAssemblies{
				FrameworkAssembly: []*meta.FrameworkAssembly{
					{
						AssemblyName: []string{
							"System.Xml",
						},
						TargetFramework: ".NETStandard1.4",
					},
				},
			},
			ContentFile: &meta.ContentFile{
				Files: []*meta.ContentFileItem{
					{
						Include:      "contentFiles/any/any/config.json",
						BuildAction:  "None",
						CopyToOutput: "true",
						Flatten:      "true",
					},
				},
			},
		},
	}
	require.Equal(t, want, spec)
	for _, f := range reader.GetFilesFromDir("package/services/metadata/core-properties") {
		require.Equal(t, "package/services/metadata/core-properties/d38055c2bfb8e5c6e2493cfa469a0350.psmdcp", f.Name)
	}
	for _, f := range reader.GetFiles() {
		require.NotEmpty(t, f.Name)
	}
}

func TestReaderNupkg_ErrorScenarios(t *testing.T) {
	buf := &bytes.Buffer{}
	buf.WriteString("test")

	emptyNupkgPath := "testdata/empty.nuspec.test.nupkg"
	emptyFile, err := os.Open(emptyNupkgPath)
	require.NoError(t, err, "open file %s failed %v", emptyNupkgPath, err)
	t.Cleanup(func() {
		_ = emptyFile.Close()
	})
	tests := []struct {
		name   string
		reader io.Reader
		error  error
	}{
		{
			name:   "read stream return error",
			reader: os.Stdout,
			error: &fs.PathError{
				Op:   "read",
				Path: "/dev/stdout",
				Err:  syscall.Errno(9),
			},
		},
		{
			name:   "zip reader return error",
			reader: buf,
			error:  errors.New("zip: not a valid zip file"),
		},
		{
			name:   "empty nuspec return error",
			reader: emptyFile,
			error:  errors.New("no .nuspec file found in the .nupkg archive"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPackageArchiveReader(tt.reader)
			require.Equal(t, tt.error, err)
		})
	}
}
