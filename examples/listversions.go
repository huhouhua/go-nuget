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

func listAllVersionsExample() {
	// Create a new NuGet client with custom retry settings
	client, err := nuget.NewClient(
		nuget.WithSourceURL("https://your-private-feed.com/v3/index.json"),
		nuget.WithRetryMax(5),
		nuget.WithRetryWaitMinMax(time.Second*1, time.Second*10),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create client: %v", err))
	}

	// Get all versions of a package
	versions, _, err := client.FindPackageResource.ListAllVersions("MyPackage")
	if err != nil {
		log.Fatalf("Failed to get package versions: %v", err)
	}
	// print the versions
	for _, v := range versions {
		fmt.Printf("Found version %s", v.OriginalVersion)
	}

}
