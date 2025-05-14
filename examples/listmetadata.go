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

func listMetadataExample() {
	// Create a new NuGet client with custom retry settings
	client, err := nuget.NewClient(
		nuget.WithSourceURL("https://your-private-feed.com/v3/index.json"),
		nuget.WithCustomRetryMax(5),
		nuget.WithCustomRetryWaitMinMax(time.Second*1, time.Second*10),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	opt := &nuget.ListMetadataOptions{
		IncludePrerelease: true,
		IncludeUnlisted:   false,
	}
	// Get all version metadata of a  package
	packages, _, err := client.MetadataResource.ListMetadata("MyPackage", opt)
	if err != nil {
		log.Fatalf("Failed to get package metadata: %v", err)
	}
	// print the packages
	for _, p := range packages {
		fmt.Printf("version:%s", p.Version)
		fmt.Printf("listed: %v", p.IsListed)
		fmt.Printf("tags: %v", p.Tags)
		fmt.Printf("description: %s", p.Description)
	}

}
