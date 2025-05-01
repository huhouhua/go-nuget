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

func searchPackageExample() {
	// Create a new NuGet client with custom retry settings
	client, err := nuget.NewClient(
		nuget.WithBaseURL("https://your-private-feed.com/"),
		nuget.WithCustomRetryMax(5),
		nuget.WithCustomRetryWaitMinMax(time.Second*1, time.Second*10),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	opt := &nuget.SearchOptions{
		SearchTerm:        "MyPackage",
		IncludePrerelease: true,
		Skip:              0,
		Take:              20,
	}
	// Search for a specific package
	packages, _, err := client.SearchResource.Search(opt)
	if err != nil {
		log.Fatalf("Failed to search package: %v", err)
	}
	// print the packages
	for _, p := range packages {
		fmt.Printf("Found package %s %s", p.PackageId, p.Version)
	}
}
