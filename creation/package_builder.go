// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/huhouhua/go-nuget"
)

// MaxPackageIdLength Max allowed length for package Id.
const MaxPackageIdLength = 100

var (
	idRegex    = regexp.MustCompile(`(?i)^\w+([.-]\w+)*$`)
	defaultURL *url.URL
	// MaxIconFileSize the Maximum Icon file size: 1 megabyte
	MaxIconFileSize  = int64(1024 * 1024)
	zipFormatMinDate = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
	zipFormatMaxDate = time.Date(2107, 12, 31, 23, 59, 58, 0, time.UTC)
	SymbolsPackage   = &PackageType{Name: "SymbolsPackage", Version: nil}
)

func init() {
	defaultURL, _ = url.Parse("http://defaultcontainer/")
}

// IsValidPackageId checks if the package ID is valid using regex.
func IsValidPackageId(packageId string) bool {
	if strings.TrimSpace(packageId) == "" {
		return false
	}
	return idRegex.MatchString(packageId)
}

// ValidatePackageId checks if the package ID is valid and within the allowed length.
// It returns an error if the validation fails.
func ValidatePackageId(packageId string) error {
	if len(packageId) > MaxPackageIdLength {
		return errors.New("id must not exceed 100 characters")
	}
	if !IsValidPackageId(packageId) {
		return fmt.Errorf(
			"the package ID '%s' contains invalid characters. Examples of valid package IDs include 'MyPackage' and 'MyPackage.Sample'",
			packageId,
		)
	}
	return nil
}

// PackageType It is important that this type remains immutable due to the cloning of package specs
type PackageType struct {
	Name    string
	Version *nuget.Version
}

func (p *PackageType) Equals(other *PackageType) bool {
	if other == nil {
		return false
	}
	if !strings.EqualFold(p.Name, other.Name) {
		return false
	}
	switch {
	case p.Version == nil && other.Version == nil:
		return true
	case p.Version != nil && other.Version != nil:
		return p.Version.Semver.Equal(other.Version.Semver)
	default:
		return false
	}
}

type PackageBuilder struct {
	includeEmptyDirectories bool
	deterministic           bool
	logger                  *log.Logger
	Id                      string

	Version *nuget.Version

	Repository *nuget.RepositoryMetadata

	LicenseMetadata *LicenseMetadata

	HasSnapshotVersion bool

	Title string

	Authors []string

	Owners []string

	IconURL *url.URL

	Icon string

	LicenseURL *url.URL

	ProjectURL *url.URL

	RequireLicenseAcceptance bool

	EmitRequireLicenseAcceptance bool

	Serviceable bool

	DevelopmentDependency bool

	Description string

	Summary string

	ReleaseNotes string

	Language string

	OutputName string

	Tags []string

	Readme string

	Properties map[string]string

	Copyright string

	DependencyGroups []*PackageDependencyGroup

	Files []PackageFile

	FrameworkReferences []*FrameworkAssemblyReference

	FrameworkReferenceGroups []*FrameworkReferenceGroup

	TargetFrameworks []*Framework

	// ContentFiles section from the manifest for content v2
	ContentFiles []*ManifestContentFiles

	PackageAssemblyReferences []*PackageReferenceSet

	PackageTypes []*PackageType

	MinClientVersion *nuget.Version
}

func NewPackageBuilder(includeEmptyDirectories, deterministic bool, logger *log.Logger) *PackageBuilder {
	return &PackageBuilder{
		includeEmptyDirectories:   includeEmptyDirectories,
		deterministic:             deterministic,
		logger:                    logger,
		Files:                     make([]PackageFile, 0),
		DependencyGroups:          make([]*PackageDependencyGroup, 0),
		FrameworkReferences:       make([]*FrameworkAssemblyReference, 0),
		FrameworkReferenceGroups:  make([]*FrameworkReferenceGroup, 0),
		ContentFiles:              make([]*ManifestContentFiles, 0),
		PackageAssemblyReferences: make([]*PackageReferenceSet, 0),
		PackageTypes:              make([]*PackageType, 0),
		Authors:                   make([]string, 0),
		Owners:                    make([]string, 0),
		Tags:                      make([]string, 0),
		TargetFrameworks:          make([]*Framework, 0),
		Properties:                make(map[string]string),
	}
}

func (p *PackageBuilder) Save(w io.Writer) error {
	// Make sure we're saving a valid package id
	if err := ValidatePackageId(p.Id); err != nil {
		return err
	}
	isHasDependencyGroups := nuget.Some(p.DependencyGroups, func(group *PackageDependencyGroup) bool {
		return group.Packages == nil || len(group.Packages) == 0
	})

	if p.Files == nil && len(p.Files) == 0 && isHasDependencyGroups &&
		(p.FrameworkReferences == nil || len(p.FrameworkReferences) == 0) &&
		(p.FrameworkReferenceGroups == nil || len(p.FrameworkReferenceGroups) == 0) {
		return fmt.Errorf("cannot create a package that has no dependencies nor content")
	}
	if errs := p.validate(); len(errs) != 0 {
		return errors.Join(errs...)
	}
	writerPackage := zip.NewWriter(w)
	if psmdcp, err := calcPsmdcpName(p.Files, p.deterministic); err != nil {
		return err
	} else {
		// Validate and write the manifest
		psmdcpPath := fmt.Sprintf("package/services/metadata/core-properties/%s.psmdcp", psmdcp)
		if err = p.writeManifest(writerPackage, determineMinimumSchemaVersion(p.Files, p.DependencyGroups), psmdcpPath); err != nil {
			return err
		}
		if err = p.writeOpcPackageProperties(writerPackage, psmdcpPath); err != nil {
			return err
		}
	}
	// Write the files to the package
	filesWithoutExtensions := map[string]bool{}
	if extensions, err := p.writeFiles(writerPackage, filesWithoutExtensions); err != nil {
		return err
	} else {
		extensions["nuspec"] = true
		if err = p.writeOpcContentTypes(writerPackage, extensions, filesWithoutExtensions); err != nil {
			return err
		}
	}
	if err := writerPackage.Flush(); err != nil {
		return err
	}
	return writerPackage.Close()
}

func (p *PackageBuilder) PopulateFiles(basePath string, files []*ManifestFile) error {
	for _, file := range files {
		if err := p.AddFiles(basePath, file.Source, file.Target, file.Exclude); err != nil {
			return err
		}
	}
	return nil
}
func (p *PackageBuilder) AddFiles(basePath, source, destination, exclude string) error {
	exclude = strings.ReplaceAll(exclude, "\\", string(filepath.Separator))
	sourcePattern := strings.ReplaceAll(source, "\\", string(filepath.Separator))
	searchFiles, err := resolveSearchPattern(basePath, sourcePattern, destination, p.includeEmptyDirectories)
	if err != nil {
		return err
	}
	if p.includeEmptyDirectories {
		// we only allow empty directories which are under known root folders.
		searchFiles = nuget.Filter(searchFiles, func(file *PhysicalPackageFile) bool {
			return path.Base(file.targetPath) != nuget.PackageEmptyFileName && isKnownFolder(file.targetPath)
		})
	}
	p.excludeFiles(searchFiles, basePath, exclude)
	if !strings.Contains(source, "*") && !nuget.IsDirectoryPath(source) && len(searchFiles) == 0 &&
		strings.TrimSpace(exclude) == "" {
		return fmt.Errorf("%s file not found", source)
	}
	for _, file := range searchFiles {
		p.Files = append(p.Files, file)
	}
	return nil
}
func (p *PackageBuilder) excludeFiles(searchFiles []*PhysicalPackageFile, basePath, exclude string) {
	if strings.TrimSpace(exclude) == "" {
		return
	}
	exclusions := nuget.SplitWithFilter(exclude, []rune{';'})
	for _, exclusion := range exclusions {
		wildCard := nuget.NormalizeWildcardForExcludedFiles(basePath, exclusion)
		nuget.GetFilteredPackageFiles(&searchFiles, func(file *PhysicalPackageFile) string {
			return file.sourcePath
		}, []string{wildCard})
	}
}

func (p *PackageBuilder) writeManifest(zipWriter *zip.Writer, minimumManifestVersion int, psmdcpPath string) error {
	manifestPath := fmt.Sprintf("%s%s", p.Id, nuget.NuspecExtension)
	if err := p.writeOpcManifestRelationship(zipWriter, manifestPath, psmdcpPath); err != nil {
		return err
	}
	if relsEntry, err := createEntry(zipWriter, manifestPath, p.deterministic); err != nil {
		return err
	} else {
		version := p.GetVersion()
		if minimumManifestVersion > version {
			version = minimumManifestVersion
		}
		if schemaNamespace, err := VersionToSchemaMaps.GetSchemaNamespace(version); err != nil {
			return err
		} else {
			return p.save(relsEntry, schemaNamespace)
		}
	}
}

func (p *PackageBuilder) writeOpcManifestRelationship(zipWriter *zip.Writer, path, psmdcpPath string) error {
	var (
		writer io.Writer
		err    error
	)
	if writer, err = createEntry(zipWriter, "_rels/.rels", p.deterministic); err != nil {
		return err
	}
	if _, err = writer.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")); err != nil {
		return err
	}
	tokens := []xml.Token{
		xml.StartElement{Name: xml.Name{Local: "Relationships"},
			Attr: []xml.Attr{
				NewXMLAttr("xmlns", "http://schemas.openxmlformats.org/package/2006/relationships"),
			}},
	}
	var childTokens []xml.Token

	targetPath := "/" + strings.TrimPrefix(path, "/")
	targetId := generateRelationshipId(targetPath)

	childTokens = append(childTokens, NewElement("Relationship", "",
		NewXMLAttr("Type", "http://schemas.microsoft.com/packaging/2010/07/manifest"),
		NewXMLAttr("Target", xmlEscape(targetPath)),
		NewXMLAttr("Id", xmlEscape(targetId)))...)

	psmdcpTarget := "/" + strings.TrimPrefix(psmdcpPath, "/")
	psmdcpId := generateRelationshipId(psmdcpPath)

	childTokens = append(childTokens, NewElement("Relationship", "",
		NewXMLAttr("Type", "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties"),
		NewXMLAttr("Target", xmlEscape(psmdcpTarget)),
		NewXMLAttr("Id", xmlEscape(psmdcpId)))...)

	tokens = append(tokens, childTokens...)
	tokens = append(tokens, xml.EndElement{Name: xml.Name{Local: "Relationships"}})
	return BuildXml(writer, tokens)
}

func (p *PackageBuilder) writeOpcContentTypes(
	zipWriter *zip.Writer,
	extensions, filesWithoutExtensions map[string]bool,
) error {
	var (
		writer io.Writer
		err    error
	)
	if writer, err = createEntry(zipWriter, "[Content_Types].xml", p.deterministic); err != nil {
		return err
	}
	if _, err = writer.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")); err != nil {
		return err
	}
	tokens := []xml.Token{
		xml.StartElement{Name: xml.Name{Local: "Types"},
			Attr: []xml.Attr{
				NewXMLAttr("xmlns", "http://schemas.openxmlformats.org/package/2006/content-types"),
			}},
	}
	var childTokens []xml.Token
	childTokens = append(childTokens, NewElement("Default", "", NewXMLAttr("Extension", "rels"),
		NewXMLAttr("ContentType", "application/vnd.openxmlformats-package.relationships+xml"))...)

	childTokens = append(childTokens, NewElement("Default", "", NewXMLAttr("Extension", "psmdcp"),
		NewXMLAttr("ContentType", "application/vnd.openxmlformats-package.core-properties+xml"))...)

	for ext := range extensions {
		childTokens = append(childTokens, NewElement("Default", "", NewXMLAttr("Extension", xmlEscape(ext)),
			NewXMLAttr("ContentType", "application/octet"))...)
	}

	for file := range filesWithoutExtensions {
		partName := file
		if !strings.HasPrefix(partName, "/") {
			partName = "/" + partName
		}
		childTokens = append(childTokens, NewElement("Override", "", NewXMLAttr("PartName", xmlEscape(partName)),
			NewXMLAttr("ContentType", "application/octet"))...)
	}

	tokens = append(tokens, childTokens...)
	tokens = append(tokens, xml.EndElement{Name: xml.Name{Local: "Types"}})
	return BuildXml(writer, tokens)
}

// writeOpcPackageProperties OPC backwards compatibility for package properties
func (p *PackageBuilder) writeOpcPackageProperties(zipWriter *zip.Writer, psmdcpPath string) error {
	var (
		writer io.Writer
		err    error
	)
	header := &zip.FileHeader{
		Name:   psmdcpPath,
		Method: zip.Deflate,
	}
	if writer, err = zipWriter.CreateHeader(header); err != nil {
		return err
	}
	if _, err = writer.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")); err != nil {
		return err
	}
	tokens := []xml.Token{
		xml.StartElement{Name: xml.Name{Local: "cp:coreProperties"},
			Attr: []xml.Attr{
				NewXMLAttr("xmlns:dc", "http://purl.org/dc/elements/1.1/"),
				NewXMLAttr("xmlns:dcterms", "http://purl.org/dc/terms/"),
				NewXMLAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance"),
				NewXMLAttr("xmlns:cp", "http://schemas.openxmlformats.org/package/2006/metadata/core-properties"),
			}},
	}
	var childTokens []xml.Token
	childTokens = append(childTokens, NewElement("dc:creator", xmlEscape(strings.Join(p.Authors, ", ")))...)
	childTokens = append(childTokens, NewElement("dc:description", xmlEscape(p.Description))...)
	childTokens = append(childTokens, NewElement("dc:identifier", xmlEscape(p.Id))...)
	childTokens = append(childTokens, NewElement("dc:version", xmlEscape(p.Version.OriginalVersion))...)
	childTokens = append(childTokens, NewElement("dc:keywords", xmlEscape(strings.Join(p.Tags, " ")))...)
	childTokens = append(childTokens, NewElement("dc:lastModifiedBy", xmlEscape(""))...)
	tokens = append(tokens, childTokens...)
	tokens = append(tokens, xml.EndElement{Name: xml.Name{Local: "cp:coreProperties"}})

	// Write XML
	return BuildXml(writer, tokens)
}

func (p *PackageBuilder) writeFiles(
	zipWriter *zip.Writer,
	filesWithoutExtensionsMap map[string]bool,
) (map[string]bool, error) {
	extensions := make(map[string]bool)
	warningMessage := &strings.Builder{}

	// Add files that might not come from expanding files on disk
	for _, file := range p.Files {
		stream, err := file.GetStream()
		if err != nil {
			return nil, err
		}
		lastWriteTime := file.GetLastWriteTime()
		if p.deterministic {
			lastWriteTime = zipFormatMinDate
		}
		err = createPart(zipWriter, file.GetPath(), stream, lastWriteTime, warningMessage)
		if err != nil {
			return nil, err
		}
		if fileExtension := path.Ext(file.GetPath()); strings.TrimSpace(fileExtension) != "" {
			withoutExtension := fileExtension[1:]
			if !extensions[withoutExtension] {
				extensions[withoutExtension] = true
			}
		} else {
			withoutExtension := "/" + strings.ReplaceAll(file.GetPath(), "\\", "/")
			if !filesWithoutExtensionsMap[withoutExtension] {
				filesWithoutExtensionsMap[withoutExtension] = true
			}
		}
	}
	var warningMessageString = warningMessage.String()
	if strings.TrimSpace(warningMessageString) != "" {
		p.logger.Printf(
			"The zip format supports a limited date range. The following files are outside the supported range \n %s \n",
			warningMessageString,
		)
	}
	return extensions, nil
}

func (p *PackageBuilder) GetVersion() int {
	if p.PackageAssemblyReferences != nil {
		referencesHasTargetFramework := nuget.Some(p.PackageAssemblyReferences, func(set *PackageReferenceSet) bool {
			return set.TargetFramework != nil && set.TargetFramework.IsSpecificFramework()
		})
		if referencesHasTargetFramework {
			return TargetFrameworkSupportForReferencesVersion
		}
	}
	if p.DependencyGroups != nil {
		dependencyHasTargetFramework := nuget.Some(p.DependencyGroups, func(group *PackageDependencyGroup) bool {
			return group.TargetFramework != nil && group.TargetFramework.IsSpecificFramework()
		})
		if dependencyHasTargetFramework {
			return TargetFrameworkSupportForDependencyContentsAndToolsVersion
		}
	}
	if p.Version != nil && p.Version.Semver.Prerelease() != "" {
		return SemverVersion
	}
	return DefaultVersion
}
