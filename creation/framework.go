// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/huhouhua/go-nuget/version"

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

	// isNet5Era True if this framework is Net5 or later, until we invent something new.
	isNet5Era bool

	// Framework Target framework
	Framework string

	// Version Target framework version
	Version *version.Version

	// Platform Framework Platform (net5.0+)
	Platform string

	// PlatformVersion Framework Platform Version (net5.0+)
	PlatformVersion *version.Version

	// Target framework profile
	Profile string
}

func NewFramework(framework string) *Framework {
	return NewFrameworkWithVersion(framework, nuget.EmptyVersion)
}
func NewFrameworkWithVersion(framework string, version *version.Version) *Framework {
	return NewFrameworkWithProfile(framework, version, "")
}
func NewFrameworkWithProfile(framework string, version *version.Version, profile string) *Framework {
	return newFrameworkFrom(framework, version, profile, "", nuget.EmptyVersion)
}

func NewFrameworkWithPlatform(
	framework string,
	version *version.Version,
	platform string,
	platformVersion *version.Version,
) *Framework {
	return newFrameworkFrom(framework, version, "", platform, platformVersion)
}

func newFrameworkFrom(
	framework string,
	version *version.Version,
	profile string,
	platform string,
	platformVersion *version.Version,
) *Framework {
	nf := &Framework{
		Framework: framework,
		Version:   version,
		Profile:   profile,
	}
	nf.isNet5Era = nf.Version.Semver.Major() >= 5 &&
		strings.EqualFold(strings.ToLower(framework), strings.ToLower(nuget.NetCoreApp))

	if nf.isNet5Era {
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
	return unsupportedFramework.Equals(f)
}

// IsAgnostic  True if this framework is non-specific. Always compatible.
func (f *Framework) IsAgnostic() bool {
	return agnosticFramework.Equals(f)
}

// IsAny True if this is the any framework. Always compatible.
func (f *Framework) IsAny() bool {
	return anyFramework.Equals(f)
}

// IsSpecificFramework True if this framework is real and not one of the special identifiers.
func (f *Framework) IsSpecificFramework() bool {
	return !f.IsAgnostic() && !f.IsAny() && !f.IsUnsupported()
}

// AllFrameworkVersions True if this framework matches for all versions.
func (f *Framework) AllFrameworkVersions() bool {
	return f.Version.Semver.Major() == 0 && f.Version.Semver.Minor() == 0 && f.Version.Semver.Patch() == 0 &&
		f.Version.Semver.Metadata() == ""
}

// IsPCL Portable class library check
func (f *Framework) IsPCL() bool {
	return strings.EqualFold(f.Framework, nuget.Portable) && f.Version.Semver.Major() < 5
}

// GetFrameworkString which is relevant for building packages. This isn't needed for net5.0+ frameworks.
func (f *Framework) GetFrameworkString() (string, error) {
	isNet5Era := f.Version.Semver.Major() >= 5 && strings.EqualFold(nuget.NetCoreApp, f.Framework)
	if isNet5Era {
		return f.GetShortFolderName()
	}
	frameworkName, err := NewFrameworkName(f.GetDotNetFrameworkName())
	if err != nil {
		return "", err
	}
	version := frameworkName.GetVersion()
	original := version.OriginalVersion
	name := fmt.Sprintf("%s%s", frameworkName.GetIdentifier(), original)
	if strings.TrimSpace(frameworkName.GetProfile()) == "" {
		return name, nil
	}
	return fmt.Sprintf("%s-%s", name, frameworkName.GetProfile()), nil
}

// getFrameworkIdentifier Helper that is .NET 5 Era aware to replace identifier when appropriate
func (f *Framework) getFrameworkIdentifier() string {
	if f.isNet5Era {
		return nuget.Net
	}
	return f.Framework
}

// GetShortFolderName Creates the shortened version of the framework using the default mappings.
func (f *Framework) GetShortFolderName() (string, error) {
	return f.getShortFolderName(GetProviderInstance())
}

// GetShortFolderName Creates the shortened version of the framework using the given mappings.
func (f *Framework) getShortFolderName(mappings FrameworkNameProvider) (string, error) {
	// Check for rewrites
	framework := mappings.GetShortNameReplacement(f)
	var sb strings.Builder
	if !f.IsSpecificFramework() {
		// unsupported, any, agnostic
		return strings.ToLower(f.Framework), nil
	}
	// get the framework
	shortFramework := mappings.GetShortIdentifier(f.getFrameworkIdentifier())
	if strings.TrimSpace(shortFramework) == "" {
		shortFramework = getLettersAndDigitsOnly(framework.Framework)
	}
	if strings.TrimSpace(shortFramework) == "" {
		return "", fmt.Errorf("invalid framework identifier '%s'", shortFramework)
	}
	// add framework
	sb.WriteString(shortFramework)

	// add the version if it is non-empty
	if !f.AllFrameworkVersions() {
		sb.WriteString(mappings.GetVersionString(framework.Framework, framework.Version.Semver))
	}
	if f.IsPCL() {
		sb.WriteString("-")
		frameworkErr := fmt.Errorf(
			"invalid portable frameworks for '%s'. A portable framework must have at least one framework in the profile",
			framework.GetDotNetFrameworkName(),
		)
		if strings.TrimSpace(framework.Profile) != "" {
			if frameworks, err := mappings.GetPortableFrameworksWithInclude(framework.Profile, false); err != nil {
				return "", err
			} else {
				if len(frameworks) > 0 {
					if frameworks, err = mappings.GetPortableFrameworksWithInclude(framework.Profile, false); err != nil {
						return "", err
					}
					sortedFrameworks := make([]string, 0)
					for _, fw := range frameworks {
						if name, err := fw.getShortFolderName(mappings); err != nil {
							return "", err
						} else {
							sortedFrameworks = append(sortedFrameworks, name)
						}
					}
					// sort the PCL frameworks by alphabetical order
					sort.Slice(sortedFrameworks, func(i, j int) bool {
						return strings.ToLower(sortedFrameworks[i]) < strings.ToLower(sortedFrameworks[j])
					})
					sb.WriteString(strings.Join(sortedFrameworks, "+"))
				} else {
					return "", frameworkErr
				}
			}
		} else {
			return "", frameworkErr
		}
	} else if f.isNet5Era {
		if strings.TrimSpace(framework.Platform) != "" {
			sb.WriteString("-")
			sb.WriteString(strings.ToLower(framework.Platform))
			if framework.PlatformVersion.Semver.Equal(nuget.EmptyVersion.Semver) {
				sb.WriteString(mappings.GetVersionString(framework.Framework, framework.PlatformVersion.Semver))
			}
		}
	} else {
		// add the profile
		if shortProfile := mappings.GetShortProfile(framework.Profile); strings.TrimSpace(shortProfile) == "" {
			// if we have a profile, but can't get a mapping, just use the profile as is
			shortProfile = framework.Profile
		} else {
			sb.WriteString("-")
			sb.WriteString(shortProfile)
		}
	}
	return strings.ToLower(sb.String()), nil
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
		parts = append(parts, fmt.Sprintf("Version=v%s", getDisplayVersion(framework.Version.Semver)))
		if strings.TrimSpace(framework.Profile) != "" {
			parts = append(parts, fmt.Sprintf("Profile=%s", framework.Profile))
		}
		return strings.Join(parts, ",")
	} else {
		return fmt.Sprintf("%s,Version=v0.0", framework.Framework)
	}

}

func (f *Framework) String() (string, error) {
	if f.isNet5Era {
		return f.GetShortFolderName()
	}
	return f.GetDotNetFrameworkName(), nil
}

func (f *Framework) Equals(other *Framework) bool {
	if other == nil {
		return f == other
	}
	return f.Version.Semver.Equal(other.Version.Semver) &&
		strings.EqualFold(f.Framework, other.Framework) &&
		strings.EqualFold(f.Profile, other.Profile) &&
		strings.EqualFold(f.Platform, other.Platform) &&
		f.PlatformVersion.Semver.Equal(other.PlatformVersion.Semver) &&
		!f.IsUnsupported()
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
func getLettersAndDigitsOnly(s string) string {
	var sb strings.Builder
	for _, c := range s {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			sb.WriteRune(c)
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
