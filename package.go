// Copyright (c) 2025 Kevin Berger <huhouhuam@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/huhouhua/go-nuget/internal/meta"

	nugetVersion "github.com/huhouhua/go-nuget/version"
)

type FindPackageResource struct {
	client *Client
}

// ListAllVersions gets all package versions for a package ID.
func (f *FindPackageResource) ListAllVersions(
	id string,
	options ...RequestOptionFunc,
) ([]*nugetVersion.Version, *http.Response, error) {
	packageId, err := parseID(id)
	if err != nil {
		return nil, nil, err
	}
	baseURL := f.client.getResourceURL(PackageBaseAddress)
	u := fmt.Sprintf("%s/%s/index.json", baseURL.Path, PathEscape(packageId))

	req, err := f.client.NewRequest(http.MethodGet, u, baseURL, nil, options)
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

	var versions []*nugetVersion.Version
	for _, v := range version.Versions {
		if nv, err := nugetVersion.Parse(v); err != nil {
			return nil, resp, err
		} else {
			versions = append(versions, nv)
		}
	}
	return versions, resp, nil
}

// GetDependencyInfo gets dependency information for a specific package.
func (f *FindPackageResource) GetDependencyInfo(
	id, version string,
	options ...RequestOptionFunc,
) (*meta.PackageDependencyInfo, *http.Response, error) {
	packageId, err := parseID(id)
	if err != nil {
		return nil, nil, err
	}
	packageId = PathEscape(packageId)
	baseURL := f.client.getResourceURL(PackageBaseAddress)
	u := fmt.Sprintf("%s/%s/%s/%s.nuspec", baseURL.Path, packageId, PathEscape(version), packageId)

	req, err := f.client.NewRequest(http.MethodGet, u, baseURL, nil, options)
	if err != nil {
		return nil, nil, err
	}
	var nuspec meta.Nuspec
	resp, err := f.client.Do(req, &nuspec, DecoderTypeXML)
	if err != nil {
		return nil, resp, err
	}
	dependencyInfo := &meta.PackageDependencyInfo{
		DependencyGroups:         make([]*meta.PackageDependencyGroup, 0),
		FrameworkReferenceGroups: make([]*meta.FrameworkSpecificGroup, 0),
	}
	if err = meta.ConfigureDependencyInfo(dependencyInfo, nuspec); err != nil {
		return nil, resp, err
	}
	return dependencyInfo, resp, nil
}

type CopyNupkgOptions struct {
	Version string
	Writer  *bytes.Buffer
}

// CopyNupkgToStream downloads a specific package version and copies it to the provided writer.
func (f *FindPackageResource) CopyNupkgToStream(
	id string,
	opt *CopyNupkgOptions,
	options ...RequestOptionFunc,
) (*http.Response, error) {
	// Parse package ID
	packageId, err := parseID(id)
	if err != nil {
		return nil, err
	}
	packageId, version := PathEscape(packageId), PathEscape(opt.Version)
	baseURL := f.client.getResourceURL(PackageBaseAddress)
	// Construct the download URL
	u := fmt.Sprintf("%s/%s/%s/%s.%s.nupkg", baseURL.Path, packageId, version, packageId, version)

	// Create request
	req, err := f.client.NewRequest(http.MethodGet, u, baseURL, nil, options)
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
