// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	nugetVersion "github.com/huhouhua/go-nuget/version"
)

type PackageMetadataResource struct {
	client *Client
}

type PackageSearchMetadataRegistration struct {
	*SearchMetadata
	qwnersList []string `json:"-"`

	Authors string `json:"authors"`

	Owners string `json:"owners"`

	CatalogURL string `json:"@id"`

	ReadmeFileURL *url.URL `json:"-"`

	ReportAbuseURL *url.URL `json:"-"`

	PackageDetailsURL *url.URL `json:"-"`
}

// SearchMetadata Package metadata only containing select fields relevant to search results processing and presenting.
type SearchMetadata struct {
	identity *PackageIdentity `json:"-"`

	PackageId string `json:"id"`

	Version string `json:"version"`

	DependencySets []*PackageDependencyGroup `json:"dependencyGroups"`

	Description string `json:"description"`

	DownloadCount int64 `json:"totalDownloads"`

	IconURL string `json:"iconURL"`

	Language string `json:"language"`

	LicenseExpression string `json:"licenseExpression"`

	LicenseExpressionVersion string `json:"licenseExpressionVersion"`

	LicenseURL string `json:"licenseUrl"`

	ProjectURL string `json:"projectUrl"`

	ReadmeURL string `json:"readmeUrl"`

	Published time.Time `json:"published"`

	RequireLicenseAcceptance bool `json:"requireLicenseAcceptance"`

	Summary string `json:"summary"`

	Tags []string `json:"tags"`

	Title string `json:"title"`

	IsListed bool `json:"listed"`

	DeprecationMetadata *PackageDeprecationMetadata `json:"deprecation"`

	Vulnerabilities []*PackageVulnerabilityMetadata `json:"vulnerabilities"`

	PrefixReserved bool `json:"verified"`
}

type PackageDeprecationMetadata struct {
	Message          string                    `json:"message"`
	Reasons          []string                  `json:"reasons"`
	AlternatePackage *AlternatePackageMetadata `json:"alternatePackage"`
}

type AlternatePackageMetadata struct {
	PackageId string `json:"id"`
	Range     string `json:"range"`
}

type PackageVulnerabilityMetadata struct {
	AdvisoryURl string `json:"advisoryURl"`
	Severity    int    `json:"severity"`
}

func (p *SearchMetadata) Identity() (*PackageIdentity, error) {
	if p.identity == nil {
		if identity, err := NewPackageIdentity(p.PackageId, p.Version); err != nil {
			return nil, err
		} else {
			p.identity = identity
		}
	}
	return p.identity, nil
}

func (p *PackageSearchMetadataRegistration) OwnersList() []string {
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

// ListMetadata List of package metadata.
func (p *PackageMetadataResource) ListMetadata(
	id string,
	opt *ListMetadataOptions,
	options ...RequestOptionFunc,
) ([]*PackageSearchMetadataRegistration, *http.Response, error) {
	return p.getMetadata(id, opt, nugetVersion.All, options...)
}

// GetMetadata returns the registration metadata for the id and version
func (p *PackageMetadataResource) GetMetadata(
	id, version string,
	options ...RequestOptionFunc,
) (*PackageSearchMetadataRegistration, *http.Response, error) {
	opt := &ListMetadataOptions{
		IncludePrerelease: true,
		IncludeUnlisted:   true,
	}
	v, err := nugetVersion.Parse(version)
	if err != nil {
		return nil, nil, err
	}
	versionRange, _ := nugetVersion.NewVersionRange(v, v, true, true, nil, "")
	if list, resp, err := p.getMetadata(id, opt, versionRange, options...); err != nil {
		return nil, nil, err
	} else {
		if len(list) > 0 {
			return list[0], resp, nil
		}
		return nil, resp, fmt.Errorf("%s %s not find", id, version)
	}
}

// getMetadata retrieves metadata for a given package ID and version range.
func (p *PackageMetadataResource) getMetadata(
	id string,
	opt *ListMetadataOptions,
	versionRange *nugetVersion.VersionRange,
	options ...RequestOptionFunc,
) ([]*PackageSearchMetadataRegistration, *http.Response, error) {
	packageId, err := parseID(id)
	if err != nil {
		return nil, nil, err
	}
	baseURL := p.client.getResourceURL(RegistrationsBaseURL)
	u := fmt.Sprintf("%s/%s/index.json", baseURL.Path, PathEscape(packageId))
	req, err := p.client.NewRequest(http.MethodGet, u, baseURL, nil, options)
	if err != nil {
		return nil, nil, err
	}
	index := registrationIndex{}
	resp, err := p.client.Do(req, &index, DecoderTypeJSON)
	if err != nil {
		return nil, resp, err
	}
	packages := make([]*PackageSearchMetadataRegistration, 0)
	for _, item := range index.Items {
		if item == nil {
			return nil, resp, fmt.Errorf("invalid %s", baseURL.String())
		}
		if err = p.addMetadataToPackages(&packages, item, opt, versionRange); err != nil {
			return nil, nil, err
		}
	}
	return packages, resp, nil
}

// addMetadataToPackages adds metadata to the given packages slice based on the provided registration page and options.
func (p *PackageMetadataResource) addMetadataToPackages(
	packages *[]*PackageSearchMetadataRegistration,
	page *registrationPage,
	opt *ListMetadataOptions,
	versionRange *nugetVersion.VersionRange,
) error {
	Lower, err := nugetVersion.Parse(page.Lower)
	if err != nil {
		return err
	}
	upper, err := nugetVersion.Parse(page.Upper)
	if err != nil {
		return err
	}
	if !versionRange.DoesRangeSatisfy(Lower, upper) {
		return nil
	}
	for _, leafItem := range page.Items {
		var v *nugetVersion.Version
		v, err = nugetVersion.Parse(leafItem.CatalogEntry.Version)
		if err != nil {
			return err
		}
		if versionRange.Satisfies(v) && (opt.IncludePrerelease || strings.TrimSpace(v.Semver.Prerelease()) == "") &&
			(opt.IncludeUnlisted || leafItem.CatalogEntry.IsListed) {
			if err = p.configureMetadataURL(leafItem.CatalogEntry); err != nil {
				return err
			} else {
				if leafItem.CatalogEntry.DependencySets != nil {
					for _, depSet := range leafItem.CatalogEntry.DependencySets {
						for _, dependency := range depSet.Packages {
							if err = dependency.Parse(); err != nil {
								return err
							}
						}
					}
				}
				*packages = append(*packages, leafItem.CatalogEntry)
			}
		}
	}
	return nil
}

// configureMetadataURL configures the metadata URLs for the given PackageSearchMetadataRegistration.
func (p *PackageMetadataResource) configureMetadataURL(catalogEntry *PackageSearchMetadataRegistration) error {
	reportAbuseURL := p.client.getResourceURL(ReportAbuseURLTemplate)
	detailURL := p.client.getResourceURL(PackageDetailsURLTemplate)
	readmeURL := p.client.getResourceURL(ReadmeURLTemplate)
	return ApplyMetadataRegistration(catalogEntry,
		WithReportAbuseURL(reportAbuseURL),
		WithPackageDetailsURL(detailURL),
		WithReadmeFileURL(readmeURL))
}

// MetadataRegistrationFunc is a function that modifies the PackageSearchMetadataRegistration.
type MetadataRegistrationFunc func(page *PackageSearchMetadataRegistration) error

// ApplyMetadataRegistration applies a list of MetadataRegistrationFunc to a PackageSearchMetadataRegistration.
func ApplyMetadataRegistration(page *PackageSearchMetadataRegistration, options ...MetadataRegistrationFunc) error {
	for _, opt := range options {
		if err := opt(page); err != nil {
			return err
		}
	}
	return nil
}

// Helper function to parse and replace placeholders in URL templates
func parseAndReplaceURL(template *url.URL, replacements map[string]string) (*url.URL, error) {
	if template == nil {
		return nil, nil
	}
	decodedTemplate, err := url.QueryUnescape(template.String())
	if err != nil {
		return nil, err
	}
	for placeholder, value := range replacements {
		decodedTemplate = strings.ReplaceAll(decodedTemplate, placeholder, value)
	}
	return url.Parse(decodedTemplate)
}

// WithReportAbuseURL sets the ReportAbuseURL field of the PackageSearchMetadataRegistration.
func WithReportAbuseURL(urlTemplate *url.URL) MetadataRegistrationFunc {
	return func(page *PackageSearchMetadataRegistration) error {
		replacements := map[string]string{
			"{id}":      strings.ToLower(page.PackageId),
			"{version}": page.Version,
		}
		if u, err := parseAndReplaceURL(urlTemplate, replacements); err != nil {
			return err
		} else {
			page.ReportAbuseURL = u
		}
		return nil
	}
}

// WithPackageDetailsURL sets the PackageDetailsURL field of the PackageSearchMetadataRegistration.
func WithPackageDetailsURL(urlTemplate *url.URL) MetadataRegistrationFunc {
	return func(page *PackageSearchMetadataRegistration) error {
		replacements := map[string]string{
			"{id}":      strings.ToLower(page.PackageId),
			"{version}": page.Version,
		}
		if u, err := parseAndReplaceURL(urlTemplate, replacements); err != nil {
			return err
		} else {
			page.PackageDetailsURL = u
		}
		return nil
	}
}

// WithReadmeFileURL sets the ReadmeFileURL field of the PackageSearchMetadataRegistration.
func WithReadmeFileURL(urlTemplate *url.URL) MetadataRegistrationFunc {
	return func(page *PackageSearchMetadataRegistration) error {
		replacements := map[string]string{
			"{lower_id}":      strings.ToLower(page.PackageId),
			"{lower_version}": strings.ToLower(page.Version),
		}
		if u, err := parseAndReplaceURL(urlTemplate, replacements); err != nil {
			return err
		} else {
			page.ReadmeFileURL = u
		}
		return nil
	}
}
