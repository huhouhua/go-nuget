// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

func main() {

	logger.Println("Listing resource...")
	listRequestResourceExample()

	logger.Println("Listing package versions...")
	listAllVersionsExample()

	logger.Println("Downloading package...")
	downloadNupkgExample()

	logger.Println("Get package metadata...")
	listMetadataExample()

	logger.Println("Get package dependency...")
	packageDependencyInfoExample()

	logger.Println("Searching packages...")
	searchPackageExample()

	logger.Println("Reading a package...")
	readPackageExample()

	logger.Println("Pushing a package...")
	pushPackageExample()

	logger.Println("Deleting a package...")
	deletePackageExample()
}
