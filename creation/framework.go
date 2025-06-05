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

type Framework struct {
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

	// TODO IsUnsupported True if this framework was invalid or unknown. This framework is only compatible with Any and
	// Agnostic.
	IsUnsupported bool

	// TODO IsSpecificFramework True if this framework is real and not one of the special identifiers.
	IsSpecificFramework bool
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

// GetFrameworkString TODO
func (f *Framework) GetFrameworkString() string {
	return ""
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
	if strictParsing || nugetFramework.IsSpecificFramework {
		*effectivePath = frameworkPath[len(targetFrameworkString)+1:]
		return nugetFramework, err
	}
	return nil, fmt.Errorf("framework '%s' is not specific and strict parsing is disabled", targetFrameworkString)
}
