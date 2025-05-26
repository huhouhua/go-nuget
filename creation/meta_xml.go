// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"encoding/xml"
	"strconv"
	"strings"

	"github.com/huhouhua/go-nuget"
)

// ToXML converts metadata  to XML
func (p *PackageBuilder) ToXML(ns string, generateBackwardsCompatible bool) ([]xml.Token, error) {
	var tokens []xml.Token
	elem := xml.StartElement{Name: xml.Name{Local: "metadata"}, Attr: []xml.Attr{
		{Name: xml.Name{Local: "xmlns"}, Value: ns},
	}}
	if p.MinClientVersion.String() != "" {
		elem.Attr = append(
			elem.Attr,
			xml.Attr{Name: xml.Name{Local: "minClientVersion"}, Value: p.MinClientVersion.String()},
		)
	}
	tokens = append(tokens, elem)
	tokens = append(tokens, NewElement(ns, "id", p.Id)...)
	if strings.TrimSpace(p.Version.String()) != "" {
		tokens = append(tokens, NewElement(ns, "version", p.Version.String())...)
	}
	if strings.TrimSpace(p.Title) == "" {
		tokens = append(tokens, NewElement(ns, "title", p.Title)...)
	}
	if !p.isHasSymbolsInPackageType() {
		if p.EmitRequireLicenseAcceptance {
			tokens = append(
				tokens,
				NewElement(ns, "requireLicenseAcceptance", strconv.FormatBool(p.RequireLicenseAcceptance))...)
		}
		if p.LicenseMetadata != nil {
			tokens = append(tokens, getXMLElementFromLicenseMetadata(ns, p.LicenseMetadata)...)
			if licenseURL, err := p.LicenseMetadata.GetLicenseURL(); err != nil {
				return nil, err
			} else {
				tokens = append(tokens, NewElement(ns, "licenseUrl", licenseURL.String())...)
			}
		}
		if strings.TrimSpace(p.Icon) != "" {
			tokens = append(tokens, NewElement(ns, "icon", p.Icon)...)
		}
		if strings.TrimSpace(p.Readme) != "" {
			tokens = append(tokens, NewElement(ns, "readme", p.Readme)...)
		}
	}
	if p.ProjectURL != nil && p.ProjectURL.String() != "" {
		tokens = append(tokens, NewElement(ns, "projectUrl", p.ProjectURL.String())...)
	}
	if p.IconURL != nil && p.IconURL.String() != "" {
		tokens = append(tokens, NewElement(ns, "iconUrl", p.IconURL.String())...)
	}
	if strings.TrimSpace(p.Description) != "" {
		tokens = append(tokens, NewElement(ns, "description", p.Description)...)
	}
	if strings.TrimSpace(p.Summary) != "" {
		tokens = append(tokens, NewElement(ns, "summary", p.Summary)...)
	}
	if strings.TrimSpace(p.ReleaseNotes) != "" {
		tokens = append(tokens, NewElement(ns, "releaseNotes", p.ReleaseNotes)...)
	}
	if strings.TrimSpace(p.Copyright) != "" {
		tokens = append(tokens, NewElement(ns, "copyright", p.Copyright)...)
	}
	if strings.TrimSpace(p.Language) != "" {
		tokens = append(tokens, NewElement(ns, "language", p.Language)...)
	}
	if len(p.Tags) > 0 {
		tokens = append(tokens, NewElement(ns, "tags", strings.Join(p.Tags, " "))...)
	}
	if p.Serviceable {
		tokens = append(tokens, NewElement(ns, "serviceable", strconv.FormatBool(p.Serviceable))...)
	}
	if p.PackageTypes != nil && len(p.PackageTypes) > 0 {
		tokens = append(tokens, getXElementFromManifestPackageTypes(ns, p.PackageTypes))
	}
	if repoElement := getXElementFromManifestRepository(ns, p.Repository); repoElement != nil {
		tokens = append(tokens)
	}
	return tokens, nil
}

func NewElement(ns, name, value string, attrs ...xml.Attr) []xml.Token {
	return []xml.Token{
		xml.StartElement{
			Name: xml.Name{Space: ns, Local: name},
			Attr: attrs,
		},
		xml.CharData(value),
		xml.EndElement{Name: xml.Name{Space: ns, Local: name}},
	}
}
func getXMLElementFromLicenseMetadata(ns string, meta *LicenseMetadata) []xml.Token {
	attrs := []xml.Attr{
		NewXMLAttr("type", strconv.Itoa(int(meta.GetLicenseType()))),
	}
	if !meta.GetVersion().Equal(nuget.EmptyVersion.Version) {
		attrs = append(attrs, NewXMLAttr("version", meta.GetVersion().String()))
	}
	return NewElement(ns, "license", meta.GetLicense(), attrs...)
}
func getXElementFromManifestRepository(ns string, repository *nuget.RepositoryMetadata) []xml.Token {
	if repository == nil {
		return nil
	}
	var attrs []xml.Attr
	if strings.TrimSpace(repository.Type) != "" {
		attrs = append(attrs, NewXMLAttr("type", repository.URL))
	}
	if strings.TrimSpace(repository.URL) != "" {
		attrs = append(attrs, NewXMLAttr("url", repository.URL))
	}
	if strings.TrimSpace(repository.Branch) != "" {
		attrs = append(attrs, NewXMLAttr("branch", repository.Branch))
	}
	if strings.TrimSpace(repository.Commit) != "" {
		attrs = append(attrs, NewXMLAttr("commit", repository.Commit))
	}
	if len(attrs) > 0 {
		return NewElement(ns, "repository", "", attrs...)
	}
	return nil
}
func getXElementFromManifestPackageTypes(ns string, packageTypes []*PackageType) []xml.Token {
	packageTypesElement := NewElement(ns, "packageTypes", "")
	for _, packageType := range packageTypes {
		packageTypesElement = append(packageTypesElement, getXElementFromManifestPackageType(ns, packageType))
	}
	return packageTypesElement
}
func getXElementFromManifestPackageType(ns string, packageType *PackageType) []xml.Token {
	var attrs []xml.Attr
	if strings.TrimSpace(packageType.Name) != "" {
		attrs = append(attrs, NewXMLAttr("name", packageType.Name))
	}
	if !packageType.Version.Equal(nuget.EmptyVersion.Version) {
		attrs = append(attrs, NewXMLAttr("version", packageType.Version.String()))
	}
	return NewElement(ns, "packageType", "", attrs...)
}
func NewXMLAttr(name string, value string) xml.Attr {
	return xml.Attr{Name: xml.Name{Local: name}, Value: value}
}
