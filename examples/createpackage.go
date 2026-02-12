// Copyright (c) 2025 Kevin Berger <huhouhuam@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"os"

	"github.com/huhouhua/go-nuget"
	"github.com/huhouhua/go-nuget/creation"
	version "github.com/huhouhua/go-nuget/version"
)

func createPackageExample() {
	builder := creation.NewPackageBuilder(false, false, logger)
	builder.Id = "MyPackage"
	builder.Version = version.NewVersionFrom(1, 0, 0, "beta", "")
	builder.Description = "My package created from the API."
	builder.Authors = append(builder.Authors, "Sample author")
	netstandard14, err := creation.Parse("netstandard1.4")
	if err != nil {
		log.Fatalf("Failed to parse nuget framework: %v", err)
		return
	}
	versionRange, err := version.ParseRange("10.0.1")
	if err != nil {
		log.Fatalf("Failed to parse version range: %v", err)
		return
	}
	builder.DependencyGroups = append(builder.DependencyGroups, &creation.PackageDependencyGroup{
		TargetFramework: netstandard14,
		Packages: []*nuget.Dependency{
			{
				Id:           "Newtonsoft.Json",
				VersionRange: versionRange,
			},
		},
	})
	nupkgPath := "MyPackage.nupkg"
	file, _ := os.Create(nupkgPath)
	if err = builder.Save(file); err != nil {
		log.Fatalf("Failed create package: %v", err)
	}
}
