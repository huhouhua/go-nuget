// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"time"

	"github.com/huhouhua/go-nuget"
)

func pushPackageExample() {
	// Create a new NuGet client with custom retry settings
	client, err := nuget.NewOAuthClient(
		"my-api-key",
		nuget.WithSourceURL("https://your-private-feed.com/v3/index.json"),
		nuget.WithCustomRetryMax(5),
		nuget.WithCustomRetryWaitMinMax(time.Second*1, time.Second*10),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	opt := &nuget.PushPackageOptions{
		TimeoutInDuration: time.Second * 60,
	}
	packagePath := "MyPackage.nupkg"

	// Push the package to the NuGet repository
	_, err = client.UpdateResource.Push(packagePath, opt)
	if err != nil {
		log.Fatalf("Failed to push package: %v", err)
	}
	log.Printf("push package %s successfully", packagePath)
}
