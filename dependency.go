// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"net/http"

	"github.com/huhouhua/go-nuget/internal/framework"
)

type DependencyInfoResource struct {
	client *Client
}

// ResolvePackage Retrieve dependency info for a single package.
// Returns dependency info for the given package if it exists. If the package is not found null is
func (d *DependencyInfoResource) ResolvePackage(id, version string, fw *framework.Framework, options ...RequestOptionFunc) (*http.Response, error) {
	return nil, nil
}

// ResolveAllPackage Retrieve the available packages and their dependencies.
func (d *DependencyInfoResource) ResolveAllPackage() {

}

// ResolveAllPackageFromRemote Retrieve the available packages and their dependencies.
func (d *DependencyInfoResource) ResolveAllPackageFromRemote() {

}
