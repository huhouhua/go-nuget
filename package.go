// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"encoding/xml"
	"fmt"
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
	resp, err := f.client.Do(req, &version)
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
	escapeId := PathEscape(packageId)
	u := fmt.Sprintf("-flatcontainer/%s/%s/%s.nuspec", escapeId, PathEscape(version), escapeId)

	req, err := f.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}
	resp, err := f.client.Do(req, nil)
	if err != nil {
		return nil, resp, err
	}
	var nuspec Nuspec
	err = xml.NewDecoder(resp.Body).Decode(&nuspec)
	if err != nil {
		return nil, resp, err
	}
	dependencyInfo := &PackageDependencyInfo{
		PackageIdentity: &PackageIdentity{},
	}
	err = ApplyPackageDependency(dependencyInfo, WithIdentity(nuspec.Metadata))
	if err != nil {
		return nil, resp, err
	}
	return dependencyInfo, resp, nil
}
