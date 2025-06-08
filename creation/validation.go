// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/huhouhua/go-nuget"
)

func (p *PackageBuilder) validate() []error {
	var errs []error
	errs = append(errs, p.validateDependencies())
	errs = append(errs, p.validateFilesUnique())
	errs = append(errs, p.ValidateReferenceAssemblies())
	errs = append(errs, p.validateFrameworkAssemblies())
	errs = append(errs, p.validateLicenseFile())
	errs = append(errs, p.validateIconFile())
	errs = append(errs, p.validateReadmeFile())
	errs = append(errs, p.validateDependencyGroups())
	errs = append(errs, p.validateManifest())

	return nuget.Filter(errs, func(err error) bool {
		return err != nil
	})
}
func (p *PackageBuilder) validateManifest() error {
	var results []string
	results = append(results, p.validateArgs()...)
	for _, reference := range p.PackageAssemblyReferences {
		results = append(results, reference.Validate()...)
	}
	if len(results) == 0 {
		return nil
	}
	return errors.New(strings.Join(results, "\n"))
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

func (p *PackageBuilder) ValidateReferenceAssemblies() error {
	frameworks := make([]*Framework, 0)
	for _, group := range p.PackageAssemblyReferences {
		frameworks = append(frameworks, group.TargetFramework)
	}
	if err := validatorPlatformVersion(frameworks); err != nil {
		return err
	}
	libFiles := make([]string, 0)
	for _, file := range p.Files {
		fp := file.GetPath()
		if strings.TrimSpace(fp) != "" && strings.HasPrefix(strings.ToLower(fp), "lib") {
			libFiles = append(libFiles, filepath.Base(fp))
		}
	}
	for _, group := range p.PackageAssemblyReferences {
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

func (p *PackageBuilder) validateFilesUnique() error {
	seen := make(map[string]bool)
	duplicates := make(map[string]bool)
	for _, file := range p.Files {
		if strings.TrimSpace(file.GetPath()) == "" {
			continue
		}
		destination := getPathWithDirectorySeparator(file.GetPath(), os.PathSeparator)
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

func (p *PackageBuilder) validateLicenseFile() error {
	if p.isHasSymbolsInPackageType() || p.LicenseMetadata == nil || p.LicenseMetadata.GetLicenseType() != nuget.File {
		return nil
	}
	ext := path.Ext(p.LicenseMetadata.GetLicense())
	if strings.TrimSpace(ext) != "" &&
		!strings.EqualFold(ext, ".txt") &&
		!strings.EqualFold(ext, nuget.ReadmeExtension) {
		return fmt.Errorf(
			"the license file '%s' has an invalid extension. Valid options are .txt, .md or none",
			p.LicenseMetadata.GetLicense(),
		)
	}
	var licenseFilePathWithIncorrectCase *string
	if findFileInPackage(p.LicenseMetadata.GetLicense(), p.Files, licenseFilePathWithIncorrectCase) == nil {
		if &licenseFilePathWithIncorrectCase == nil {
			return fmt.Errorf("the license file '%s' does not exist in the package", p.LicenseMetadata.GetLicense())
		} else {
			return fmt.Errorf("the license file '%s' does not exist in the package. (Did you mean '%s'?)",
				p.LicenseMetadata.GetLicense(), *licenseFilePathWithIncorrectCase)
		}
	}
	return nil
}

// validateIconFile Given a list of resolved files, determine which file will be used as the icon file and validate its
// size and extension.
func (p *PackageBuilder) validateIconFile() error {
	if p.isHasSymbolsInPackageType() || strings.TrimSpace(p.Icon) == "" {
		return nil
	}
	ext := path.Ext(p.Icon)
	if strings.TrimSpace(ext) == "" || (!strings.EqualFold(ext, ".jpeg") &&
		!strings.EqualFold(ext, ".jpg") &&
		!strings.EqualFold(ext, ".png")) {
		return fmt.Errorf(
			"the 'icon' element '%s' has an invalid file extension. Valid options are .png, .jpg or .jpeg",
			p.Icon,
		)
	}
	var iconPathWithIncorrectCase *string
	iconFile := findFileInPackage(p.Icon, p.Files, iconPathWithIncorrectCase)
	if iconFile == nil {
		if &iconPathWithIncorrectCase == nil {
			return fmt.Errorf("the icon file '%s' does not exist in the package", p.Icon)
		} else {
			return fmt.Errorf("the icon file '%s' does not exist in the package. (Did you mean '%s'?)",
				p.Icon, *iconPathWithIncorrectCase)
		}
	}
	if file, err := iconFile.GetStream(); err != nil {
		return err
	} else {
		if stat, err := file.Stat(); err != nil {
			return err
		} else {
			if stat.Size() > MaxIconFileSize {
				return fmt.Errorf("the icon file size must not exceed 1 megabyte")
			}
			if stat.Size() == 0 {
				return fmt.Errorf("the icon file is empty")
			}
		}
	}
	return nil
}

func (p *PackageBuilder) validateReadmeFile() error {
	if p.isHasSymbolsInPackageType() || strings.TrimSpace(p.Readme) == "" {
		return nil
	}
	ext := path.Ext(p.Readme)
	if strings.TrimSpace(ext) != "" || !strings.EqualFold(ext, nuget.ReadmeExtension) {
		return fmt.Errorf("the readme file '%s' has an invalid extension. It must end in .md", p.Readme)
	}
	readmePathStripped := stripLeadingDirectorySeparators(p.Readme)
	readmeFileList := nuget.Filter(p.Files, func(file PackageFile) bool {
		return strings.EqualFold(readmePathStripped, stripLeadingDirectorySeparators(file.GetPath()))
	})
	if len(readmeFileList) == 0 {
		return fmt.Errorf("the readme file '%s' does not exist in the package", p.Readme)
	}
	readmeFile := readmeFileList[0]
	if file, err := readmeFile.GetStream(); err != nil {
		return err
	} else {
		if stat, err := file.Stat(); err != nil {
			return err
		} else {
			if stat.Size() == 0 {
				return fmt.Errorf("the readme file '%s' is empty", p.Readme)
			}
		}
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

func (p *PackageBuilder) validateDependencyGroups() error {
	for _, group := range p.DependencyGroups {
		depSet := make(map[string]struct{}) // use map for uniqueness

		for _, dep := range group.Packages {
			key := strings.ToLower(dep.Id)
			// Throw an error if this dependency has been defined more than once
			if _, exists := depSet[key]; exists {
				return fmt.Errorf(fmt.Sprintf("'%s' already has a dependency defined for '%s'.", p.Id, dep.Id))
			}
			depSet[key] = struct{}{}
			if dep.VersionRange == nil {
				continue
			}
			if dep.VersionRange.MinVersion != nil && dep.VersionRange.MaxVersion != nil {
				if (!dep.VersionRange.IncludeMax || !dep.VersionRange.IncludeMin) &&
					dep.VersionRange.MaxVersion.Equal(dep.VersionRange.MinVersion) {
					return fmt.Errorf(fmt.Sprintf("dependency '%s' has an invalid version.", dep.Id))
				}
				if dep.VersionRange.MinVersion.GreaterThan(dep.VersionRange.MaxVersion) {
					return fmt.Errorf("dependency '%s' has an invalid version", dep.Id)
				}
			}
		}
	}
	return nil
}
func (p *PackageBuilder) validateArgs() []string {
	var errs []string
	if strings.TrimSpace(p.Id) == "" {
		errs = append(errs, "id is required.")
	} else {
		if len(p.Id) > MaxPackageIdLength {
			errs = append(errs, "id must not exceed 100 characters.")
		} else if !IsValidPackageId(p.Id) {
			errs = append(errs, fmt.Sprintf("the package ID '%s' contains invalid characters. Examples of valid package IDs include 'MyPackage' and 'MyPackage.Sample'.", p.Id))
		}
	}
	if p.Version == nil {
		errs = append(errs, "version is required.")
	}
	if p.Authors == nil {
		errs = append(errs, "authors is required.")
	}
	isHasEmpty := nuget.Some(p.Authors, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})

	if isHasEmpty && !p.isHasSymbolsInPackageType() {
		errs = append(errs, "authors is required.")
	}
	if strings.TrimSpace(p.Description) == "" {
		errs = append(errs, "description is required.")
	}
	if p.LicenseURL != nil && strings.TrimSpace(p.LicenseURL.String()) == "" {
		errs = append(errs, "licenseURL cannot be empty.")
	}
	if p.IconURL != nil && strings.TrimSpace(p.IconURL.String()) == "" {
		errs = append(errs, "iconURL cannot be empty.")
	}
	if p.ProjectURL != nil && strings.TrimSpace(p.ProjectURL.String()) == "" {
		errs = append(errs, "projectURL cannot be empty.")
	}
	if p.RequireLicenseAcceptance {
		if p.LicenseURL == nil && p.LicenseMetadata == nil {
			errs = append(
				errs,
				"enabling license acceptance requires a license or a licenseUrl to be specified. The licenseUrl will be deprecated, consider using the license metadata.",
			)
		}
		if !p.EmitRequireLicenseAcceptance {
			errs = append(
				errs,
				"emitRequireLicenseAcceptance must not be set to false if RequireLicenseAcceptance is set to true.",
			)
		}
	}
	if p.LicenseURL != nil && p.LicenseMetadata != nil &&
		(strings.TrimSpace(p.LicenseURL.String()) == "" || !strings.EqualFold(p.LicenseURL.String(), p.LicenseMetadata.GetLicense())) {
		errs = append(errs, "the licenseUrl and license elements cannot be used together.")
	}
	return errs
}

func (p *PackageBuilder) isHasSymbolsInPackageType() bool {
	return nuget.Some(p.PackageTypes, func(packageType *PackageType) bool {
		return packageType.Equals(SymbolsPackage)
	})
}

func validatorPlatformVersion(frameworks []*Framework) error {
	platformVersions := make([]string, 0)
	for _, framework := range frameworks {
		if framework != nil && strings.TrimSpace(framework.Platform) != "" &&
			framework.PlatformVersion.Equal(nuget.EmptyVersion) {
			if name, err := framework.GetShortFolderName(); err != nil {
				return err
			} else {
				platformVersions = append(platformVersions, name)
			}
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

// findFileInPackage Looks for the specified file within the package
func findFileInPackage(filepath string, packageFiles []PackageFile, filePathIncorrectCase *string) PackageFile {
	filePathIncorrectCase = nil
	strippedFilePath := stripLeadingDirectorySeparators(filepath)
	for _, file := range packageFiles {
		// This must use a case-sensitive string comparison, even on systems where file paths are normally
		// case-sensitive.
		strippedPackageFilePath := stripLeadingDirectorySeparators(file.GetPath())
		if strings.EqualFold(strippedFilePath, strippedPackageFilePath) {
			// Found the requested file in the package
			filePathIncorrectCase = nil
			return file
			// Check for files that exist with the wrong file casing
		} else if filePathIncorrectCase == nil && strings.EqualFold(strippedPackageFilePath, strippedFilePath) {
			filePathIncorrectCase = &strippedPackageFilePath
		}
	}
	// We searched all of the package files and didn't find what we were looking for
	return nil
}
