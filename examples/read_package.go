// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/huhouhua/go-nuget"
)

func readPackageExample() {
	nupkgPath := "MyPackage.nupkg"
	file, err := os.Open(nupkgPath)
	if err != nil {
		log.Fatalf("Failed to open %s package: %v", nupkgPath, err)
	}
	defer file.Close()
	var dest io.Writer = &bytes.Buffer{}
	if _, err = io.Copy(dest, file); err != nil {
		log.Fatal(err)
	}
	reader, err := nuget.NewPackageArchiveReader(dest)
	if err != nil {
		log.Fatalf("Failed to parse nuget package archive: %v", err)
	}
	// get nuspec file content
	spec, err := reader.Nuspec()
	if err != nil {
		log.Fatalf("Failed Get nuspec file content: %v", err)
	}

	fmt.Printf("ID: %s", spec.Metadata.ID)
	fmt.Printf("Version: %s", spec.Metadata.Version)
	fmt.Printf("Description: %s", spec.Metadata.Description)
	fmt.Printf("Authors: %s", spec.Metadata.Authors)

	for _, group := range spec.Metadata.Dependencies.Groups {

		fmt.Printf("- %s", group.TargetFramework)

		for _, dependency := range group.Dependencies {
			fmt.Printf("-  %s %s", dependency.Id, dependency.VersionRangeRaw)
		}
	}
}
