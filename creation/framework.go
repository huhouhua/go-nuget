// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/huhouhua/go-nuget"
)

var (
	// An unknown or invalid framework
	unsupportedFramework = NewFramework(Unsupported)

	//  A framework with no specific target framework. This can be used for content only packages.
	agnosticFramework = NewFramework(Agnostic)

	// A wildcard matching all frameworks
	anyFramework = NewFramework(Any)
)

type Framework struct {
	targetFrameworkMoniker string

	// Framework Target framework
	Framework string

	// Version Target framework version
	Version *semver.Version

	// Platform Framework Platform (net5.0+)
	Platform string

	// PlatformVersion Framework Platform Version (net5.0+)
	PlatformVersion *semver.Version

	// Target framework profile
	Profile string

	// TODO ShortFolderName the shortened version of the framework using the default mappings.
	ShortFolderName string
}

func NewFramework(framework string) *Framework {
	return NewFrameworkWithVersion(framework, nuget.EmptyVersion)
}
func NewFrameworkWithVersion(framework string, version *semver.Version) *Framework {
	return NewFrameworkWithProfile(framework, version, "")
}
func NewFrameworkWithProfile(framework string, version *semver.Version, profile string) *Framework {
	return newFrameworkFrom(framework, version, profile, "", nuget.EmptyVersion)
}

func NewFrameworkWithPlatform(
	framework string,
	version *semver.Version,
	platform string,
	platformVersion *semver.Version,
) *Framework {
	return newFrameworkFrom(framework, version, "", platform, platformVersion)
}

func newFrameworkFrom(
	framework string,
	version *semver.Version,
	profile string,
	platform string,
	platformVersion *semver.Version,
) *Framework {
	nf := &Framework{
		Framework: framework,
		Version:   version,
		Profile:   profile,
	}
	isNet5Era := nf.Version.Major() >= 5 &&
		strings.EqualFold(strings.ToLower(framework), strings.ToLower(nuget.NetCoreApp))

	if isNet5Era {
		nf.Platform = platform
		nf.PlatformVersion = platformVersion
	} else {
		nf.Platform = ""
		nf.PlatformVersion = nuget.EmptyVersion
	}
	return nf
}

// IsUnsupported True if this framework was invalid or unknown. This framework is only compatible with Any and
func (f *Framework) IsUnsupported() bool {
	return unsupportedFramework == f
}

// IsAgnostic  True if this framework is non-specific. Always compatible.
func (f *Framework) IsAgnostic() bool {
	return agnosticFramework == f
}

// IsAny True if this is the any framework. Always compatible.
func (f *Framework) IsAny() bool {
	return anyFramework == f
}

// IsSpecificFramework True if this framework is real and not one of the special identifiers.
func (f *Framework) IsSpecificFramework() bool {
	return !f.IsAgnostic() && !f.IsAny() && !f.IsUnsupported()
}

// GetFrameworkString which is relevant for building packages. This isn't needed for net5.0+ frameworks.
func (f *Framework) GetFrameworkString() (string, error) {
	isNet5Era := f.Version.Major() >= 5 && strings.EqualFold(nuget.NetCoreApp, f.Framework)
	if isNet5Era {
		return f.GetShortFolderName(), nil
	}
	frameworkName, err := NewFrameworkName(f.GetDotNetFrameworkName())
	if err != nil {
		return "", err
	}
	name := fmt.Sprintf("%s%s", frameworkName.GetIdentifier(), frameworkName.GetVersion().String())
	if strings.TrimSpace(frameworkName.GetProfile()) == "" {
		return name, nil
	}
	return fmt.Sprintf("%s-%s", name, frameworkName.GetProfile()), nil
}

// GetShortFolderName Creates the shortened version of the framework using the default mappings.
// Ex: net45
func (f *Framework) GetShortFolderName() string {
	return ""
}

// GetDotNetFrameworkName The TargetFrameworkMoniker identifier of the current NuGetFramework.
func (f *Framework) GetDotNetFrameworkName() string {
	if f.targetFrameworkMoniker == "" {
		f.targetFrameworkMoniker = f.getDotNetFrameworkName(GetProviderInstance())
	}
	return f.targetFrameworkMoniker
}

// getDotNetFrameworkName The TargetFrameworkMoniker identifier of the current NuGetFramework.
func (f *Framework) getDotNetFrameworkName(mappings FrameworkNameProvider) string {
	// Check for rewrites
	framework := mappings.GetFullNameReplacement(f)
	if framework.IsSpecificFramework() {
		parts := []string{f.Framework}
		parts = append(parts, fmt.Sprintf("Version=v%s", getDisplayVersion(framework.Version)))
		if strings.TrimSpace(framework.Profile) != "" {
			parts = append(parts, fmt.Sprintf("Profile=%s", framework.Profile))
		}
		return strings.Join(parts, ",")
	} else {
		return fmt.Sprintf("%s,Version=v0.0", framework.Framework)
	}

}

func getDisplayVersion(v *semver.Version) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%d.%d", v.Major(), v.Minor()))

	if v.Patch() > 0 || v.Metadata() != "" {
		sb.WriteString(fmt.Sprintf(".%d", v.Patch()))

		if v.Metadata() != "" {
			sb.WriteString("." + v.Metadata())
		}
	}

	return sb.String()
}

type FrameworkAssemblyReference struct {
	AssemblyName        string
	SupportedFrameworks []*Framework
}

func ParseNuGetFrameworkFromFilePath(filePath string, effectivePath *string) *Framework {
	for _, knownFolder := range nuget.Known {
		folderPrefix := fmt.Sprintf("%s%s", knownFolder, string(os.PathSeparator))
		if len(filePath) > len(folderPrefix) &&
			strings.HasPrefix(strings.ToLower(filePath), strings.ToLower(folderPrefix)) {
			frameworkPart := filePath[len(folderPrefix):]
			name, err := ParseNuGetFrameworkFolderName(frameworkPart, knownFolder == nuget.Lib, effectivePath)
			if err != nil {
				// if the parsing fails, we treat it as if this file
				// doesn't have target framework.
				*effectivePath = frameworkPart
				return nil
			}
			return name
		}
	}
	*effectivePath = filePath
	return nil
}

// ParseNuGetFrameworkFolderName Parses the specified string into FrameworkName object.
func ParseNuGetFrameworkFolderName(
	frameworkPath string,
	strictParsing bool,
	effectivePath *string,
) (*Framework, error) {
	dir := filepath.Dir(frameworkPath)
	targetFrameworkString := strings.Split(dir, string(filepath.Separator))[0]
	if effectivePath != nil {
		*effectivePath = frameworkPath
	}
	if strings.TrimSpace(targetFrameworkString) == "" {
		return nil, fmt.Errorf("invalid targetFrameworkString: cannot be empty or blank")
	}
	nugetFramework, err := ParseFolderFromDefault(targetFrameworkString)
	if err != nil {
		return nil, err
	}
	if strictParsing || nugetFramework.IsSpecificFramework() {
		*effectivePath = frameworkPath[len(targetFrameworkString)+1:]
		return nugetFramework, err
	}
	return nil, fmt.Errorf("framework '%s' is not specific and strict parsing is disabled", targetFrameworkString)
}
