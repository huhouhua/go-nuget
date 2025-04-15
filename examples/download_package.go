// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/huhouhua/go-nuget"
)

// downloadNupkgExample demonstrates how to download a NuGet package
func downloadNupkgExample() {
	// Create a new client
	client, err := nuget.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Package details
	packageID := "Newtonsoft.Json"
	versionStr := "12.0.1"

	// Parse version
	version, err := nuget.Parse(versionStr)
	if err != nil {
		log.Fatalf("Failed to parse version: %v", err)
	}

	opt := &nuget.CopyNupkgOptions{
		Version: versionStr,
		Writer:  &bytes.Buffer{},
	}

	// Download the package
	resp, err := client.FindPackage.CopyNupkgToStream(packageID, opt)
	if err != nil {
		log.Fatalf("Failed to download package: %v", err)
	}

	// Create downloads directory if it doesn't exist
	if err := os.MkdirAll("downloads", 0755); err != nil {
		log.Fatalf("Failed to create downloads directory: %v", err)
	}

	// Save the package to a file
	outputFile := filepath.Join("downloads", fmt.Sprintf("%s.%s.nupkg", packageID, versionStr))
	if err := os.WriteFile(outputFile, opt.writer.(*bytes.Buffer).Bytes(), 0644); err != nil {
		log.Fatalf("Failed to save package: %v", err)
	}

	fmt.Printf("Downloaded package %s %s to %s\n", packageID, versionStr, outputFile)

	reader, err := nuget.NewPackageArchiveReader(opt.Writer)
	// get nuspec file content
	spec, err := reader.Nuspec()

	// TODO: Add package reading functionality
	// In C# this would use PackageArchiveReader, but we'll need to implement
	// our own package reading functionality or use a third-party library
}
