// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"archive/zip"
	"bytes"
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
	idRegex          = regexp.MustCompile(`(?i)^\w+([.-]\w+)*$`)
	defaultURL       *url.URL
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
	Version *nuget.NuGetVersion
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
		return p.Version.Equal(other.Version.Version)
	default:
		return false
	}
}

type PackageBuilder struct {
	includeEmptyDirectories bool
	deterministic           bool
	logger                  *log.Logger
	Id                      string

	Version *nuget.NuGetVersion

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

	PackageTypes []PackageType

	MinClientVersion *nuget.NuGetVersion
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
		PackageTypes:              make([]PackageType, 0),
		Authors:                   make([]string, 0),
		Owners:                    make([]string, 0),
		TargetFrameworks:          make([]*Framework, 0),
		Properties:                make(map[string]string),
	}
}

func (p *PackageBuilder) Save(reader io.Reader) error {
	// Make sure we're saving a valid package id
	if err := ValidatePackageId(p.Id); err != nil {
		return err
	}
	if p.Files == nil || len(p.Files) == 0 || nuget.Some(p.DependencyGroups, func(group *PackageDependencyGroup) bool {
		return group.Packages == nil || len(group.Packages) == 0
	}) {
		return fmt.Errorf("cannot create a package that has no dependencies nor content")
	}
	if errs := p.validate(); len(errs) != 0 {
		return errors.Join(errs...)
	}

	return nil
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
	path := fmt.Sprintf("%s%s", p.Id, nuget.NuspecExtension)
	if err := p.writeOpcManifestRelationship(zipWriter, path, psmdcpPath); err != nil {
		return err
	}
	if relsEntry, err := createEntry(zipWriter, path, p.deterministic); err != nil {
		return err
	} else {

		_, err = io.Copy(relsEntry, nil)
		return err
	}
}

func (p *PackageBuilder) writeOpcManifestRelationship(zipWriter *zip.Writer, path, psmdcpPath string) error {
	var buf bytes.Buffer

	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	buf.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` + "\n")

	target1 := "/" + strings.TrimPrefix(path, "/")
	id1 := generateRelationshipId(target1)
	buf.WriteString(fmt.Sprintf(
		`  <Relationship Type="http://schemas.microsoft.com/packaging/2010/07/manifest" Target="%s" Id="%s"/>\n`,
		xmlEscape(target1),
		xmlEscape(id1),
	))

	target2 := "/" + strings.TrimPrefix(psmdcpPath, "/")
	id2 := generateRelationshipId(target2)
	buf.WriteString(fmt.Sprintf(
		`  <Relationship Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="%s" Id="%s"/>\n`,
		xmlEscape(target2),
		xmlEscape(id2),
	))
	buf.WriteString(`</Relationships>`)
	if relsEntry, err := createEntry(zipWriter, "_rels/.rels", p.deterministic); err != nil {
		return err
	} else {
		_, err = io.Copy(relsEntry, &buf)
		return err
	}
}

func (p *PackageBuilder) writeOpcContentTypes(
	zipWriter *zip.Writer,
	extensions, filesWithoutExtensions map[string]struct{},
) error {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	buf.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` + "\n")
	buf.WriteString(
		`  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>` + "\n",
	)
	buf.WriteString(
		`  <Default Extension="psmdcp" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>` + "\n",
	)
	for ext := range extensions {
		buf.WriteString(`  <Default Extension="` + xmlEscape(ext) + `" ContentType="application/octet"/>` + "\n")
	}
	for file := range filesWithoutExtensions {
		partName := file
		if !strings.HasPrefix(partName, "/") {
			partName = "/" + partName
		}
		buf.WriteString(`  <Override PartName="` + xmlEscape(partName) + `" ContentType="application/octet"/>` + "\n")
	}
	buf.WriteString(`</Types>`)
	if relsEntry, err := createEntry(zipWriter, "[Content_Types].xml", p.deterministic); err != nil {
		return err
	} else {
		_, err = io.Copy(relsEntry, &buf)
		return err
	}
}

// writeOpcPackageProperties OPC backwards compatibility for package properties
func (p *PackageBuilder) writeOpcPackageProperties(zipWriter *zip.Writer, psmdcpPath string) error {
	var buf bytes.Buffer

	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>\n`)
	buf.WriteString(
		`<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">\n`,
	)

	buf.WriteString(fmt.Sprintf("  <dc:creator>%s</dc:creator>\n", xmlEscape(strings.Join(p.Authors, ", "))))
	buf.WriteString(fmt.Sprintf("  <dc:description>%s</dc:description>\n", xmlEscape(p.Description)))
	buf.WriteString(fmt.Sprintf("  <dc:identifier>%s</dc:identifier>\n", xmlEscape(p.Id)))
	buf.WriteString(fmt.Sprintf("  <cp:version>%s</cp:version>\n", xmlEscape(p.Version.String())))
	buf.WriteString(fmt.Sprintf("  <cp:keywords>%s</cp:keywords>\n", xmlEscape(strings.Join(p.Tags, ""))))
	buf.WriteString(fmt.Sprintf("  <cp:lastModifiedBy>%s</cp:lastModifiedBy>\n", xmlEscape("")))

	buf.WriteString(`</cp:coreProperties>`)

	header := &zip.FileHeader{
		Name:   psmdcpPath,
		Method: zip.Deflate,
	}
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, &buf)
	return err
}

func (p *PackageBuilder) writeFiles(zipWriter *zip.Writer, filesWithoutExtensions []string) ([]string, error) {
	extensions := make([]string, 0)
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
			extensions = append(extensions, fileExtension[1:])
		} else {
			filesWithoutExtensions = append(filesWithoutExtensions, "/"+strings.ReplaceAll(file.GetPath(), "\\", "/"))
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
