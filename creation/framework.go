// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/huhouhua/go-nuget"
)

type KeyValuePair[K comparable, V any] struct {
	Key   K
	Value V
}

// FrameworkSpecificMapping A key value pair specific to a framework identifier
type FrameworkSpecificMapping struct {
	FrameworkIdentifier string

	Mapping *KeyValuePair[string, string]
}

type PortableFrameworkMappings interface {
	// GetProfileFrameworkMap  Ex: 5 -> net4, win8
	GetProfileFrameworkMap() []KeyValuePair[int, []*Framework]

	// GetProfileOptionalFrameworkMap  Additional optional frameworks supported in a portable profile.
	// Ex: 5 -> MonoAndroid1+MonoTouch1
	GetProfileOptionalFrameworkMap() []KeyValuePair[int, []*Framework]
}

type FrameworkMappings interface {
	// GetIdentifierSynonymsMap Synonym &#8210;&gt; Identifier
	// Ex: NET Framework &#8210;&gt; .NET Framework
	GetIdentifierSynonymsMap() []*KeyValuePair[string, string]

	// GetIdentifierShortNameMap Ex: .NET Framework &#8210;&gt; net
	GetIdentifierShortNameMap() []*KeyValuePair[string, string]

	// GetProfileShortNamesMap Ex: WindowsPhone &#8210;&gt; wp
	GetProfileShortNamesMap() []*FrameworkSpecificMapping

	// GetEquivalentFrameworkMap Equal frameworks. Used for legacy conversions.
	// ex: Framework: Win8 &lt;&#8210;&gt; Framework: NetCore45 Platform: Win8
	GetEquivalentFrameworkMap() []*KeyValuePair[*Framework, *Framework]
}

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

//func ParseNuGetFrameworkFromFilePath(filePath string, effectivePath *string) *Framework {
//	for _, knownFolder := range nuget.Known {
//		folderPrefix := fmt.Sprintf("%s%s", knownFolder, string(os.PathSeparator))
// 		if len(filePath) > len(folderPrefix) && strings.HasPrefix(strings.ToLower(filePath), strings.ToLower(folderPrefix))
// {
//			frameworkPart := filePath[len(folderPrefix):]
//
//		}
//	}
//}
//
//// ParseNuGetFrameworkFolderName Parses the specified string into FrameworkName object.
//func ParseNuGetFrameworkFolderName(frameworkPath string, strictParsing bool, effectivePath *string) *Framework {
//	dir := filepath.Dir(frameworkPath)
//	targetFrameworkString := strings.Split(dir, string(filepath.Separator))[0]
//	effectivePath = &frameworkPath
//	if strings.TrimSpace(targetFrameworkString) == "" {
//		return nil
//	}
//
//}
