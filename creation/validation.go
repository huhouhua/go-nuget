// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/huhouhua/go-nuget"
)

func (p *PackageBuilder) validate() []error {
	var errors []error
	errors = append(errors, p.validateDependencies())
	errors = append(errors, p.validateFilesUnique())
	errors = append(errors, p.ValidateReferenceAssemblies())
	errors = append(errors, p.validateFrameworkAssemblies())
	errors = append(errors, p.validateLicenseFile())
	errors = append(errors, p.validateIconFile())
	errors = append(errors, p.validateFileFrameworks())
	errors = append(errors, p.validateReadmeFile())
	return errors
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
