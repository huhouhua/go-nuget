// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"slices"
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
	//if len(p.Files)==0 &&  {
	//
	//}
	return nil
}

func (p *PackageBuilder) ValidateReferenceAssemblies(
	files []PackageFile,
	packageAssemblyReferences []*PackageReferenceSet,
) error {
	frameworks := make([]*Framework, 0)
	for _, group := range packageAssemblyReferences {
		frameworks = append(frameworks, group.TargetFramework)
	}
	if err := validatorPlatformVersion(frameworks); err != nil {
		return err
	}
	libFiles := make([]string, 0)
	for _, file := range files {
		fp := file.GetPath()
		if strings.TrimSpace(fp) != "" && strings.HasPrefix(strings.ToLower(fp), "lib") {
			libFiles = append(libFiles, filepath.Base(fp))
		}
	}
	for _, group := range packageAssemblyReferences {
		for _, reference := range group.References {
			if !contains(libFiles, reference) && !contains(libFiles, reference+".dll") &&
				!contains(libFiles, reference+".exe") && !contains(libFiles, reference+".winmd") {
				return fmt.Errorf(
					"invalid assembly reference '%s'. Ensure that a file named '%s' exists in the lib directory",
					reference,
					reference,
				)
			}
		}
	}
	return nil
}

func (p *PackageBuilder) validateFrameworkAssemblies() error {
	frameworks := make([]*Framework, 0)
	for _, group := range p.FrameworkReferences {
		frameworks = append(frameworks, group.SupportedFrameworks...)
	}
	if err := validatorPlatformVersion(frameworks); err != nil {
		return err
	}
	frameworks = frameworks[:0]
	for _, group := range p.FrameworkReferenceGroups {
		frameworks = append(frameworks, group.TargetFramework)
	}
	if err := validatorPlatformVersion(frameworks); err != nil {
		return err
	}
	return nil
}

func contains(slice []string, item string) bool {
	return nuget.Some(slice, func(s string) bool {
		return strings.Contains(s, item)
	})
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

func (p *PackageBuilder) validateFilesUnique() error {
	seen := make(map[string]bool)
	duplicates := make(map[string]bool)
	for _, file := range p.Files {
		if strings.TrimSpace(file.GetPath()) == "" {
			continue
		}
		destination := nuget.GetPathWithDirectorySeparator(file.GetPath())
		if seen[destination] {
			duplicates[destination] = true
		} else {
			seen[destination] = true
		}
	}
	if len(duplicates) > 0 {
		return fmt.Errorf(
			"attempted to pack multiple files into the same location(s). The following destinations were used multiple times: %s",
			strings.Join(slices.Sorted(maps.Keys(duplicates)), ", "),
		)
	}
	return nil
}

func (p *PackageBuilder) validateDependencies() error {
	targetFramework := make([]*Framework, 0)
	for _, group := range p.DependencyGroups {
		for _, dep := range group.Packages {
			if err := ValidatePackageId(dep.Id); err != nil {
				return err
			}
		}
		targetFramework = append(targetFramework, group.TargetFramework)
	}
	if err := validatorPlatformVersion(targetFramework); err != nil {
		return err
	}
	if p.Version == nil {
		// We have independent validation for null-versions.
		return nil
	}
	return nil
}

func validatorPlatformVersion(frameworks []*Framework) error {
	platformVersions := make([]string, 0)
	for _, framework := range frameworks {
		if framework != nil && strings.TrimSpace(framework.Platform) != "" &&
			framework.PlatformVersion.Equal(nuget.EmptyVersion.Version) {
			platformVersions = append(platformVersions, framework.ShortFolderName)
		}
	}
	if len(platformVersions) > 0 {
		return fmt.Errorf(
			"some dependency group TFMs are missing a platform version: %v",
			strings.Join(platformVersions, ","),
		)
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

func resolveSearchPattern(
	basePath, searchPath, targetPath string,
	includeEmptyDirectories bool,
) ([]*PhysicalPackageFile, error) {
	searchResults, normalizedBasePath, err := nuget.PerformWildcardSearch(basePath, searchPath, includeEmptyDirectories)
	if err != nil {
		return nil, err
	}
	files := make([]*PhysicalPackageFile, 0)
	for _, result := range searchResults {
		file := &PhysicalPackageFile{
			sourcePath: result.Path,
			targetPath: resolvePackagePath(normalizedBasePath, searchPath, result.Path, targetPath),
		}
		if !result.IsFile {
			file.targetPath = path.Join(file.targetPath, nuget.PackageEmptyFileName)
		}
		files = append(files, file)
	}
	return files, nil
}

// resolvePackagePath the path of the file inside a package. For recursive wildcard paths, we preserve the
// path portion beginning
// with the wildcard. For non-recursive wildcard paths, we use the file name from the actual file path on disk.
func resolvePackagePath(searchDirectory, searchPattern, fullPath, targetPath string) string {
	var packagePath string
	isWildcardSearch := strings.Contains(searchPattern, "*")
	isRecursiveWildcardSearch := isWildcardSearch && strings.Contains(searchPattern, "**")
	if (isRecursiveWildcardSearch || isWildcardSearch) && strings.HasPrefix(fullPath, searchDirectory) {
		// The search pattern is recursive. Preserve the non-wildcard portion of the path.
		// e.g. Search: X:\foo\**\*.cs results in SearchDirectory: X:\foo and a file path of X:\foo\bar\biz\boz.cs
		// Truncating X:\foo\ would result in the package path.
		relPath := fullPath[len(searchDirectory):]
		packagePath = strings.TrimLeft(relPath, string(filepath.Separator))
	} else if !isWildcardSearch && strings.EqualFold(path.Ext(searchPattern), path.Ext(targetPath)) {
		// If the search does not contain wild cards, and the target path shares the same extension, copy it
		// e.g. <file src="ie\css\style.css" target="Content\css\ie.css" /> --> Content\css\ie.css
		return targetPath
	} else {
		packagePath = path.Base(fullPath)
	}
	return path.Join(targetPath, packagePath)
}

// isKnownFolder Returns true if the path uses a known folder root.
func isKnownFolder(targetPath string) bool {
	if strings.TrimSpace(targetPath) == "" {
		return false
	}
	parts := nuget.SplitWithFilter(targetPath, []rune{'\\', '/'})
	if len(parts) > 1 {
		topLevelDirectory := parts[0]
		return nuget.Some(nuget.Known, func(folder nuget.Folder) bool {
			return strings.EqualFold(string(folder), topLevelDirectory)
		})
	}
	return false
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
func createPart(zipWriter *zip.Writer, filePath string,
	sourceStream io.Reader, lastWriteTime time.Time, warningMessage *strings.Builder) error {
	if strings.HasSuffix(strings.ToLower(filePath), nuget.NuspecExtension) {
		return nil
	}
	// Split on '/', '\\', and OS-specific separator (assuming Unix-like system here)
	separators := []string{"/", "\\"}
	for _, sep := range separators {
		filePath = strings.ReplaceAll(filePath, sep, "/")
	}

	// Escape each segment
	segments := strings.Split(filePath, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}

	escapedPath := strings.Join(segments, "/")

	//Create an absolute URI to get the refinement on the relative path
	partURL, err := defaultURL.Parse(escapedPath)
	if err != nil {
		return err
	}
	cleanPath := path.Clean(partURL.Path)

	entryName, err := url.PathUnescape(cleanPath)
	if err != nil {
		return err
	}
	entry, err := createPackageFileEntry(zipWriter, entryName, lastWriteTime, warningMessage)
	if err != nil {
		return err
	}
	_, err = io.Copy(entry, sourceStream)
	return err
}

func createPackageFileEntry(
	zipWriter *zip.Writer,
	entryName string,
	timeOffset time.Time,
	warningMessage *strings.Builder,
) (io.Writer, error) {
	header := &zip.FileHeader{
		Name:     entryName,
		Method:   zip.Deflate,
		Modified: timeOffset,
	}
	if timeOffset.Before(zipFormatMinDate) {
		warningMessage.WriteString(fmt.Sprintf("Timestamp for '%s' (%s) is before minimum. Adjusted to %s.\n",
			entryName, timeOffset.Format("2006-01-02"), zipFormatMinDate.Format("2006-01-02")))
		header.Modified = zipFormatMinDate
	} else if timeOffset.After(zipFormatMaxDate) {
		warningMessage.WriteString(fmt.Sprintf("Timestamp for '%s' (%s) is after maximum. Adjusted to %s.\n",
			entryName, timeOffset.Format("2006-01-02"), zipFormatMaxDate.Format("2006-01-02")))
		header.Modified = zipFormatMaxDate
	}
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return nil, err
	}
	return writer, nil
}
