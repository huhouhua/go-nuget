// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"net/http"
	"net/url"
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

	ReadmeFileUrl *url.URL `json:"-"`

	ReportAbuseUrl *url.URL `json:"-"`

	PackageDetailsUrl *url.URL `json:"-"`

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

type ListMetadataOptions struct {
	IncludePrerelease bool
	IncludeUnlisted   bool
}

func (p *PackageMetadataResource) ListMetadata(id string, opt *ListMetadataOptions, options ...RequestOptionFunc) ([]*PackageSearchMetadataRegistration, *http.Response, error) {
	packageId, err := parseID(id)
	if err != nil {
		return nil, nil, err
	}
	baseURL := p.client.getResourceUrl(RegistrationsBaseUrl)
	u := fmt.Sprintf("%s/%s/index.json", baseURL.Path, PathEscape(packageId))
	req, err := p.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}
	index := registrationIndex{}
	resp, err := p.client.Do(req, &index, DecoderTypeJSON)
	if err != nil {
		return nil, resp, err
	}
	packages := make([]*PackageSearchMetadataRegistration, len(index.Items))
	versionRange := All
	for _, item := range index.Items {
		if item == nil {
			return nil, resp, fmt.Errorf("invalid %s", baseURL.String())
		}
		err = p.addMetadataToPackages(packages, item, opt, versionRange)
		if err != nil {
			return nil, nil, err
		}
	}
	return packages, resp, nil
}

func (p *PackageMetadataResource) addMetadataToPackages(packages []*PackageSearchMetadataRegistration, page *registrationPage, opt *ListMetadataOptions, versionRange *VersionRange) error {
	Lower, err := semver.NewVersion(page.Lower)
	if err != nil {
		return err
	}
	upper, err := semver.NewVersion(page.Upper)
	if err != nil {
		return err
	}
	catalogItemVersionRange := NewVersionRange(Lower, upper, true, true)
	if !versionRange.DoesRangeSatisfy(catalogItemVersionRange) {
		return nil
	}
	for _, leafItem := range page.Items {
		v := &NuGetVersion{}
		v.Version, err = semver.NewVersion(leafItem.CatalogEntry.Version)
		if err != nil {
			return err
		}
		if versionRange.Satisfies(v.Version) && (opt.IncludePrerelease || v.IsPrerelease()) && (opt.IncludeUnlisted || leafItem.CatalogEntry.IsListed) {
			if err = p.configureMetadataUrl(leafItem.CatalogEntry); err != nil {
				return err
			} else {
				packages = append(packages, leafItem.CatalogEntry)
			}
		}
	}
	return nil
}

func (p *PackageMetadataResource) configureMetadataUrl(catalogEntry *PackageSearchMetadataRegistration) error {
	reportAbuseUrl := p.client.getResourceUrl(ReportAbuseUriTemplate)
	detailUrl := p.client.getResourceUrl(PackageDetailsUriTemplate)
	readmeUrl := p.client.getResourceUrl(ReportAbuseUriTemplate)
	return ApplyMetadataRegistration(catalogEntry,
		WithReportAbuseUrl(reportAbuseUrl.Path),
		WithPackageDetailsUrl(detailUrl.Path),
		WithReadmeFileUrl(readmeUrl.Path))
}

type MetadataRegistrationFunc func(catalogEntry *PackageSearchMetadataRegistration) error

// ApplyMetadataRegistration applies a list of MetadataRegistrationFunc to a PackageSearchMetadataRegistration.
func ApplyMetadataRegistration(info *PackageSearchMetadataRegistration, options ...MetadataRegistrationFunc) error {
	for _, opt := range options {
		if err := opt(info); err != nil {
			return err
		}
	}
	return nil
}

func WithReportAbuseUrl(urlTemplate string) MetadataRegistrationFunc {
	return func(catalogEntry *PackageSearchMetadataRegistration) error {

		ut := strings.ReplaceAll(urlTemplate, "{id}", strings.ToLower(catalogEntry.PackageId))
		ut = strings.ReplaceAll(urlTemplate, "{version}", catalogEntry.Version)

		if u, err := url.Parse(ut); err != nil {
			return err
		} else {
			catalogEntry.ReportAbuseUrl = u
		}
		return nil
	}
}

func WithPackageDetailsUrl(urlTemplate string) MetadataRegistrationFunc {
	return func(catalogEntry *PackageSearchMetadataRegistration) error {
		ut := strings.ReplaceAll(urlTemplate, "{id}", strings.ToLower(catalogEntry.PackageId))
		ut = strings.ReplaceAll(urlTemplate, "{version}", catalogEntry.Version)

		if u, err := url.Parse(ut); err != nil {
			return err
		} else {
			catalogEntry.PackageDetailsUrl = u
		}
		return nil
	}
}

func WithReadmeFileUrl(urlTemplate string) MetadataRegistrationFunc {
	return func(catalogEntry *PackageSearchMetadataRegistration) error {
		ut := strings.ReplaceAll(urlTemplate, "{lower_id}", strings.ToLower(catalogEntry.PackageId))
		ut = strings.ReplaceAll(urlTemplate, "{lower_version}", strings.ToLower(catalogEntry.Version))

		if u, err := url.Parse(ut); err != nil {
			return err
		} else {
			catalogEntry.ReadmeFileUrl = u
		}
		return nil
	}
}
