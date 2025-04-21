// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"net/http"
)

// SearchFilterType The type of filter to apply to the search.
type SearchFilterType int

type SearchOrderBy int

const (
	// IsLatestVersion Only select the latest stable version of a package per package ID. Given the server supports IsAbsoluteLatestVersion,
	//a package that is IsLatestVersion should never be prerelease. Also, it does not make sense to
	//look for a IsLatestVersion package when also including prerelease.
	IsLatestVersion SearchFilterType = iota

	// IsAbsoluteLatestVersion Only select the absolute latest version of a package per package ID.
	// It does not make sense to look for a IsAbsoluteLatestVersion when excluding prerelease.
	IsAbsoluteLatestVersion SearchFilterType = iota

	// Id Order the resulting packages by package ID.
	Id SearchOrderBy = 3
)

type PackageSearchResource struct {
	client *Client
}

type SearchOptions struct {

	// SearchTerm The term we're searching for.
	SearchTerm string `url:"q,omitempty" json:"searchTerm,omitempty"`

	// IncludePrerelease Include prerelease packages in search
	IncludePrerelease bool `url:"prerelease,omitempty" json:"includePrerelease,omitempty"`

	// IncludeDelisted Include unlisted packages in search
	IncludeDelisted bool `url:"includeDelisted,omitempty" json:"IncludeDelisted,omitempty"`

	// PackageTypes Restrict the search to certain package types.
	PackageTypes []string `url:"packageTypeFilter,omitempty" json:"PackageTypes,omitempty"`

	//// Filter The optional filter type. Absense of this value indicates that all versions should be returned
	//Filter SearchFilterType
	//
	//// OrderBy The optional order by. Absense of this value indicates that search results should be ordered by relevance.
	//OrderBy SearchOrderBy

	// SupportedFrameworks Filter to only the list of packages compatible with these frameworks.
	SupportedFrameworks []string `url:"supportedFramework,omitempty" json:"supportedFrameworks,omitempty"`

	// Skip skip how many items from beginning of list.
	Skip int `url:"skip,omitempty" json:"skip,omitempty"`

	// Take return how many items.
	Take int `url:"take,omitempty" json:"take,omitempty"`
}

type V3SearchResult struct {
	TotalHits uint64                   `json:"totalHits"`
	Data      []*PackageSearchMetadata `json:"data"`
}

type PackageSearchMetadata struct {
	*SearchMetadata
	Versions []*VersionInfo `json:"versions"`
	Authors  []string       `json:"authors"`
	Owners   []string       `json:"owners"`
}

type VersionInfo struct {
	Url           string `json:"@id"`
	Version       string `json:"version"`
	DownloadCount uint64 `json:"downloads"`
}

func (v *VersionInfo) ParseVersion() (*NuGetVersion, error) {
	ver, err := semver.NewVersion(v.Version)
	if err != nil {
		return nil, err
	}
	return &NuGetVersion{
		Version: ver,
	}, err
}

// Search Retrieves search results
func (p *PackageSearchResource) Search(opt *SearchOptions, options ...RequestOptionFunc) ([]*PackageSearchMetadata, *http.Response, error) {
	baseURL := p.client.getResourceUrl(SearchQueryService)
	req, err := p.client.NewRequest(http.MethodGet, baseURL.Path, opt, options)
	if err != nil {
		return nil, nil, err
	}
	req.URL.RawQuery = fmt.Sprintf("%s&semVerLevel=2.0.0", req.URL.RawQuery)
	result := V3SearchResult{}
	resp, err := p.client.Do(req, &result, DecoderTypeJSON)
	if err != nil {
		return nil, resp, err
	}
	return result.Data, resp, nil
}
