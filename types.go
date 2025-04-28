// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import "strings"

type ServiceTypes []string

type ServiceType string

func (s ServiceType) String() string {
	return string(s)
}

const (
	Version200                            = "/2.0.0"
	Version300beta                        = "/3.0.0-beta"
	Version300rc                          = "/3.0.0-rc"
	Version300                            = "/3.0.0"
	Version340                            = "/3.4.0"
	Version360                            = "/3.6.0"
	Versioned                             = "/Versioned"
	Version470                            = "/4.7.0"
	Version490                            = "/4.9.0"
	Version500                            = "/5.0.0"
	Version510                            = "/5.1.0"
	Version670                            = "/6.7.0"
	Version6110                           = "/6.11.0"
	Version6130                           = "/6.13.0"
	SearchQueryService        ServiceType = "SearchQueryService"
	RegistrationsBaseUrl      ServiceType = "RegistrationsBaseUrl"
	SearchAutocompleteService ServiceType = "SearchAutocompleteService"
	ReportAbuseUriTemplate    ServiceType = "ReportAbuseUriTemplate"
	ReadmeUriTemplate         ServiceType = "ReadmeUriTemplate"
	PackageDetailsUriTemplate ServiceType = "PackageDetailsUriTemplate"
	LegacyGallery             ServiceType = "LegacyGallery"
	PackagePublish            ServiceType = "PackagePublish"
	PackageBaseAddress        ServiceType = "PackageBaseAddress"
	RepositorySignatures      ServiceType = "RepositorySignatures"
	SymbolPackagePublish      ServiceType = "SymbolPackagePublish"
	VulnerabilityInfo         ServiceType = "VulnerabilityInfo"
	OwnerDetailsUriTemplate   ServiceType = "OwnerDetailsUriTemplate"
)

var (
	SearchQueryServiceTypes = ServiceTypes{
		string(SearchQueryService + Versioned),
		string(SearchQueryService + Version340),
		string(SearchQueryService + Version300beta),
	}

	RegistrationsBaseUrlTypes = ServiceTypes{
		string(RegistrationsBaseUrl + Versioned),
		string(RegistrationsBaseUrl + Version360),
		string(RegistrationsBaseUrl + Version340),
		string(RegistrationsBaseUrl + Version300rc),
		string(RegistrationsBaseUrl + Version300beta),
		string(RegistrationsBaseUrl),
	}

	SearchAutocompleteServiceTypes = ServiceTypes{
		string(SearchAutocompleteService + Versioned),
		string(SearchAutocompleteService + Version300beta),
	}

	ReportAbuseTypes = ServiceTypes{
		string(ReportAbuseUriTemplate + Versioned),
		string(ReportAbuseUriTemplate + Version300),
	}

	ReadmeFileUrlTypes = ServiceTypes{
		string(ReadmeUriTemplate + Versioned),
		string(ReadmeUriTemplate + Version6130),
	}

	PackageDetailsUriTemplateTypes = ServiceTypes{
		string(PackageDetailsUriTemplate + Version510),
	}

	LegacyGalleryTypes = ServiceTypes{
		string(LegacyGallery + Versioned),
		string(LegacyGallery + Version200),
	}

	PackagePublishTypes = ServiceTypes{
		string(PackagePublish + Versioned),
		string(PackagePublish + Version200),
	}

	PackageBaseAddressTypes = ServiceTypes{
		string(PackageBaseAddress + Versioned),
		string(PackageBaseAddress + Version300),
	}

	RepositorySignaturesTypes = ServiceTypes{
		string(RepositorySignatures + Version500),
		string(RepositorySignatures + Version490),
		string(RepositorySignatures + Version470),
	}

	SymbolPackagePublishTypes = ServiceTypes{
		string(SymbolPackagePublish + Version490),
	}

	VulnerabilityInfoTypes = ServiceTypes{
		string(VulnerabilityInfo + Version670),
	}

	OwnerDetailsUriTemplateTypes = ServiceTypes{
		string(OwnerDetailsUriTemplate + Version6110),
	}
	typesMap map[ServiceType]*ServiceTypeOptions
)

type ServiceTypeOptions struct {
	Types      ServiceTypes
	DefaultUrl string
}

func init() {
	typesMap = map[ServiceType]*ServiceTypeOptions{
		SearchQueryService:        newTypeOptions(SearchQueryServiceTypes, ""),
		RegistrationsBaseUrl:      newTypeOptions(RegistrationsBaseUrlTypes, ""),
		SearchAutocompleteService: newTypeOptions(SearchAutocompleteServiceTypes, ""),
		ReportAbuseUriTemplate:    newTypeOptions(ReportAbuseTypes, "https://www.nuget.org/packages/{id}/{version}/ReportAbuse"),
		ReadmeUriTemplate:         newTypeOptions(ReadmeFileUrlTypes, ""),
		PackageDetailsUriTemplate: newTypeOptions(PackageDetailsUriTemplateTypes, ""),
		LegacyGallery:             newTypeOptions(LegacyGalleryTypes, ""),
		PackagePublish:            newTypeOptions(PackagePublishTypes, ""),
		PackageBaseAddress:        newTypeOptions(PackageBaseAddressTypes, ""),
		RepositorySignatures:      newTypeOptions(RepositorySignaturesTypes, ""),
		SymbolPackagePublish:      newTypeOptions(SymbolPackagePublishTypes, ""),
		VulnerabilityInfo:         newTypeOptions(VulnerabilityInfoTypes, ""),
		OwnerDetailsUriTemplate:   newTypeOptions(OwnerDetailsUriTemplateTypes, ""),
	}
}

// newTypeOptions creates a new ServiceTypeOptions instance with the given types and default URL.
func newTypeOptions(types ServiceTypes, defaultUrl string) *ServiceTypeOptions {
	return &ServiceTypeOptions{
		Types:      types,
		DefaultUrl: defaultUrl,
	}
}

func (s ServiceTypes) Exist(value string) bool {
	for _, item := range s {
		if strings.EqualFold(item, value) {
			return true
		}
	}
	return false
}

// SearchFilterType The type of filter to apply to the search.
type SearchFilterType int

// SearchOrderBy Order the resulting packages by the specified field.
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

var (
	All = NewVersionRange(nil, nil, true, true)
)

// FloatBehavior represents how version floating should behave
type FloatBehavior int

const (
	// None means no floating behavior
	None FloatBehavior = iota
	// Prerelease allows floating to prerelease versions
	Prerelease
	// Patch allows floating to patch versions
	Patch
	// Minor allows floating to minor versions
	Minor
	// Major allows floating to major versions
	Major
)

const (
	PackageExtension    = ".nupkg"
	SnupkgExtension     = ".snupkg"
	SymbolsExtension    = ".symbols" + PackageExtension
	NuGetSymbolHostName = "nuget.smbsrc.net"
	ServiceEndpoint     = "/api/v2/package"
)
