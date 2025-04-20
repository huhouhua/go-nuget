// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"net/http"
	"strings"
)

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

type ServiceIndex struct {
	Version   string          `json:"version"`
	Resources []*Resource     `json:"resources"`
	Context   *ServiceContext `json:"@context"`
}

type Resource struct {
	Id            string `json:"@id"`
	Type          string `json:"@type"`
	Comment       string `json:"comment"`
	ClientVersion string `json:"clientVersion"`
}

type ServiceContext struct {
	Vocab   string `json:"@vocab"`
	Comment string `json:"comment"`
}

type ServiceResource struct {
	client *Client
}

// GetIndex retrieves the service resources from the NuGet server.
func (s *ServiceResource) GetIndex(options ...RequestOptionFunc) (*ServiceIndex, *http.Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "/v3/index.json", nil, options)
	if err != nil {
		return nil, nil, err
	}
	var svc ServiceIndex
	resp, err := s.client.Do(req, &svc, DecoderTypeJSON)
	if err != nil {
		return nil, resp, err
	}
	return &svc, resp, nil
}
