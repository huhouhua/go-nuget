// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"fmt"
	"net/http"
	"sync"
)

var (
	rwMu                       sync.RWMutex
	parsedNuGetVersionsMapping map[string]*NuGetVersion
)

func init() {
	parsedNuGetVersionsMapping = make(map[string]*NuGetVersion)
}

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
