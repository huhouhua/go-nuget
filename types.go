// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

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
	RegistrationsBaseURL      ServiceType = "RegistrationsBaseUrl"
	SearchAutocompleteService ServiceType = "SearchAutocompleteService"
	ReportAbuseURLTemplate    ServiceType = "ReportAbuseUriTemplate"
	ReadmeURLTemplate         ServiceType = "ReadmeUriTemplate"
	PackageDetailsURLTemplate ServiceType = "PackageDetailsUriTemplate"
	LegacyGallery             ServiceType = "LegacyGallery"
	PackagePublish            ServiceType = "PackagePublish"
	PackageBaseAddress        ServiceType = "PackageBaseAddress"
	RepositorySignatures      ServiceType = "RepositorySignatures"
	SymbolPackagePublish      ServiceType = "SymbolPackagePublish"
	VulnerabilityInfo         ServiceType = "VulnerabilityInfo"
	OwnerDetailsURLTemplate   ServiceType = "OwnerDetailsUriTemplate"
)

var (
	SearchQueryServiceTypes = ServiceTypes{
		string(SearchQueryService + Versioned),
		string(SearchQueryService + Version340),
		string(SearchQueryService + Version300beta),
	}

	RegistrationsBaseURLTypes = ServiceTypes{
		string(RegistrationsBaseURL + Versioned),
		string(RegistrationsBaseURL + Version360),
		string(RegistrationsBaseURL + Version340),
		string(RegistrationsBaseURL + Version300rc),
		string(RegistrationsBaseURL + Version300beta),
		string(RegistrationsBaseURL),
	}

	SearchAutocompleteServiceTypes = ServiceTypes{
		string(SearchAutocompleteService + Versioned),
		string(SearchAutocompleteService + Version300beta),
	}

	ReportAbuseTypes = ServiceTypes{
		string(ReportAbuseURLTemplate + Versioned),
		string(ReportAbuseURLTemplate + Version300),
	}

	ReadmeFileURLTypes = ServiceTypes{
		string(ReadmeURLTemplate + Versioned),
		string(ReadmeURLTemplate + Version6130),
	}

	PackageDetailsURLTemplateTypes = ServiceTypes{
		string(PackageDetailsURLTemplate + Version510),
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

	OwnerDetailsURLTemplateTypes = ServiceTypes{
		string(OwnerDetailsURLTemplate + Version6110),
	}
	typesMap map[ServiceType]*ServiceTypeOptions
)

type ServiceTypeOptions struct {
	Types      ServiceTypes
	DefaultURL string
}

func init() {
	typesMap = map[ServiceType]*ServiceTypeOptions{
		SearchQueryService:        newTypeOptions(SearchQueryServiceTypes, ""),
		RegistrationsBaseURL:      newTypeOptions(RegistrationsBaseURLTypes, ""),
		SearchAutocompleteService: newTypeOptions(SearchAutocompleteServiceTypes, ""),
		ReportAbuseURLTemplate: newTypeOptions(
			ReportAbuseTypes,
			"https://www.nuget.org/packages/{id}/{version}/ReportAbuse",
		),
		ReadmeURLTemplate:         newTypeOptions(ReadmeFileURLTypes, ""),
		PackageDetailsURLTemplate: newTypeOptions(PackageDetailsURLTemplateTypes, ""),
		LegacyGallery:             newTypeOptions(LegacyGalleryTypes, ""),
		PackagePublish:            newTypeOptions(PackagePublishTypes, ""),
		PackageBaseAddress:        newTypeOptions(PackageBaseAddressTypes, ""),
		RepositorySignatures:      newTypeOptions(RepositorySignaturesTypes, ""),
		SymbolPackagePublish:      newTypeOptions(SymbolPackagePublishTypes, ""),
		VulnerabilityInfo:         newTypeOptions(VulnerabilityInfoTypes, ""),
		OwnerDetailsURLTemplate:   newTypeOptions(OwnerDetailsURLTemplateTypes, ""),
	}
}

// newTypeOptions creates a new ServiceTypeOptions instance with the given types and default URL.
func newTypeOptions(types ServiceTypes, defaultURL string) *ServiceTypeOptions {
	return &ServiceTypeOptions{
		Types:      types,
		DefaultURL: defaultURL,
	}
}

// SearchFilterType The type of filter to apply to the search.
type SearchFilterType int

// SearchOrderBy Order the resulting packages by the specified field.
type SearchOrderBy int

const (
	// IsLatestVersion Only select the latest stable version of a package per package ID. Given the server supports
	// IsAbsoluteLatestVersion,
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
	PackageExtension          = ".nupkg"
	SnupkgExtension           = ".snupkg"
	SymbolsExtension          = ".symbols" + PackageExtension
	NuGetSymbolHostName       = "nuget.smbsrc.net"
	ServiceEndpoint           = "/api/v2/package"
	DefaultGalleryServerURL   = "https://www.nuget.org"
	TempApiKeyServiceEndpoint = "create-verification-key/%s/%s"
)
