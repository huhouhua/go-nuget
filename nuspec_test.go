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
	"strings"
	"syscall"
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

	reader, err := NewPackageArchiveReader(file)
	require.NoError(t, err, "Failed to parse nuget package archive")
	_, err = reader.Nuspec()
	require.NoErrorf(t, err, "Failed Get nuspec file content: %v", err)
	spec, _ := reader.Nuspec()

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
				Repository: &RepositoryMetadata{
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

	for _, f := range reader.GetFilesFromDir("testLibrary") {
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

func TestFromFile(t *testing.T) {
	_, err := FromFile("non_existent_file.nuspec")
	require.Error(t, err, "expected error when file does not exist")

	nuspec, err := FromFile("testdata/myTestLibrary.nuspec")
	require.NoError(t, err)
	require.NotNil(t, nuspec)
}
func TestFromReader(t *testing.T) {
	_, err := FromReader(&errorReader{})
	if err == nil || !strings.Contains(err.Error(), "read error") {
		t.Fatal("expected read error")
	}
	nuspecFile, err := os.Open("testdata/myTestLibrary.nuspec")
	require.NoError(t, err)

	nuspec, err := FromReader(nuspecFile)
	require.NoError(t, err)
	require.NotNil(t, nuspec)
}

func TestFromBytes(t *testing.T) {
	t.Run("invalid xml", func(t *testing.T) {
		invalidXML := []byte("<invalid><xml>")
		_, err := FromBytes(invalidXML)
		if err == nil {
			t.Fatal("expected error for invalid XML")
		}
	})
	t.Run("empty input", func(t *testing.T) {
		_, err := FromBytes([]byte{})
		if err == nil {
			t.Fatal("expected error for empty input")
		}
	})
	t.Run("valid xml", func(t *testing.T) {
		validXML := []byte(`
		<package>
			<metadata>
				<id>TestPackage</id>
			</metadata>
		</package>`)

		nuspec, err := FromBytes(validXML)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if nuspec.Metadata.ID != "TestPackage" {
			t.Errorf("expected ID to be 'TestPackage', got '%s'", nuspec.Metadata.ID)
		}
	})
	t.Run("read return success", func(t *testing.T) {
		nuspecFile, err := os.Open("testdata/myTestLibrary.nuspec")
		require.NoError(t, err)

		nuspecBytes, err := io.ReadAll(nuspecFile)
		require.NoError(t, err)

		nuspec, err := FromBytes(nuspecBytes)
		require.NoError(t, err)
		require.NotNil(t, nuspec)
	})
}

type errorReader struct{}

func (e *errorReader) Close() error {
	return nil
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}
