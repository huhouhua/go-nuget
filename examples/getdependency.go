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

func getPackageDependencyInfoExample() {
	// Create a new NuGet client with custom retry settings
	client, err := nuget.NewClient(
		nuget.WithSourceURL("https://your-private-feed.com/v3/index.json"),
		nuget.WithCustomRetryMax(5),
		nuget.WithCustomRetryWaitMinMax(time.Second*1, time.Second*10),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create client: %v", err))
	}

	// Get the dependency information for a specific package
	dependency, _, err := client.FindPackageResource.GetDependencyInfo("MyPackage", "1.0.0-beta")
	if err != nil {
		log.Fatalf("Failed to get package dependency information: %v", err)
	}

	// print the dependency information
	for _, group := range dependency.DependencyGroups {
		for _, p := range group.Packages {
			fmt.Printf("package: %s version:  %s", p.Id, p.VersionRaw)
		}
	}
}
