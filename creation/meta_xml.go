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
	if xmlToken, err := p.ToXML(); err != nil {
		return err
	} else {
		tokens = append(tokens, xmlToken...)
	}
	tokens = append(tokens, xml.EndElement{Name: xml.Name{Space: ns, Local: "package"}})

	if _, err := stream.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>` + "\n")); err != nil {
		return err
	}
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
func (p *PackageBuilder) ToXML() ([]xml.Token, error) {
	var tokens []xml.Token
	elem := xml.StartElement{Name: xml.Name{Local: "metadata"}}
	if p.MinClientVersion != nil && p.MinClientVersion.String() != "" {
		elem.Attr = append(
			elem.Attr,
			xml.Attr{Name: xml.Name{Local: "minClientVersion"}, Value: p.MinClientVersion.String()},
		)
	}
	tokens = append(tokens, elem)
	tokens = append(tokens, NewElement("id", p.Id)...)
	if strings.TrimSpace(p.Version.String()) != "" {
		tokens = append(tokens, NewElement("version", p.Version.String())...)
	}
	if strings.TrimSpace(p.Title) != "" {
		tokens = append(tokens, NewElement("title", p.Title)...)
	}
	if !p.isHasSymbolsInPackageType() {
		if p.Authors != nil && len(p.Authors) > 0 {
			tokens = append(tokens, NewElement("authors", strings.Join(p.Authors, ","))...)
		}
		if p.Owners != nil && len(p.Owners) > 0 {
			tokens = append(tokens, NewElement("owners", strings.Join(p.Owners, ","))...)
		}
		if p.DevelopmentDependency {
			tokens = append(tokens, NewElement("developmentDependency", strconv.FormatBool(p.DevelopmentDependency))...)
		}
		if p.EmitRequireLicenseAcceptance {
			tokens = append(
				tokens,
				NewElement("requireLicenseAcceptance", strconv.FormatBool(p.RequireLicenseAcceptance))...)
		}
		licenseUrlToWrite := p.LicenseURL
		if p.LicenseMetadata != nil {
			tokens = append(tokens, getXMLElementFromLicenseMetadata(p.LicenseMetadata)...)
			if licenseURL, err := p.LicenseMetadata.GetLicenseURL(); err != nil {
				return nil, err
			} else {
				licenseUrlToWrite = licenseURL
			}
		}
		if licenseUrlToWrite != nil {
			tokens = append(tokens, NewElement("licenseUrl", licenseUrlToWrite.String())...)
		}
		if strings.TrimSpace(p.Icon) != "" {
			tokens = append(tokens, NewElement("icon", p.Icon)...)
		}
		if strings.TrimSpace(p.Readme) != "" {
			tokens = append(tokens, NewElement("readme", p.Readme)...)
		}
	}
	if p.ProjectURL != nil && p.ProjectURL.String() != "" {
		tokens = append(tokens, NewElement("projectUrl", p.ProjectURL.String())...)
	}
	if p.IconURL != nil && p.IconURL.String() != "" {
		tokens = append(tokens, NewElement("iconUrl", p.IconURL.String())...)
	}
	if strings.TrimSpace(p.Description) != "" {
		tokens = append(tokens, NewElement("description", p.Description)...)
	}
	if strings.TrimSpace(p.Summary) != "" {
		tokens = append(tokens, NewElement("summary", p.Summary)...)
	}
	if strings.TrimSpace(p.ReleaseNotes) != "" {
		tokens = append(tokens, NewElement("releaseNotes", p.ReleaseNotes)...)
	}
	if strings.TrimSpace(p.Copyright) != "" {
		tokens = append(tokens, NewElement("copyright", p.Copyright)...)
	}
	if strings.TrimSpace(p.Language) != "" {
		tokens = append(tokens, NewElement("language", p.Language)...)
	}
	if len(p.Tags) > 0 {
		tokens = append(tokens, NewElement("tags", strings.Join(p.Tags, " "))...)
	}
	if p.Serviceable {
		tokens = append(tokens, NewElement("serviceable", strconv.FormatBool(p.Serviceable))...)
	}
	if p.PackageTypes != nil && len(p.PackageTypes) > 0 {
		tokens = append(tokens, getXElementFromManifestPackageTypes(p.PackageTypes)...)
	}
	if repoElement := getXElementFromManifestRepository(p.Repository); repoElement != nil {
		tokens = append(tokens, repoElement...)
	}
	if dependencyTokens, err := p.dependencyGroupsToTokens(); err != nil {
		return nil, err
	} else {
		tokens = append(tokens, dependencyTokens...)
	}

	if referencesTokens, err := p.packageAssemblyReferencesToTokens(); err != nil {
		return nil, err
	} else {
		tokens = append(tokens, referencesTokens...)
	}

	if frameworkTokens, err := p.frameworkReferenceGroupsToTokens(); err != nil {
		return nil, err
	} else {
		tokens = append(tokens, frameworkTokens...)
	}

	if assembliesTokens, err := getXElementFromFrameworkAssemblies(p.FrameworkReferences); err != nil {
		return nil, err
	} else {
		tokens = append(tokens, assembliesTokens...)
	}

	tokens = append(tokens, getXElementFromManifestContentFiles(p.ContentFiles)...)

	tokens = append(tokens, xml.EndElement{Name: xml.Name{Local: "metadata"}})
	return tokens, nil
}

func (p *PackageBuilder) dependencyGroupsToTokens() ([]xml.Token, error) {
	return getXElementFromGroupableItemSets(p.DependencyGroups,
		func(set *PackageDependencyGroup) bool {
			isHasDependency := nuget.Some(set.Packages, func(dependency *nuget.Dependency) bool {
				return len(dependency.Exclude) > 0 || len(dependency.Include) > 0
			})
			return set.TargetFramework.IsSpecificFramework() || isHasDependency
		}, func(set *PackageDependencyGroup) (string, error) {
			if set.TargetFramework.IsSpecificFramework() {
				return set.TargetFramework.GetFrameworkString()
			}
			return "", nil
		}, func(set *PackageDependencyGroup) []*nuget.Dependency {
			return set.Packages
		}, getXElementFromPackageDependency, "dependencies", "targetFramework")
}
func (p *PackageBuilder) packageAssemblyReferencesToTokens() ([]xml.Token, error) {
	return getXElementFromGroupableItemSets(p.PackageAssemblyReferences,
		func(set *PackageReferenceSet) bool {
			if set.TargetFramework == nil {
				return false
			}
			return set.TargetFramework.IsSpecificFramework()
		}, func(set *PackageReferenceSet) (string, error) {
			if set.TargetFramework == nil {
				return "", nil
			}
			return set.TargetFramework.GetFrameworkString()
		}, func(set *PackageReferenceSet) []string {
			return set.References
		}, getXElementFromPackageReference, "references", "targetFramework")
}

func (p *PackageBuilder) frameworkReferenceGroupsToTokens() ([]xml.Token, error) {
	return getXElementFromGroupableItemSets(p.FrameworkReferenceGroups,
		func(set *FrameworkReferenceGroup) bool {
			// the TFM is required for framework references
			return true
		}, func(set *FrameworkReferenceGroup) (string, error) {
			return set.TargetFramework.GetFrameworkString()
		}, func(set *FrameworkReferenceGroup) []*FrameworkReference {
			return set.FrameworkReferences
		}, getXElementFromFrameworkReference, "frameworkReferences", "targetFramework")

}

func NewElement(name, value string, attrs ...xml.Attr) []xml.Token {
	tokens := []xml.Token{
		xml.StartElement{
			Name: xml.Name{Local: name},
			Attr: attrs,
		},
	}
	if strings.TrimSpace(value) != "" {
		tokens = append(tokens, xml.CharData(value))
	}
	tokens = append(tokens, xml.EndElement{Name: xml.Name{Local: name}})
	return tokens
}

func getXElementFromGroupableItemSets[TSet any, TItem any](
	objectSets []TSet,
	isGroupable func(TSet) bool,
	getGroupIdentifier func(TSet) (string, error),
	getItems func(TSet) []TItem,
	getXElementFromItem func(item TItem) []xml.Token,
	parentName string,
	identifierAttributeName string,
) ([]xml.Token, error) {
	if objectSets == nil || len(objectSets) == 0 {
		return nil, nil
	}
	var groupableSets, ungroupableSets []TSet
	for _, set := range objectSets {
		if isGroupable(set) {
			groupableSets = append(groupableSets, set)
		} else {
			ungroupableSets = append(ungroupableSets, set)
		}
	}

	childElementMap := make(map[string][]xml.Token)
	if groupableSets == nil || len(groupableSets) == 0 {
		// none of the item sets are groupable, then flatten the items
		for _, set := range objectSets {
			for i, item := range getItems(set) {
				childElementMap[strconv.Itoa(i)] = getXElementFromItem(item)
			}
		}
	} else {
		// move the group with null target framework (if any) to the front just for nicer display in UI
		for i, set := range append(ungroupableSets, groupableSets...) {
			groupStart := xml.StartElement{Name: xml.Name{Local: "group"}}
			groupTokens := []xml.Token{groupStart}

			for _, item := range getItems(set) {
				groupTokens = append(groupTokens, getXElementFromItem(item)...)
			}
			key := ""
			if isGroupable(set) {
				if groupIdentifier, err := getGroupIdentifier(set); err != nil {
					return nil, err
				} else {
					if groupIdentifier != "" {
						groupStart.Attr = append(groupStart.Attr, NewXMLAttr(identifierAttributeName, groupIdentifier))
						groupTokens[0] = groupStart
						key = groupIdentifier
					}
				}
			}
			if key == "" {
				key = strconv.Itoa(i)
			}
			if childToken, ok := childElementMap[key]; ok {
				childToken = append(childToken[:len(childToken)-1], groupTokens[1:]...)
				childToken = append(childToken, xml.EndElement{Name: xml.Name{Local: "group"}})
				childElementMap[key] = childToken
			} else {
				groupTokens = append(groupTokens, xml.EndElement{Name: xml.Name{Local: "group"}})
				childElementMap[key] = groupTokens
			}
		}
	}

	tokens := []xml.Token{
		xml.StartElement{Name: xml.Name{Local: parentName}},
	}
	for _, token := range childElementMap {
		tokens = append(tokens, token...)
	}
	tokens = append(tokens, xml.EndElement{Name: xml.Name{Local: parentName}})
	return tokens, nil
}
func getXElementFromPackageDependency(dependency *nuget.Dependency) []xml.Token {
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
	return NewElement("dependency", "", attrs...)
}
func getXElementFromFrameworkAssemblies(references []*FrameworkAssemblyReference) ([]xml.Token, error) {
	if references == nil || len(references) == 0 {
		return nil, nil
	}
	var childTokens []xml.Token
	for _, reference := range references {
		attrs := []xml.Attr{
			NewXMLAttr("assemblyName", reference.AssemblyName),
		}
		if reference.SupportedFrameworks != nil && len(reference.SupportedFrameworks) > 0 {
			frameworkStrs := make([]string, 0)
			for _, framework := range reference.SupportedFrameworks {
				if framework.IsSpecificFramework() {
					if frameworkString, err := framework.GetFrameworkString(); err != nil {
						return nil, err
					} else {
						frameworkStrs = append(frameworkStrs, frameworkString)
					}
				}
			}
			attrs = append(attrs, NewXMLAttr("targetFramework", strings.Join(frameworkStrs, ", ")))
		}
		childTokens = append(childTokens, NewElement("frameworkAssembly", "", attrs...)...)
	}
	tokens := []xml.Token{
		xml.StartElement{Name: xml.Name{Local: "frameworkAssemblies"}},
	}
	tokens = append(tokens, childTokens...)
	tokens = append(tokens, xml.EndElement{Name: xml.Name{Local: "frameworkAssemblies"}})
	return tokens, nil
}
func getXElementFromManifestContentFiles(contentFiles []*ManifestContentFiles) []xml.Token {
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
			childTokens = append(childTokens, NewElement("files", "", attrs...)...)
		}
	}
	tokens := []xml.Token{
		xml.StartElement{Name: xml.Name{Local: "contentFiles"}},
	}
	tokens = append(tokens, childTokens...)
	tokens = append(tokens, xml.EndElement{Name: xml.Name{Local: "contentFiles"}})
	return tokens
}
func getXElementFromPackageReference(reference string) []xml.Token {
	return NewElement("reference", "", NewXMLAttr("file", reference))
}
func getXElementFromFrameworkReference(frameworkReference *FrameworkReference) []xml.Token {
	return NewElement("frameworkReference", "", NewXMLAttr("name", frameworkReference.Name))
}
func getXMLElementFromLicenseMetadata(meta *LicenseMetadata) []xml.Token {
	attrs := []xml.Attr{
		NewXMLAttr("type", meta.GetLicenseType().String()),
	}
	if !meta.GetVersion().Equal(LicenseEmptyVersion) {
		attrs = append(attrs, NewXMLAttr("version", meta.GetVersion().String()))
	}
	return NewElement("license", meta.GetLicense(), attrs...)
}
func getXElementFromManifestRepository(repository *nuget.RepositoryMetadata) []xml.Token {
	if repository == nil {
		return nil
	}
	var attrs []xml.Attr
	if strings.TrimSpace(repository.Type) != "" {
		attrs = append(attrs, NewXMLAttr("type", repository.Type))
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
		return NewElement("repository", "", attrs...)
	}
	return nil
}
func getXElementFromManifestPackageTypes(packageTypes []*PackageType) []xml.Token {
	var childTokens []xml.Token
	for _, packageType := range packageTypes {
		childTokens = append(childTokens, getXElementFromManifestPackageType(packageType)...)
	}
	tokens := []xml.Token{
		xml.StartElement{Name: xml.Name{Local: "packageTypes"}},
	}
	tokens = append(tokens, childTokens...)
	tokens = append(tokens, xml.EndElement{Name: xml.Name{Local: "packageTypes"}})
	return tokens
}
func getXElementFromManifestPackageType(packageType *PackageType) []xml.Token {
	var attrs []xml.Attr
	if strings.TrimSpace(packageType.Name) != "" {
		attrs = append(attrs, NewXMLAttr("name", packageType.Name))
	}
	if !packageType.Version.Equal(nuget.EmptyVersion) {
		attrs = append(attrs, NewXMLAttr("version", packageType.Version.String()))
	}
	return NewElement("packageType", "", attrs...)
}
func NewXMLAttr(name string, value string) xml.Attr {
	return xml.Attr{Name: xml.Name{Local: name}, Value: value}
}
