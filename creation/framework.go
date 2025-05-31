// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
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

// parseDeprecatedFramework Attempt to parse a common but deprecated framework using an exact string match
// Support for these should be dropped as soon as possible.
func parseDeprecatedFramework(s string) *Framework {
	switch s {
	case "45":
	case "4.5":
		return Net45
	case "40":
	case "4.0":
	case "4":
		return Net4
	case "35":
	case "3.5":
		return Net35
	case "20":
	case "2":
	case "2.0":
		return Net2
	}
	return nil
}

// rawParse parses a framework string like "net45-client" into identifier, version, and profile.
func rawParse(s string) (identifier string, version string, profile string) {
	profile = ""
	var versionStr string
	chars := []rune(s)

	versionStart := 0
	for versionStart < len(chars) && isLetterOrDot(chars[versionStart]) {
		versionStart++
	}

	if versionStart > 0 {
		identifier = s[:versionStart]
	} else {
		// invalid, we no longer support names like: 40
		return "", "", ""
	}

	profileStart := versionStart
	for profileStart < len(chars) && isDigitOrDot(chars[profileStart]) {
		profileStart++
	}

	versionLength := profileStart - versionStart
	if versionLength > 0 {
		versionStr = s[versionStart:profileStart]
	}

	if profileStart < len(chars) {
		if chars[profileStart] == '-' {
			actualProfileStart := profileStart + 1
			if actualProfileStart == len(chars) {
				// empty profiles are not allowed
				return "", "", ""
			}

			profile = s[actualProfileStart:]
			for _, c := range profile {
				// validate the profile string to AZaz09-+.
				if !isValidProfileChar(c) {
					return "", "", ""
				}
			}
		} else {
			// invalid profile
			return "", "", ""
		}
	}

	return identifier, versionStr, profile
}

func isLetterOrDot(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '.'
}

func isDigitOrDot(r rune) bool {
	return (r >= '0' && r <= '9') || r == '.'
}

// isValidProfileChar reports whether r is a valid character for a profile segment
// Acceptable characters: letter (a-zA-Z), digit (0-9), '.', '-', '+'
func isValidProfileChar(r rune) bool {
	return (r >= '0' && r <= '9') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= 'a' && r <= 'z') ||
		r == '.' ||
		r == '+' ||
		r == '-'
}

func parseSpecialFramework(frameworkString string) *Framework {
	// wildcard matching all frameworks
	if strings.EqualFold(frameworkString, Any) {
		return NewFramework(Any)
	}
	// framework with no specific target framework. This can be used for content only packages.
	if strings.EqualFold(frameworkString, Agnostic) {
		return NewFramework(Agnostic)
	}
	// unknown or invalid framework
	if strings.EqualFold(frameworkString, Unsupported) {
		return NewFramework(Unsupported)
	}
	return nil
}

// ParseCommonFramework A set of special and common frameworks that can be returned from the list of constants without
// parsing
// Using the interned frameworks here optimizes comparisons since they can be checked by reference.
// This is designed to optimize
func ParseCommonFramework(frameworkString string) *Framework {
	frameworkString = strings.ToLower(frameworkString)

	switch frameworkString {
	case "dotnet":
	case "dotnet50":
	case "dotnet5.0":
		return DotNet50
	case "net40":
	case "net4":
		return Net4
	case "net403":
		return Net403
	case "net45":
		return Net45
	case "net451":
		return Net451
	case "net452":
		return Net452
	case "net46":
		return Net46
	case "net461":
		return Net461
	case "net462":
		return Net462
	case "net463":
		return Net463
	case "net47":
		return Net47
	case "net471":
		return Net471
	case "net472":
		return Net472
	case "net48":
		return Net48
	case "net481":
		return Net481
	case "win8":
		return Win8
	case "win81":
		return Win81
	case "netstandard":
		return NetStandard
	case "netstandard1.0":
	case "netstandard10":
		return NetStandard10
	case "netstandard1.1":
	case "netstandard11":
		return NetStandard11
	case "netstandard1.2":
	case "netstandard12":
		return NetStandard12
	case "netstandard1.3":
	case "netstandard13":
		return NetStandard13
	case "netstandard1.4":
	case "netstandard14":
		return NetStandard14
	case "netstandard1.5":
	case "netstandard15":
		return NetStandard15
	case "netstandard1.6":
	case "netstandard16":
		return NetStandard16
	case "netstandard1.7":
	case "netstandard17":
		return NetStandard17
	case "netstandard2.0":
	case "netstandard20":
		return NetStandard20
	case "netstandard2.1":
	case "netstandard21":
		return NetStandard21
	case "netcoreapp1.0":
		return NetCoreApp10
	case "netcoreapp1.1":
		return NetCoreApp11
	case "netcoreapp2.0":
		return NetCoreApp20
	case "netcoreapp2.1":
	case "netcoreapp21":
		return NetCoreApp21
	case "netcoreapp2.2":
		return NetCoreApp22
	case "netcoreapp3.0":
	case "netcoreapp30":
		return NetCoreApp30
	case "netcoreapp3.1":
	case "netcoreapp31":
		return NetCoreApp31
	case "netcoreapp5.0":
	case "netcoreapp50":
	case "net5.0":
	case "net50":
		return Net50
	case "netcoreapp6.0":
	case "netcoreapp60":
	case "net6.0":
	case "net60":
		return Net60
	case "netcoreapp7.0":
	case "netcoreapp70":
	case "net7.0":
	case "net70":
		return Net70
	case "netcoreapp8.0":
	case "netcoreapp80":
	case "net8.0":
	case "net80":
		return Net80
	case "net9.0":
		return Net90
	case "net10.0":
		return Net10_0
	}
	return nil
}
