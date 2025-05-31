// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"net/http"
	"net/url"

	"github.com/Masterminds/semver/v3"
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

func (v *VersionInfo) ParseVersion() (*semver.Version, error) {
	ver, err := semver.NewVersion(v.Version)
	if err != nil {
		return nil, err
	}
	return ver, err
}

// Search retrieves search results
func (p *PackageSearchResource) Search(
	opt *SearchOptions,
	options ...RequestOptionFunc,
) ([]*PackageSearchMetadata, *http.Response, error) {
	baseURL := p.client.getResourceURL(SearchQueryService)
	req, err := p.client.NewRequest(http.MethodGet, baseURL.Path, baseURL, opt, options)
	if err != nil {
		return nil, nil, err
	}
	addSemVer(req.URL)
	result := V3SearchResult{}
	resp, err := p.client.Do(req, &result, DecoderTypeJSON)
	if err != nil {
		return nil, resp, err
	}
	return result.Data, resp, nil
}

func addSemVer(u *url.URL) {
	params := u.Query()
	if !params.Has("semVerLevel") {
		params.Add("semVerLevel", "2.0.0")
	}
	u.RawQuery = params.Encode()
}
