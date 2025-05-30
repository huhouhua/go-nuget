// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"

	"github.com/huhouhua/go-nuget"
)

func (p *PackageBuilder) save(stream io.Writer, ns string) error {
	var tokens []xml.Token

	tokens = append(tokens, xml.StartElement{Name: xml.Name{Space: ns, Local: "package"}})
	if xmlToken, err := p.ToXML(ns); err != nil {
		return err
	} else {
		tokens = append(tokens, xmlToken...)
	}
	tokens = append(tokens, xml.EndElement{Name: xml.Name{Space: ns, Local: "package"}})

	// Write XML
	encoder := xml.NewEncoder(stream)
	encoder.Indent("", "  ")
	for _, token := range tokens {
		if err := encoder.EncodeToken(token); err != nil {
			return err
		}
	}
	return encoder.Flush()
}

// ToXML converts metadata  to XML
func (p *PackageBuilder) ToXML(ns string) ([]xml.Token, error) {
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

	tokens = append(tokens, p.dependencyGroupsToTokens(ns)...)
	tokens = append(tokens, p.packageAssemblyReferencesToTokens(ns)...)
	tokens = append(tokens, p.frameworkReferenceGroupsToTokens(ns)...)
	tokens = append(tokens, getXElementFromFrameworkAssemblies(ns, p.FrameworkReferences)...)
	tokens = append(tokens, getXElementFromManifestContentFiles(ns, p.ContentFiles)...)

	return tokens, nil
}

func (p *PackageBuilder) dependencyGroupsToTokens(ns string) []xml.Token {
	dependencyGroupsTokens := getXElementFromGroupableItemSets(ns, p.DependencyGroups,
		func(set *PackageDependencyGroup) bool {
			isHasDependency := nuget.Some(set.Packages, func(dependency *nuget.Dependency) bool {
				return len(dependency.Exclude) > 0 || len(dependency.Include) > 0
			})
			return set.TargetFramework.IsSpecificFramework || isHasDependency
		}, func(set *PackageDependencyGroup) string {
			if set.TargetFramework.IsSpecificFramework {
				return set.TargetFramework.GetFrameworkString()
			}
			return ""
		}, func(set *PackageDependencyGroup) []*nuget.Dependency {
			return set.Packages
		}, getXElementFromPackageDependency, "dependencies", "targetFramework")

	return dependencyGroupsTokens
}
func (p *PackageBuilder) packageAssemblyReferencesToTokens(ns string) []xml.Token {
	packageAssemblyReferencesTokens := getXElementFromGroupableItemSets(ns, p.PackageAssemblyReferences,
		func(set *PackageReferenceSet) bool {
			if set.TargetFramework == nil {
				return false
			}
			return set.TargetFramework.IsSpecificFramework
		}, func(set *PackageReferenceSet) string {
			if set.TargetFramework == nil {
				return ""
			}
			return set.TargetFramework.GetFrameworkString()
		}, func(set *PackageReferenceSet) []string {
			return set.References
		}, getXElementFromPackageReference, "references", "targetFramework")

	return packageAssemblyReferencesTokens
}

func (p *PackageBuilder) frameworkReferenceGroupsToTokens(ns string) []xml.Token {
	frameworkReferenceGroupsTokens := getXElementFromGroupableItemSets(ns, p.FrameworkReferenceGroups,
		func(set *FrameworkReferenceGroup) bool {
			// the TFM is required for framework references
			return true
		}, func(set *FrameworkReferenceGroup) string {
			return set.TargetFramework.GetFrameworkString()
		}, func(set *FrameworkReferenceGroup) []*FrameworkReference {
			return set.FrameworkReferences
		}, getXElementFromFrameworkReference, "frameworkReferences", "targetFramework")

	return frameworkReferenceGroupsTokens
}

func NewElement(ns, name, value string, attrs ...xml.Attr) []xml.Token {
	tokens := []xml.Token{
		xml.StartElement{
			Name: xml.Name{Space: ns, Local: name},
			Attr: attrs,
		},
	}
	if strings.TrimSpace(value) != "" {
		tokens = append(tokens, xml.CharData(value))
	}
	tokens = append(tokens, xml.EndElement{Name: xml.Name{Space: ns, Local: name}})
	return tokens
}

func getXElementFromGroupableItemSets[TSet any, TItem any](
	ns string,
	objectSets []TSet,
	isGroupable func(TSet) bool,
	getGroupIdentifier func(TSet) string,
	getItems func(TSet) []TItem,
	getXElementFromItem func(ns string, item TItem) []xml.Token,
	parentName string,
	identifierAttributeName string,
) []xml.Token {
	if objectSets == nil || len(objectSets) == 0 {
		return nil
	}
	var groupableSets, ungroupableSets []TSet
	for _, set := range objectSets {
		if isGroupable(set) {
			groupableSets = append(groupableSets, set)
		} else {
			ungroupableSets = append(ungroupableSets, set)
		}
	}

	var childElements []xml.Token
	if groupableSets == nil || len(groupableSets) == 0 {
		// none of the item sets are groupable, then flatten the items
		for _, set := range objectSets {
			for _, item := range getItems(set) {
				childElements = append(childElements, getXElementFromItem(ns, item)...)
			}
		}
	} else {
		// move the group with null target framework (if any) to the front just for nicer display in UI
		for _, set := range append(ungroupableSets, groupableSets...) {
			groupStart := xml.StartElement{Name: xml.Name{Space: ns, Local: "group"}}
			groupTokens := []xml.Token{groupStart}

			for _, item := range getItems(set) {
				groupTokens = append(groupTokens, getXElementFromItem(ns, item)...)
			}
			if isGroupable(set) {
				groupIdentifier := getGroupIdentifier(set)
				if groupIdentifier != "" {
					groupStart.Attr = append(groupStart.Attr, NewXMLAttr(identifierAttributeName, groupIdentifier))
					groupTokens[0] = groupStart
				}
			}
			groupTokens = append(groupTokens, xml.EndElement{Name: xml.Name{Space: ns, Local: "group"}})
			childElements = append(childElements, groupTokens...)
		}
	}

	tokens := []xml.Token{
		xml.StartElement{Name: xml.Name{Space: ns, Local: parentName}},
	}
	tokens = append(tokens, childElements...)
	tokens = append(tokens, xml.EndElement{Name: xml.Name{Space: ns, Local: parentName}})
	return tokens
}
func getXElementFromPackageDependency(ns string, dependency *nuget.Dependency) []xml.Token {
	if dependency == nil {
		return nil
	}
	attrs := []xml.Attr{
		NewXMLAttr("id", dependency.Id),
	}
	if dependency.VersionRange != nil && dependency.VersionRange != nuget.All {
		attrs = append(attrs, NewXMLAttr("version", dependency.VersionRange.String()))
	}
	if dependency.Include != nil && len(dependency.Include) > 0 {
		attrs = append(attrs, NewXMLAttr("include", strings.Join(dependency.Include, ",")))
	}
	if dependency.Exclude != nil && len(dependency.Exclude) > 0 {
		attrs = append(attrs, NewXMLAttr("exclude", strings.Join(dependency.Exclude, ",")))
	}
	return NewElement(ns, "dependency", "", attrs...)
}
func getXElementFromFrameworkAssemblies(ns string, references []*FrameworkAssemblyReference) []xml.Token {
	if references == nil || len(references) == 0 {
		return nil
	}
	var childTokens []xml.Token
	for _, reference := range references {
		attrs := []xml.Attr{
			NewXMLAttr("assemblyName", reference.AssemblyName),
		}
		if reference.SupportedFrameworks != nil && len(reference.SupportedFrameworks) > 0 {
			frameworkStrs := make([]string, 0)
			for _, framework := range reference.SupportedFrameworks {
				if framework.IsSpecificFramework {
					frameworkStrs = append(frameworkStrs, framework.GetFrameworkString())
				}
			}
			attrs = append(attrs, NewXMLAttr("targetFramework", strings.Join(frameworkStrs, ", ")))
		}
		childTokens = append(childTokens, NewElement(ns, "frameworkAssembly", "", attrs...))
	}
	tokens := []xml.Token{
		xml.StartElement{Name: xml.Name{Space: ns, Local: "frameworkAssemblies"}},
	}
	tokens = append(tokens, childTokens...)
	tokens = append(tokens, xml.EndElement{Name: xml.Name{Space: ns, Local: "frameworkAssemblies"}})
	return tokens
}
func getXElementFromManifestContentFiles(ns string, contentFiles []*ManifestContentFiles) []xml.Token {
	if contentFiles == nil || len(contentFiles) == 0 {
		return nil
	}
	var childTokens []xml.Token
	for _, file := range contentFiles {
		var attrs []xml.Attr
		if strings.TrimSpace(file.Include) != "" {
			attrs = append(attrs, NewXMLAttr("include", file.Include))
		}
		if strings.TrimSpace(file.Exclude) != "" {
			attrs = append(attrs, NewXMLAttr("exclude", file.Exclude))
		}
		if strings.TrimSpace(file.BuildAction) != "" {
			attrs = append(attrs, NewXMLAttr("buildAction", file.BuildAction))
		}
		if strings.TrimSpace(file.CopyToOutput) != "" {
			attrs = append(attrs, NewXMLAttr("copyToOutput", file.CopyToOutput))
		}
		if strings.TrimSpace(file.Flatten) != "" {
			attrs = append(attrs, NewXMLAttr("flatten", file.Flatten))
		}
		if attrs != nil && len(attrs) > 0 {
			childTokens = append(childTokens, NewElement(ns, "files", "", attrs...))
		}
	}
	tokens := []xml.Token{
		xml.StartElement{Name: xml.Name{Space: ns, Local: "contentFiles"}},
	}
	tokens = append(tokens, childTokens...)
	tokens = append(tokens, xml.EndElement{Name: xml.Name{Space: ns, Local: "contentFiles"}})
	return tokens
}
func getXElementFromPackageReference(ns, reference string) []xml.Token {
	return NewElement(ns, reference, "", NewXMLAttr("file", reference))
}
func getXElementFromFrameworkReference(ns string, frameworkReference *FrameworkReference) []xml.Token {
	return NewElement(ns, "frameworkReference", "", NewXMLAttr("name", frameworkReference.Name))
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
