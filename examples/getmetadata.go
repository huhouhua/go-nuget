// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/huhouhua/go-nuget"
)

func getMetadataExample() {
	// Create a new NuGet client with custom retry settings
	client, err := nuget.NewClient(
		nuget.WithSourceURL("https://your-private-feed.com/v3/index.json"),
		nuget.WithRetryMax(5),
		nuget.WithRetryWaitMinMax(time.Second*1, time.Second*10),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create client: %v", err))
	}
	// Get metadata of a package
	packageMetadata, _, err := client.MetadataResource.GetMetadata("MyPackage", "1.0.0")

	if err != nil {
		log.Fatalf("Failed to get package metadata: %v", err)
	}
	// print the package metadata
	fmt.Printf("version:%s", packageMetadata.Version)
	fmt.Printf("listed: %v", packageMetadata.IsListed)
	fmt.Printf("tags: %v", packageMetadata.Tags)
	fmt.Printf("description: %s", packageMetadata.Description)

}
