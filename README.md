# go-nuget

A NuGet API v3 client enabling Go programs to interact with NuGet repositories in a simple and uniform way.

![Workflow ci](https://github.com/huhouhua/go-nuget/actions/workflows/go.yml/badge.svg)
[![Sourcegraph](https://sourcegraph.com/github.com/huhouhua/go-nuget/-/badge.svg)](https://sourcegraph.com/github.com/huhouhua/go-nuget?badge)
[![GoDoc](https://godoc.org/github.com/huhouhua/go-nuget?status.svg)](https://godoc.org/github.com/huhouhua/go-nuget)
[![Go Report Card](https://goreportcard.com/badge/github.com/huhouhua/go-nuget)](https://goreportcard.com/report/github.com/huhouhua/go-nuget)
[![Test Coverage](https://codecov.io/gh/huhouhua/go-nuget/branch/main/graph/badge.svg)](https://codecov.io/gh/huhouhua/go-nuget)

## Coverage

This API client package covers the NuGet API v3 endpoints and is updated regularly
to add new and/or missing endpoints. Currently, the following services are supported:

- [x] Service Index
- [x] Package Search
- [x] Package Metadata
- [x] Package Content
- [x] Package Publish
- [x] Package Delete
- [x] Package Download
- [x] Package Versioning
- [x] Package Dependencies
- [x] Package Source Configuration
- [x] Package Source Authentication
- [x] Package Source Retry Logic

## Usage

```go
import "github.com/huhouhua/go-nuget"
```

Construct a new NuGet client, then use the various methods on the client to
access different parts of the NuGet API. For example, to get the service index:

```go
client, err := nuget.NewClient()
if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}

```

There are a few `With...` option functions that can be used to customize
the API client. For example, to set a custom base URL:

```go
client, err := nuget.NewClient(
    nuget.WithBaseURL("https://your-private-feed.com/"),
    nuget.WithCustomRetryMax(5),
    nuget.WithCustomRetryWaitMinMax(time.Second*1, time.Second*10),
)
if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}
```

### Examples

The [examples](examples/) directory contains a couple of clear examples, of which one is partially listed here as well:

```go
package main

import (
	"fmt"
	"github.com/huhouhua/go-nuget"
	"log"
	"time"
)

func main() {
	// Create a new NuGet client with custom retry settings
	client, err := nuget.NewClient(
		nuget.WithBaseURL("https://your-private-feed.com/"),
		nuget.WithCustomRetryMax(5),
		nuget.WithCustomRetryWaitMinMax(time.Second*1, time.Second*10),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Get all versions of a package
	versions, _, err := client.FindPackageResource.ListAllVersions("MyPackage")
	if err != nil {
		log.Fatalf("Failed to get package versions: %v", err)
	}
	// print the versions
	for _, v := range versions {
		fmt.Printf("Found version %s", v.String())
	}
}

```

For complete usage of go-nuget, see the full [package docs](https://godoc.org/github.com/huhouhua/go-nuget).


## Issues

If you have an issue: report it on the [issue tracker](https://github.com/huhouhua/go-nuget/issues)

## Author

Kevin Berger (<huhouhuam@outlook.com>)

## Contributing

Contributions are always welcome. For more information, check out the [contributing guide](CONTRIBUTING.md)

## License

Licensed under the MIT License. See [LICENSE](LICENSE) for details.
