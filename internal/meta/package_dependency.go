// Copyright (c) 2025 Kevin Berger <huhouhuam@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package meta

import (
	"fmt"
	"strings"

	nugetVersion "github.com/huhouhua/go-nuget/version"
)

type PackageDependencyInfo struct {
	PackageIdentity          *PackageIdentity
	DependencyGroups         []*PackageDependencyGroup
	FrameworkReferenceGroups []*FrameworkSpecificGroup
}

// PackageDependencyGroup  Package dependencies grouped to a target framework.
type PackageDependencyGroup struct {
	// TargetFramework Dependency group target framework
	TargetFramework string `json:"targetFramework,omitempty"`

	// Packages Package dependencies
	Packages []*Dependency `json:"dependencies"`
}

// NewPackageDependencyGroup new Dependency group
func NewPackageDependencyGroup(targetFramework string, packages ...*Dependency) (*PackageDependencyGroup, error) {
	for _, pkg := range packages {
		if err := pkg.Parse(); err != nil {
			return nil, err
		}
	}
	return &PackageDependencyGroup{
		TargetFramework: targetFramework,
		Packages:        packages,
	}, nil
}

type PackageIdentity struct {
	Id      string                `json:"id"`
	Version *nugetVersion.Version `json:"version,omitempty"`
}

func NewPackageIdentity(id, version string) (*PackageIdentity, error) {
	nv, err := nugetVersion.Parse(version)
	if err != nil {
		return nil, err
	}
	return &PackageIdentity{
		Id:      id,
		Version: nv,
	}, err
}

// HasVersion True if the version is non-null
func (p *PackageIdentity) HasVersion() bool {
	return p.Version != nil
}

// FrameworkSpecificGroup Framework specific group
type FrameworkSpecificGroup struct {
	Items           []string
	HasEmptyFolder  bool
	TargetFramework string
}

// NewFrameworkSpecificGroup New a Framework specific group
func NewFrameworkSpecificGroup(TargetFramework string, items ...string) (*FrameworkSpecificGroup, error) {
	if items == nil {
		return nil, fmt.Errorf("items cannot be nil")
	}
	f := &FrameworkSpecificGroup{
		TargetFramework: TargetFramework,
		Items:           make([]string, 0, len(items)),
	}
	for _, item := range items {
		if strings.HasSuffix(item, "/_._") {
			f.HasEmptyFolder = true
			continue
		}
		f.Items = append(f.Items, item)
	}
	return f, nil
}

// PackageDependencyInfoFunc can be used to customize a new PackageDependencyInfo.
type PackageDependencyInfoFunc func(*PackageDependencyInfo) error

// ConfigureDependencyInfo configures a PackageDependencyInfo from a Nuspec.
func ConfigureDependencyInfo(info *PackageDependencyInfo, nuspec Nuspec) error {
	return ApplyPackageDependency(info,
		WithIdentity(nuspec.Metadata),
		WithDependencyGroups(nuspec.Metadata.Dependencies),
		WithFrameworkReferenceGroups(nuspec.Metadata.FrameworkAssemblies))
}

// ApplyPackageDependency applies a list of PackageDependencyInfoFunc to a PackageDependencyInfo.
func ApplyPackageDependency(info *PackageDependencyInfo, options ...PackageDependencyInfoFunc) error {
	for _, opt := range options {
		if err := opt(info); err != nil {
			return err
		}
	}
	return nil
}

// WithIdentity can be used to set a package identity for the PackageDependencyInfo.
func WithIdentity(meta *Metadata) PackageDependencyInfoFunc {
	return func(info *PackageDependencyInfo) error {
		if identity, err := NewPackageIdentity(meta.ID, meta.Version); err != nil {
			return err
		} else {
			info.PackageIdentity = identity
		}
		return nil
	}
}

// WithDependencyGroups can be used to set the dependency groups for the PackageDependencyInfo.
func WithDependencyGroups(dependencies *Dependencies) PackageDependencyInfoFunc {
	return func(info *PackageDependencyInfo) error {
		if dependencies == nil {
			return nil
		}
		groupFound := false
		if dependencies.Groups != nil {
			for _, groups := range dependencies.Groups {
				groupFound = true
				if group, err := NewPackageDependencyGroup(groups.TargetFramework, groups.Dependencies...); err != nil {
					return err
				} else {
					info.DependencyGroups = append(info.DependencyGroups, group)
				}
			}
		}
		if !groupFound {
			for _, dependency := range dependencies.Dependency {
				if group, err := NewPackageDependencyGroup("Any", dependency); err != nil {
					return err
				} else {
					info.DependencyGroups = append(info.DependencyGroups, group)
				}
			}
		}
		return nil
	}
}

// WithFrameworkReferenceGroups can be used to set the framework reference groups for the PackageDependencyInfo.
func WithFrameworkReferenceGroups(framework *FrameworkAssemblies) PackageDependencyInfoFunc {
	return func(info *PackageDependencyInfo) error {
		if framework == nil || framework.FrameworkAssembly == nil {
			return nil
		}
		for _, assembly := range framework.FrameworkAssembly {
			if group, err := NewFrameworkSpecificGroup(assembly.TargetFramework, assembly.AssemblyName...); err != nil {
				return err
			} else {
				info.FrameworkReferenceGroups = append(info.FrameworkReferenceGroups, group)
			}
		}
		return nil
	}
}
