// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"errors"
	"fmt"
	"maps"
	"os"
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
	errs = append(errs, p.validateFileFrameworks())
	errs = append(errs, p.validateReadmeFile())
	errs = append(errs, p.validateDependencyGroups())
	errs = append(errs, p.validateManifest())
	return errs
}
func (p *PackageBuilder) validateManifest() error {
	var results []string
	results = append(results, p.validateArgs()...)
	for _, reference := range p.PackageAssemblyReferences {
		results = append(results, reference.Validate()...)
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

	return nil
}

func (p *PackageBuilder) validateIconFile() error {
	return nil
}

func (p *PackageBuilder) validateFileFrameworks() error {
	return nil
}
func (p *PackageBuilder) validateReadmeFile() error {
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
					dep.VersionRange.MaxVersion.Equal(dep.VersionRange.MinVersion.Version) {
					return fmt.Errorf(fmt.Sprintf("dependency '%s' has an invalid version.", dep.Id))
				}
				if dep.VersionRange.MinVersion.GreaterThan(dep.VersionRange.MaxVersion.Version) {
					return fmt.Errorf("dependency '%s' has an invalid version", dep.Id)
				}
			}
		}
	}
	return nil
}
func (p *PackageBuilder) validateArgs() []string {
	var errors []string
	if strings.TrimSpace(p.Id) == "" {
		errors = append(errors, "id is required.")
	} else {
		if len(p.Id) > MaxPackageIdLength {
			errors = append(errors, "id must not exceed 100 characters.")
		} else if !IsValidPackageId(p.Id) {
			errors = append(errors, fmt.Sprintf("the package ID '%s' contains invalid characters. Examples of valid package IDs include 'MyPackage' and 'MyPackage.Sample'.", p.Id))
		}
	}
	if p.Version == nil {
		errors = append(errors, "version is required.")
	}
	if p.Authors == nil {
		errors = append(errors, "authors is required.")
	}
	isHasEmpty := nuget.Some(p.Authors, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
	isSymbols := nuget.Some(p.PackageTypes, func(packageType PackageType) bool {
		return packageType.Equals(SymbolsPackage)
	})
	if isHasEmpty && !isSymbols {
		errors = append(errors, "authors is required.")
	}
	if strings.TrimSpace(p.Description) == "" {
		errors = append(errors, "description is required.")
	}
	if p.LicenseURL != nil && strings.TrimSpace(p.LicenseURL.String()) == "" {
		errors = append(errors, "licenseURL cannot be empty.")
	}
	if p.IconURL != nil && strings.TrimSpace(p.IconURL.String()) == "" {
		errors = append(errors, "iconURL cannot be empty.")
	}
	if p.ProjectURL != nil && strings.TrimSpace(p.ProjectURL.String()) == "" {
		errors = append(errors, "projectURL cannot be empty.")
	}
	if strings.TrimSpace(p.Icon) == "" {
		errors = append(errors, "the element 'icon' cannot be empty.")
	}
	if strings.TrimSpace(p.Readme) == "" {
		errors = append(errors, "the element 'readme' cannot be empty.")
	}
	if p.RequireLicenseAcceptance {
		if p.LicenseURL == nil && p.LicenseMetadata == nil {
			errors = append(
				errors,
				"enabling license acceptance requires a license or a licenseUrl to be specified. The licenseUrl will be deprecated, consider using the license metadata.",
			)
		}
		if !p.EmitRequireLicenseAcceptance {
			errors = append(
				errors,
				"emitRequireLicenseAcceptance must not be set to false if RequireLicenseAcceptance is set to true.",
			)
		}
	}
	if p.LicenseURL != nil && p.LicenseMetadata != nil &&
		(strings.TrimSpace(p.LicenseURL.String()) == "" || !strings.EqualFold(p.LicenseURL.String(), p.LicenseMetadata.GetLicense())) {
		errors = append(errors, "the licenseUrl and license elements cannot be used together.")
	}
	return errors
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
