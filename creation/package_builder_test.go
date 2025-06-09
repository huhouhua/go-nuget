// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"archive/zip"
	"github.com/Masterminds/semver/v3"
	"github.com/huhouhua/go-nuget"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestCreatePackage(t *testing.T) {
	builder := NewPackageBuilder(false, false, &log.Logger{})
	builder.Id = "MyPackage"
	builder.Version = semver.New(1, 0, 0, "beta", "")
	builder.Description = "My test package created from the API."
	builder.Title = "My Full Sample Package"
	builder.Summary = "This is a summary for MyPackage."
	builder.ReleaseNotes = "Initial beta release."
	builder.Copyright = "Copyright 2025 by Sample author"
	builder.Language = "en-US"
	builder.Authors = append(builder.Authors, "Sample author", "Sample author2")
	builder.Owners = append(builder.Owners, "Sample author", "Sample author2")
	builder.Tags = append(builder.Tags, "utility", "sample")
	projectURL, _ := url.Parse("https://example.com/mypackage")
	licensesURL, _ := url.Parse("https://licenses.nuget.org/MIT")
	iconURL, _ := url.Parse("https://example.com/images/icon.png")
	builder.ProjectURL = projectURL
	builder.LicenseURL = licensesURL
	builder.IconURL = iconURL
	//builder.Icon = "images/icon.png"
	//builder.Readme = "docs/workflow.md"

	builder.RequireLicenseAcceptance = false
	builder.OutputName = "test"
	builder.EmitRequireLicenseAcceptance = true
	builder.DevelopmentDependency = true
	builder.Serviceable = true
	framework, err := ParseFolderFromDefault("netstandard1.4")
	require.NoError(t, err)
	builder.TargetFrameworks = append(builder.TargetFrameworks, framework)
	versionRange, err := nuget.ParseVersionRange("10.0.1")
	require.NoError(t, err)
	builder.DependencyGroups = append(builder.DependencyGroups, &PackageDependencyGroup{
		TargetFramework: framework,
		Packages: []*nuget.Dependency{
			{
				Id:           "Newtonsoft.Json",
				VersionRange: versionRange,
			},
		},
	})
	//builder.Files = append(builder.Files, &PhysicalPackageFile{})
	nupkgPath := "../_output/MyPackage.nupkg"
	destDir := "../_output/test"
	file, err := os.Create(nupkgPath)
	require.NoError(t, err)
	//t.Cleanup(func() {
	//	_ = file.Close()
	//	_ = os.Remove(file.Name())
	//})
	err = builder.Save(file)
	require.NoError(t, err)
	t.Log(file.Name())

	_ = file.Close()

	unzip(t, nupkgPath, destDir)
}

func TestUnzip(t *testing.T) {
	nupkgPath := "../_output/MyPackage.nupkg"
	destDir := "../_output/test"
	unzip(t, nupkgPath, destDir)
}

func unzip(t *testing.T, zipPath, destDir string) {
	r, err := zip.OpenReader(zipPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = r.Close()
	})

	for _, f := range r.File {
		filePath := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}
		err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		require.NoError(t, err)

		rc, err := f.Open()
		require.NoError(t, err)

		defer rc.Close()

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		require.NoError(t, err)

		defer outFile.Close()

		_, err = io.Copy(outFile, rc)
		require.NoError(t, err)
	}
}
