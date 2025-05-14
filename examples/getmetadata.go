package main

import (
	"fmt"
	"github.com/huhouhua/go-nuget"
	"log"
	"time"
)

func getMetadataExample() {
	// Create a new NuGet client with custom retry settings
	client, err := nuget.NewClient(
		nuget.WithSourceURL("https://your-private-feed.com/v3/index.json"),
		nuget.WithCustomRetryMax(5),
		nuget.WithCustomRetryWaitMinMax(time.Second*1, time.Second*10),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
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
