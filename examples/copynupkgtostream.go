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
	"time"

	"github.com/huhouhua/go-nuget"
)

// copyNupkgToStreamExample demonstrates how to download a NuGet package
func copyNupkgToStreamExample() {
	// Create a new NuGet client with custom retry settings
	client, err := nuget.NewClient(
		nuget.WithSourceURL("https://your-private-feed.com/v3/index.json"),
		nuget.WithCustomRetryMax(5),
		nuget.WithCustomRetryWaitMinMax(time.Second*1, time.Second*10),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create client: %v", err))
	}

	// Package details
	packageID := "MyPackage"
	versionStr := "1.0.0-beta"

	opt := &nuget.CopyNupkgOptions{
		Version: versionStr,
		Writer:  &bytes.Buffer{},
	}

	// Download the package
	_, err = client.FindPackageResource.CopyNupkgToStream(packageID, opt)
	if err != nil {
		log.Fatalf("Failed to download package: %v", err)
	}

	// Create downloads directory if it doesn't exist
	if err := os.MkdirAll("downloads", 0755); err != nil {
		log.Fatalf("Failed to create downloads directory: %v", err)
	}

	// Save the package to a file
	outputFile := filepath.Join("downloads", fmt.Sprintf("%s.%s.nupkg", packageID, versionStr))
	err = os.WriteFile(outputFile, opt.Writer.Bytes(), 0644)

	fmt.Printf("Downloaded package %s %s to %s\n", packageID, versionStr, outputFile)

	reader, err := nuget.NewPackageArchiveReader(opt.Writer)
	// get nuspec file content
	spec, err := reader.Nuspec()
	if err != nil {
		log.Fatalf("Failed Get nuspec file content: %v", err)
	}

	fmt.Printf("Tags:%s", spec.Metadata.Tags)
	fmt.Printf("Description:%s", spec.Metadata.Description)
}
