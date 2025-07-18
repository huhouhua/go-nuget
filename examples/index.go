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

func listRequestResourceExample() {
	// Create a new NuGet client with custom retry settings
	client, err := nuget.NewClient(
		nuget.WithSourceURL("https://your-private-feed.com/v3/index.json"),
		nuget.WithRetryMax(5),
		nuget.WithRetryWaitMinMax(time.Second*1, time.Second*10),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create client: %v", err))
	}

	// Get request resource
	index, _, err := client.IndexResource.GetIndex()
	if err != nil {
		log.Fatalf("Failed to get resources: %v", err)
	}

	// print the resources url
	for _, r := range index.Resources {
		fmt.Printf("url: %s", r.Id)
	}
}
