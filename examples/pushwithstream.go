// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/huhouhua/go-nuget"
)

func pushWithStreamExample() {
	// Create a new NuGet client with custom retry settings
	client, err := nuget.NewOAuthClient(
		"my-api-key",
		nuget.WithSourceURL("https://your-private-feed.com/v3/index.json"),
		nuget.WithRetryMax(5),
		nuget.WithRetryWaitMinMax(time.Second*1, time.Second*10),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create client: %v", err))
	}

	opt := &nuget.PushPackageOptions{
		TimeoutInDuration: time.Second * 60,
	}
	packagePath := "MyPackage.nupkg"
	file, err := os.Open(packagePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	// Push the package to the NuGet repository
	_, err = client.UpdateResource.PushWithStream(file, opt)
	if err != nil {
		log.Fatalf("Failed to push package: %v", err)
	}
	log.Printf("push package %s successfully", packagePath)
}
