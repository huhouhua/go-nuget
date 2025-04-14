// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"fmt"
	"strings"
)

type PackageDependencyInfo struct {
	PackageIdentity          *PackageIdentity
	DependencyGroups         []PackageDependencyGroup
	FrameworkReferenceGroups []FrameworkSpecificGroup
}

func NewPackageDependencyInfo(dependencyGroups []PackageDependencyGroup, frameworkReferenceGroups []FrameworkSpecificGroup) (*PackageDependencyInfo, error) {
	if dependencyGroups == nil {
		return nil, fmt.Errorf("dependencyGroups cannot be nil")
	}
	if frameworkReferenceGroups == nil {
		return nil, fmt.Errorf("frameworkReferenceGroups cannot be nil")
	}
	return &PackageDependencyInfo{
		DependencyGroups:         dependencyGroups,
		FrameworkReferenceGroups: frameworkReferenceGroups,
	}, nil
}

// PackageDependencyGroup  Package dependencies grouped to a target framework.
type PackageDependencyGroup struct {
	// TargetFramework Dependency group target framework
	TargetFramework string `json:"targetFramework,omitempty"`

	// Packages Package dependencies
	Packages []PackageDependency `json:"dependencies"`
}

// NewPackageDependencyGroup new Dependency group
func NewPackageDependencyGroup(targetFramework string, packages []PackageDependency) (*PackageDependencyGroup, error) {
	return &PackageDependencyGroup{
		TargetFramework: targetFramework,
		Packages:        packages,
	}, nil
}

type PackageIdentity struct {
	Id      string        `json:"id"`
	Version *NuGetVersion `json:"version,omitempty"`
}

// HasVersion True if the version is non-null
func (p *PackageIdentity) HasVersion() bool {
	return p.Version != nil
}

// FrameworkSpecificGroup
type FrameworkSpecificGroup struct {
	Items           []string
	HasEmptyFolder  bool
	TargetFramework string
}

// NewFrameworkSpecificGroup New a Framework specific group
func NewFrameworkSpecificGroup(TargetFramework string, items []string) (*FrameworkSpecificGroup, error) {
	if items == nil {
		return nil, fmt.Errorf("items cannot be nil")
	}
	f := &FrameworkSpecificGroup{
		TargetFramework: TargetFramework,
		Items:           make([]string, len(items)),
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

// PackageDependency Represents a package dependency Id and allowed version range.
type PackageDependency struct {
	// Dependency package Id
	Id string
	// Include Types to include from the dependency package.
	Include []string
	// Exclude Types to exclude from the dependency package.
	Exclude []string

	// todo handler range
	Version string
}

// PackageDependencyInfoFunc can be used to customize a new PackageDependencyInfo.
type PackageDependencyInfoFunc func(*PackageDependencyInfo) error

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
		nugetVersion, err := Parse(meta.Version)
		if err != nil {
			return err
		}
		info.PackageIdentity.Id = meta.ID
		info.PackageIdentity.Version = nugetVersion
		return nil
	}
}

// WithDependencyGroups can be used to set the dependency groups for the PackageDependencyInfo.
func WithDependencyGroups(meta *Metadata) PackageDependencyInfoFunc {
	return func(info *PackageDependencyInfo) error {
		return nil
	}
}
