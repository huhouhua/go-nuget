// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// https://api.nuget.org/v3/registration5-gz-semver2/gitlabapiclient/index.json
package nuget

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type PackageMetadataResource struct {
	client *Client
}

type PackageSearchMetadataRegistration struct {
	*PackageSearchMetadata
	CatalogUri string `json:"@id"`
}

// PackageSearchMetadata Package metadata only containing select fields relevant to search results processing and presenting.
type PackageSearchMetadata struct {
	qwnersList []string         `json:"-"`
	identity   *PackageIdentity `json:"-"`

	PackageId string `json:"id"`

	Version string `json:"version"`

	Authors string `json:"authors"`

	DependencySets []*PackageDependencyGroup `json:"dependencyGroups"`

	Description string `json:"description"`

	DownloadCount int64 `json:"totalDownloads"`

	IconUrl string `json:"iconUrl"`

	Language string `json:"language"`

	LicenseExpression string `json:"licenseExpression"`

	LicenseExpressionVersion string `json:"licenseExpressionVersion"`

	LicenseUrl string `json:"licenseUrl"`

	ProjectUrl string `json:"projectUrl"`

	ReadmeUrl string `json:"readmeUrl"`

	Published time.Time `json:"published"`

	Owners string `json:"owners"`

	RequireLicenseAcceptance bool `json:"requireLicenseAcceptance"`

	Summary string `json:"summary"`

	Tags []string `json:"tags"`

	Title string `json:"title"`

	IsListed bool `json:"listed"`

	DeprecationMetadata *PackageDeprecationMetadata `json:"deprecation"`

	Vulnerabilities []*PackageVulnerabilityMetadata `json:"vulnerabilities"`

	PrefixReserved bool `json:"verified"`
}

type VersionInfo struct {
}

type PackageDeprecationMetadata struct {
	Message          string                   `json:"message"`
	Reasons          []string                 `json:"reasons"`
	AlternatePackage AlternatePackageMetadata `json:"alternatePackage"`
}

type AlternatePackageMetadata struct {
	PackageId string `json:"id"`
	Range     string `json:"range"`
}

type PackageVulnerabilityMetadata struct {
	AdvisoryUrl string `json:"advisoryUrl"`
	Severity    int    `json:"severity"`
}

func (p *PackageSearchMetadata) Identity() (*PackageIdentity, error) {
	if p.identity == nil {
		if identity, err := NewPackageIdentity(p.PackageId, p.Version); err != nil {
			return nil, err
		} else {
			p.identity = identity
		}
	}
	return p.identity, nil
}

func (p *PackageSearchMetadata) OwnersList() []string {
	if p.qwnersList == nil && p.Owners != "" {
		p.qwnersList = strings.Split(p.Owners, ",")
	}
	return p.qwnersList
}

// registrationIndex
// Source: https://docs.microsoft.com/en-us/nuget/api/registration-base-url-resource#registration-index
type registrationIndex struct {
	Items []*registrationPage `json:"items"`
}

type registrationPage struct {
	Url   string                  `json:"@id"`
	Items []*registrationLeafItem `json:"items"`
	Lower string                  `json:"lower"`
	Upper string                  `json:"upper"`
}

type registrationLeafItem struct {
	CatalogEntry   *PackageSearchMetadataRegistration `json:"catalogEntry"`
	PackageContent string                             `json:"packageContent"`
}

func (p *PackageMetadataResource) ListMetadata(id string, options ...RequestOptionFunc) ([]*PackageSearchMetadata, *http.Response, error) {
	packageId, err := parseID(id)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("-registration5-gz-semver2/%s/index.json", PathEscape(packageId))
	req, err := p.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}
	index := registrationIndex{}

	resp, err := p.client.Do(req, &index, DecoderTypeJSON)
	if err != nil {
		return nil, resp, err
	}
	packages := make([]*PackageSearchMetadata, len(index.Items))
	for _, item := range index.Items {
		for _, leafItem := range item.Items {
			packages = append(packages, leafItem.CatalogEntry.PackageSearchMetadata)
		}
	}
	return packages, resp, nil
}
