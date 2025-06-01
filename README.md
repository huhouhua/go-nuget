<div style="text-align: center">
  <p align="center">
     <a href="https://pkg.go.dev/github.com/huhouhua/go-nuget" > 
	<img src="https://raw.githubusercontent.com/huhouhua/go-nuget/main/docs/img/logo.png"
	     alt="go-nuget"
    	     width="600">
     </a>
    <br><br>
    <p>A NuGet API v3 client enabling Go programs to interact with NuGet repositories in a simple and uniform way.</p>
  </p>
</div>

<p align="center">
	<a href="#-installation-and-documentation">Installation</a> ❘
	<a href="#-coverage">Coverage</a> ❘
	<a href="#-usage">Usage</a> ❘
	<a href="#-examples">Examples</a> ❘
	<a href="#-license">License</a>
</p>

![Workflow ci](https://github.com/huhouhua/go-nuget/actions/workflows/go.yml/badge.svg)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/huhouhua/go-nuget/blob/main/LICENSE)
[![release](https://img.shields.io/github/release-pre/huhouhua/go-nuget.svg)](https://github.com/huhouhua/go-nuget/releases)
[![GoDoc](https://godoc.org/github.com/huhouhua/go-nuget?status.svg)](https://godoc.org/github.com/huhouhua/go-nuget)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/huhouhua/go-nuget?logo=go)
[![Go Report Card](https://goreportcard.com/badge/github.com/huhouhua/go-nuget)](https://goreportcard.com/report/github.com/huhouhua/go-nuget)
[![Test Coverage](https://codecov.io/gh/huhouhua/go-nuget/branch/main/graph/badge.svg)](https://codecov.io/gh/huhouhua/go-nuget)

## 🤘&nbsp; Coverage

This API client package covers the NuGet API v3 endpoints and is updated regularly
to add new and/or missing endpoints. Currently, the following services are supported:

- [x] Service Resources
- [x] Package Search
- [x] Package Metadata
- [x] Package Publish
- [x] Package Delete
- [x] Package Download
- [x] Package Versioning
- [x] Package Read
- [ ] Package Create
- [x] Package Dependencies
- [x] Package Source Configuration
- [x] Package Source Authentication

## 🚀&nbsp; Installation and Documentation

When used with Go modules, use the following import path:
```shell
go get github.com/huhouhua/go-nuget
```
You can find the docs at [go docs](https://pkg.go.dev/github.com/huhouhua/go-nuget).

## 📄&nbsp; Usage
Construct a new NuGet client, then use the various methods on the client to
access different parts of the NuGet API. For example, to get the service resources:

```go
client, err := nuget.NewClient()
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
```

There are a few `With...` option functions that can be used to customize
the API client. For example, to set a custom source URL:

```go
client, err := nuget.NewClient(
    nuget.WithSourceURL("https://your-private-feed.com/v3/index.json"),
    nuget.WithCustomRetryMax(5),
    nuget.WithCustomRetryWaitMinMax(time.Second*1, time.Second*10),
)
if err != nil {
    panic(fmt.Sprintf("Failed to create client: %v", err))
}
```

## 🥙&nbsp; Examples

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
		nuget.WithSourceURL("https://your-private-feed.com/v3/index.json"),
		nuget.WithCustomRetryMax(5),
		nuget.WithCustomRetryWaitMinMax(time.Second*1, time.Second*10),
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
		fmt.Printf("Found version %s", v.String())
	}
}

```

For complete usage of go-nuget, see the full [package docs](https://godoc.org/github.com/huhouhua/go-nuget).

##  Full Examples

### Full Examples : API Resources Operations
* [index.go](https://github.com/huhouhua/go-nuget/blob/main/examples/index.go)

### Full Examples : Package Find Operations
* [listversions.go](https://github.com/huhouhua/go-nuget/blob/main/examples/listversions.go)
* [getdependency.go](https://github.com/huhouhua/go-nuget/blob/main/examples/getdependency.go)
* [copynupkgtostream.go](https://github.com/huhouhua/go-nuget/blob/main/examples/copynupkgtostream.go)

### Full Examples : Package Read Operations
* [readpackage.go](https://github.com/huhouhua/go-nuget/blob/main/examples/readpackage.go)

### Full Examples : Package Search Operations
* [search.go](https://github.com/huhouhua/go-nuget/blob/main/examples/search.go)

### Full Examples : Package Metadata Operations
* [getmetadata.go](https://github.com/huhouhua/go-nuget/blob/main/examples/getmetadata.go)
* [listmetadata.go](https://github.com/huhouhua/go-nuget/blob/main/examples/listmetadata.go)

### Full Examples : Package Push Operations
* [pushpackage.go](https://github.com/huhouhua/go-nuget/blob/main/examples/pushpackage.go)
* [pushwithstream.go](https://github.com/huhouhua/go-nuget/blob/main/examples/pushwithstream.go)
* [deletepackage.go](https://github.com/huhouhua/go-nuget/blob/main/examples/deletepackage.go)

## 🤝&nbsp;Issues

If you have an issue: report it on the [issue tracker](https://github.com/huhouhua/go-nuget/issues)

## 👤&nbsp;Author

Kevin Berger (<huhouhuam@outlook.com>)

## 🧑‍💻&nbsp;Contributing

Contributions are always welcome. For more information, check out the [contributing guide](CONTRIBUTING.md)

## 📘&nbsp; License

Licensed under the MIT License. See [LICENSE](LICENSE) for details.
