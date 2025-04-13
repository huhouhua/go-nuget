package nuget

import "fmt"

type PackageDependencyInfo struct {
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
	TargetFramework *NuGetFramework `json:"targetFramework,omitempty"`

	// Packages Package dependencies
	Packages []PackageDependency `json:"dependencies"`
}

// NewPackageDependencyGroup new Dependency group
func NewPackageDependencyGroup(targetFramework *NuGetFramework, packages []PackageDependency) (*PackageDependencyGroup, error) {
	if targetFramework == nil {
		return nil, fmt.Errorf("targetFramework cannot be nil")
	}
	if packages == nil {
		return nil, fmt.Errorf("packages cannot be nil")
	}
	return &PackageDependencyGroup{
		TargetFramework: targetFramework,
		Packages:        packages,
	}, nil
}

// FrameworkSpecificGroup
type FrameworkSpecificGroup struct {
}

// PackageDependency Represents a package dependency Id and allowed version range.
type PackageDependency struct {
	// Dependency package Id
	Id string
	// Include Types to include from the dependency package.
	Include []string
	// Exclude Types to exclude from the dependency package.
	Exclude []string
}

// NuGetFramework
type NuGetFramework struct {
	// Framework Target framework
	Framework string
	// Version Target framework version
	Version Version

	// Platform Framework Platform (net5.0+)
	Platform string

	// PlatformVersion Framework Platform Version (net5.0+)
	PlatformVersion Version
}

func NewNuGetFramework() {

}
