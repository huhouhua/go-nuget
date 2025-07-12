// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/huhouhua/go-nuget"
)

// Parse Creates a NuGetFramework from a folder name using the given provider.
func Parse(folderName string) (*Framework, error) {
	return ParseFromDefault(folderName, GetProviderInstance())
}

// ParseFromDefault Creates a NuGetFramework from a folder name using the given provider.
func ParseFromDefault(folderName string, provider FrameworkNameProvider) (*Framework, error) {
	if strings.Contains(folderName, ",") {
		return ParseFrameworkName(folderName, provider)
	}
	return ParseFolder(folderName, provider)
}

// ParseFrameworkName Creates a NuGetFramework from a .NET FrameworkName
func ParseFrameworkName(frameworkName string, provider FrameworkNameProvider) (*Framework, error) {
	frameworkNameParts := strings.Split(frameworkName, ",")
	var parts []string
	for _, part := range frameworkNameParts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	// if the first part is a special framework, ignore the rest
	framework := parseSpecialFramework(parts[0])
	if framework == nil {
		return nil, fmt.Errorf("failed to parse special framework from: %q", parts[0])
	}
	frameworkStr, profile, version, err := parseFrameworkNameParts(provider, parts)
	if err != nil {
		return nil, err
	}
	if version.Semver.Major() >= 5 && strings.EqualFold(nuget.NetCoreApp, frameworkStr) {
		return NewFrameworkWithPlatform(frameworkStr, version, "", nuget.EmptyVersion), nil
	}
	return NewFrameworkWithProfile(frameworkStr, version, profile), nil
}

func parseFrameworkNameParts(
	provider FrameworkNameProvider,
	parts []string,
) (framework, profile string, version *nuget.Version, err error) {
	framework = provider.GetIdentifier(parts[0])
	if framework == "" {
		framework = parts[0]
	}
	version = nuget.EmptyVersion
	var versionPart, profilePart string
	versionParts := nuget.Filter(parts, func(s string) bool {
		partLower := strings.ToLower(s)
		return strings.HasPrefix(partLower, "version=")
	})
	if len(versionParts) == 1 {
		versionPart = versionParts[0]
	}
	profileParts := nuget.Filter(parts, func(s string) bool {
		partLower := strings.ToLower(s)
		return strings.HasPrefix(partLower, "profile=")
	})
	if len(profileParts) == 1 {
		profilePart = profileParts[0]
	}
	if strings.TrimSpace(versionPart) != "" {
		versionParts := strings.Split(versionPart, "=")
		versionString := strings.TrimPrefix(versionParts[1], "v")
		if !strings.Contains(versionString, ".") {
			versionString += ".0"
		}
		if version, err = nuget.ParseVersion(versionString); err != nil {
			return "", "", nil, fmt.Errorf("invalid framework version '%s'", versionString)
		}
	}
	if strings.TrimSpace(profilePart) != "" {
		profile = strings.Split(profilePart, "=")[1]
	}
	if strings.EqualFold(nuget.Portable, framework) && strings.TrimSpace(profile) != "" &&
		strings.Contains(profile, "-") {
		return "", "", nil, fmt.Errorf(
			"invalid portable frameworks '%s'. A hyphen may not be in any of the portable framework names",
			profile,
		)
	}
	return framework, profile, version, nil
}

// ParseFolderFromDefault Creates a NuGetFramework from a folder name using the default mappings.
func ParseFolderFromDefault(folderName string) (*Framework, error) {
	return ParseFolder(folderName, GetProviderInstance())
}

// ParseFolder Creates a NuGetFramework from a folder name using the given provider.
func ParseFolder(folderName string, provider FrameworkNameProvider) (*Framework, error) {
	var (
		err    error
		result *Framework
	)
	if strings.Contains(folderName, "%s") {
		if folderName, err = url.QueryUnescape(folderName); err != nil {
			return nil, err
		}
	}
	// first check if we have a special or common framework
	if result = parseSpecialFramework(folderName); result != nil {
		return result, nil
	}
	if result = parseCommonFramework(folderName); result != nil {
		return result, nil
	}
	// assume this is unsupported unless we find a match
	result = NewFramework(Unsupported)
	identifier, version, profile := rawParse(folderName)
	if strings.TrimSpace(identifier) == "" && strings.TrimSpace(version) == "" && strings.TrimSpace(profile) == "" {
		// If the framework was not recognized check if it is a deprecated framework
		if deprecated := parseDeprecatedFramework(folderName); deprecated != nil {
			result = deprecated
		}
		return result, nil
	}
	framework := provider.GetIdentifier(identifier)
	if strings.TrimSpace(framework) == "" {
		return result, nil
	}
	nugetVersion, err := provider.GetVersion(version)
	if strings.TrimSpace(version) != "" || err != nil {
		return result, nil
	}
	profileShort := profile
	if nugetVersion.Semver.Major() >= 5 &&
		(strings.EqualFold(nuget.Net, framework) || strings.EqualFold(nuget.NetCoreApp, framework)) {
		// net should be treated as netcoreapp in 5.0 and later
		framework = nuget.NetCoreApp
		if strings.TrimSpace(profileShort) != "" {
			// Find a platform version if it exists and yank it out
			platformChars := profileShort
			versionStart := 0
			for versionStart < len(platformChars) && isLetterOrDot(rune(platformChars[versionStart])) {
				versionStart++
			}
			platform := profileShort
			platformVersionString := ""
			if versionStart > 0 {
				platform = profileShort[0:versionStart]
				platformVersionString = profileShort[versionStart:]
			}
			// Parse the version if it's there.
			var platformVersion *nuget.Version
			if v, err := provider.GetPlatformVersion(platformVersionString); err == nil {
				platformVersion = v
			} else {
				platformVersion = nuget.EmptyVersion
			}
			if strings.TrimSpace(platformVersionString) == "" || platformVersion != nil {
				result = NewFrameworkWithPlatform(framework, nugetVersion, platform, platformVersion)
			} else {
				// with result == UnsupportedFramework
				return result, nil
			}
		} else {
			result = NewFrameworkWithPlatform(framework, nugetVersion, "", nuget.EmptyVersion)
		}
	} else {
		pro := ""
		if pro = provider.GetProfile(profileShort); strings.TrimSpace(pro) == "" {
			pro = profileShort
		}
		if strings.EqualFold(nuget.Portable, framework) {
			if clientFrameworks, err := provider.GetPortableFrameworks(profileShort); err != nil {
				return result, nil
			} else {
				if profileNumber := provider.GetPortableProfile(clientFrameworks); profileNumber != -1 {
					portableProfileNumber := GetPortableProfileNumberString(profileNumber)
					result = NewFrameworkWithProfile(framework, nugetVersion, portableProfileNumber)
				} else {
					result = NewFrameworkWithProfile(framework, nugetVersion, profileShort)
				}
			}
		} else {
			result = NewFrameworkWithProfile(framework, nugetVersion, pro)
		}
	}
	return result, nil
}

// parseDeprecatedFramework Attempt to parse a common but deprecated framework using an exact string match
// Support for these should be dropped as soon as possible.
func parseDeprecatedFramework(s string) *Framework {
	switch s {
	case "45", "4.5":
		return Net45
	case "40", "4.0", "4":
		return Net4
	case "35", "3.5":
		return Net35
	case "20", "2", "2.0":
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

// parseCommonFramework A set of special and common frameworks that can be returned from the list of constants without
// parsing
// Using the interned frameworks here optimizes comparisons since they can be checked by reference.
// This is designed to optimize
func parseCommonFramework(frameworkString string) *Framework {
	frameworkString = strings.ToLower(frameworkString)

	switch frameworkString {
	case "dotnet", "dotnet50", "dotnet5.0":
		return DotNet50
	case "net40", "net4":
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
	case "netstandard1.0", "netstandard10":
		return NetStandard10
	case "netstandard1.1", "netstandard11":
		return NetStandard11
	case "netstandard1.2", "netstandard12":
		return NetStandard12
	case "netstandard1.3", "netstandard13":
		return NetStandard13
	case "netstandard1.4", "netstandard14":
		return NetStandard14
	case "netstandard1.5", "netstandard15":
		return NetStandard15
	case "netstandard1.6", "netstandard16":
		return NetStandard16
	case "netstandard1.7", "netstandard17":
		return NetStandard17
	case "netstandard2.0", "netstandard20":
		return NetStandard20
	case "netstandard2.1", "netstandard21":
		return NetStandard21
	case "netcoreapp1.0":
		return NetCoreApp10
	case "netcoreapp1.1":
		return NetCoreApp11
	case "netcoreapp2.0":
		return NetCoreApp20
	case "netcoreapp2.1", "netcoreapp21":
		return NetCoreApp21
	case "netcoreapp2.2":
		return NetCoreApp22
	case "netcoreapp3.0", "netcoreapp30":
		return NetCoreApp30
	case "netcoreapp3.1", "netcoreapp31":
		return NetCoreApp31
	case "netcoreapp5.0", "netcoreapp50", "net5.0", "net50":
		return Net50
	case "netcoreapp6.0", "netcoreapp60", "net6.0", "net60":
		return Net60
	case "netcoreapp7.0", "netcoreapp70", "net7.0", "net70":
		return Net70
	case "netcoreapp8.0", "netcoreapp80", "net8.0", "net80":
		return Net80
	case "net9.0":
		return Net90
	case "net10.0":
		return Net10_0
	}
	return nil
}
func GetPortableProfileNumberString(profileNumber int) string {
	return fmt.Sprintf("Profile%v", profileNumber)
}
