// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"fmt"
	"io"
	"net/http"
)

type FindPackageResource struct {
	client *Client
}

// ListAllVersions gets all package versions for a package ID.
func (f *FindPackageResource) ListAllVersions(id string, options ...RequestOptionFunc) ([]*NuGetVersion, *http.Response, error) {
	packageId, err := parseID(id)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("-flatcontainer/%s/index.json", PathEscape(packageId))

	req, err := f.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}
	var version struct {
		Versions []string `json:"versions"`
	}
	resp, err := f.client.Do(req, &version, DecoderTypeJSON)
	if err != nil {
		return nil, resp, err
	}
	var versions []*NuGetVersion
	for _, v := range version.Versions {
		nugetVersion, err := Parse(v)
		if err != nil {
			return nil, resp, err
		}
		versions = append(versions, nugetVersion)
	}
	return versions, resp, nil
}

// GetDependencyInfo gets dependency information for a specific package.
func (f *FindPackageResource) GetDependencyInfo(id, version string, options ...RequestOptionFunc) (*PackageDependencyInfo, *http.Response, error) {
	packageId, err := parseID(id)
	if err != nil {
		return nil, nil, err
	}
	packageId = PathEscape(packageId)
	u := fmt.Sprintf("-flatcontainer/%s/%s/%s.nuspec", packageId, PathEscape(version), packageId)

	req, err := f.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}
	var nuspec Nuspec
	resp, err := f.client.Do(req, &nuspec, DecoderTypeXML)
	if err != nil {
		return nil, resp, err
	}
	dependencyInfo := &PackageDependencyInfo{
		DependencyGroups:         make([]*PackageDependencyGroup, 0),
		FrameworkReferenceGroups: make([]*FrameworkSpecificGroup, 0),
	}
	if err = ConfigureDependencyInfo(dependencyInfo, nuspec); err != nil {
		return nil, resp, err
	}
	return dependencyInfo, resp, nil
}

type CopyNupkgOptions struct {
	Version string
	Writer  io.Writer
}

// CopyNupkgToStream downloads a specific package version and copies it to the provided writer.
func (f *FindPackageResource) CopyNupkgToStream(id string, opt *CopyNupkgOptions, options ...RequestOptionFunc) (*http.Response, error) {
	// Parse package ID
	packageId, err := parseID(id)
	if err != nil {
		return nil, err
	}

	packageId, version := PathEscape(packageId), PathEscape(opt.Version)
	// Construct the download URL
	u := fmt.Sprintf("-flatcontainer/%s/%s/%s.%s.nupkg", packageId, version, packageId, version)

	// Create request
	req, err := f.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, err
	}
	// Execute request
	resp, err := f.client.Do(req, opt.Writer, DecoderEmpty)
	if err != nil {
		return resp, err
	}
	return resp, err
}
