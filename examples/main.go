// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/huhouhua/go-nuget"
	"log"
	"time"
)

func main() {
	client, err := nuget.NewClient(
		nuget.WithBaseURL("https://your-private-feed.com/api/v3/"),
		nuget.WithCustomRetryMax(5),
		nuget.WithCustomRetryWaitMinMax(time.Second*1, time.Second*10),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Get all versions of a package
	versions, _, err := client.FindPackage.ListAllVersions("Newtonsoft.Json")
	if err != nil {
		log.Fatalf("Failed to get package versions: %v", err)
	}
	for _, v := range versions {
		fmt.Printf("%s", v.OriginalVersion)
	}
}
