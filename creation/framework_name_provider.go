// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type FrameworkNameProvider struct {

	// identifierSynonyms Contains identifier -> identifier
	// Ex: .NET Framework -> .NET Framework
	// Ex: NET Framework -> .NET Framework
	// This includes self mappings.
	identifierSynonyms    map[string]string
	identifierToShortName map[string]string

	// profile -> supported frameworks, optional frameworks
	portableFrameworkMap         map[int]map[string]*Framework
	portableOptionalFrameworkMap map[int]map[string]*Framework

	profileShortToLongMap  map[string]string
	profilesToShortNameMap map[string]string

	// equivalent frameworks
	equivalentFrameworkMap map[string]map[string]*Framework
}

func (f *FrameworkNameProvider) GetIdentifier(framework string) string {
	return f.convertOrNormalize(framework, f.identifierSynonyms, f.identifierToShortName)
}
func (f *FrameworkNameProvider) GetProfile(profileShortName string) string {
	return f.convertOrNormalize(profileShortName, f.profileShortToLongMap, f.profilesToShortNameMap)
}
func (f *FrameworkNameProvider) GetVersion(versionString string) (*semver.Version, error) {
	versionString = strings.TrimSpace(versionString)
	if versionString == "" {
		return nil, fmt.Errorf("version is empty")
	}
	if strings.Contains(versionString, ".") {
		// parse the version as a normal dot delimited version
		return semver.NewVersion(versionString)
	}

	// make sure we have at least 2 digits
	if len(versionString) < 2 {
		versionString += "0"
	}
	// take only the first 4 digits and add dots
	// 451 -> 4.5.1
	// 81233 -> 8123
	if len(versionString) > 4 {
		versionString = versionString[:4]
	}
	parts := make([]byte, 0, len(versionString)*2-1)
	for i, ch := range versionString {
		if i > 0 {
			parts = append(parts, '.')
		}
		parts = append(parts, byte(ch))
	}
	return semver.NewVersion(string(parts))
}

func (f *FrameworkNameProvider) GetPlatformVersion(versionString string) (*semver.Version, error) {
	versionString = strings.TrimSpace(versionString)
	if versionString == "" {
		return nil, fmt.Errorf("version is empty")
	}
	if !strings.Contains(versionString, ".") {
		versionString += ".0"
	}
	return semver.NewVersion(versionString)
}

func (f *FrameworkNameProvider) GetPortableFrameworks(shortPortableProfiles string) ([]*Framework, error) {
	var shortNames []string
	for _, part := range strings.Split(shortPortableProfiles, "+") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			shortNames = append(shortNames, trimmed)
		}
	}
	var result []*Framework
	for _, name := range shortNames {
		if framework, err := Parse(name, *f); err != nil {
			return nil, err
		} else {
			if strings.TrimSpace(framework.Profile) != "" {
				return nil, fmt.Errorf("invalid portable frameworks '%s'. A hyphen may not be in any of the portable framework names", shortPortableProfiles)
			}
			result = append(result, framework)
		}
	}
	return result, nil
}

func (f *FrameworkNameProvider) GetPortableProfile(supportedFrameworks []*Framework) int {
	if supportedFrameworks == nil {
		return -1
	}
	// Remove duplicate frameworks, ex: win+win8 -> win
	profileFrameworkMap := f.removeDuplicateFramework(supportedFrameworks)
	reduced := make(map[string]*Framework)
	for k, vMap := range f.portableFrameworkMap {
		// to match the required set must be less than or the same count as the input
		// if we knew which frameworks were optional in the input we could rule out the lesser ones also
		if len(vMap) <= len(profileFrameworkMap) {
			for curKey, curFw := range profileFrameworkMap {
				var isOptional bool
				for _, optionalFramework := range f.getOptionalFrameworks(k) {
					if reflect.DeepEqual(optionalFramework, curFw) &&
						strings.EqualFold(optionalFramework.Profile, curFw.Profile) &&
						curFw.Version.Compare(optionalFramework.Version) >= 0 {
						isOptional = true
					}
				}
				if !isOptional {
					reduced[curKey] = curFw
				}
			}
			// check all frameworks while taking into account equivalent variations
			for _, permutationMap := range f.getEquivalentPermutations(vMap) {
				if len(reduced) != len(permutationMap) {
					continue
				}
				var isHasNoFound bool
				for _, fw := range reduced {
					if _, ok := permutationMap[fw.Framework]; !ok {
						isHasNoFound = true
						break
					}
				}
				// found a match
				if !isHasNoFound {
					return k
				}
			}
		}
		reduced = map[string]*Framework{}
	}
	return -1
}

func (f *FrameworkNameProvider) convertOrNormalize(key string, mappings, reverse map[string]string) string {
	if val, ok := mappings[key]; ok {
		return val
	}
	if _, ok := reverse[key]; !ok {
		return ""
	}
	for k := range reverse {
		if strings.EqualFold(key, k) {
			return k
		}
	}
	return ""
}

// GetEquivalentPermutations find all combinations that are equivalent
// ex: net4+win8 <-> net4+netcore45
func (f *FrameworkNameProvider) getEquivalentPermutations(frameworks map[string]*Framework) []map[string]*Framework {
	var results []map[string]*Framework

	if len(frameworks) == 0 {
		return results
	}

	// Select first framework from the set
	var current *Framework
	remaining := make(map[string]*Framework)
	first := true
	for k, fw := range frameworks {
		if first {
			current = fw
			first = false
			continue
		}
		remaining[k] = fw
	}

	// Get equivalent frameworks of current
	equalFrameworks := make(map[string]*Framework)
	if current != nil {
		if eqs, ok := f.equivalentFrameworkMap[current.Framework]; ok {
			equalFrameworks = eqs
		}
		// include ourselves
		equalFrameworks[current.Framework] = current
	}

	for _, fw := range equalFrameworks {
		if len(remaining) > 0 {
			subPermutations := f.getEquivalentPermutations(remaining)
			for _, perm := range subPermutations {
				perm[fw.Framework] = fw
				// work backwards adding the frameworks into the sets
				results = append(results, perm)
			}
		} else {
			results = append(results, map[string]*Framework{fw.Framework: fw})
		}
	}

	return results
}

func (f *FrameworkNameProvider) removeDuplicateFramework(supportedFrameworks []*Framework) map[string]*Framework {
	result := make(map[string]*Framework)
	existingFrameworks := make(map[string]*Framework)

	for _, framework := range supportedFrameworks {
		if _, ok := existingFrameworks[framework.Framework]; !ok {
			result[framework.Framework] = framework
			// Add in the existing framework (included here) and all equivalent frameworks
			for k, eq := range f.getAllEquivalentFrameworks(framework) {
				existingFrameworks[k] = eq
			}
		}
	}
	return result
}

// getAllEquivalentFrameworks Get all equivalent frameworks including the given framework
func (f *FrameworkNameProvider) getAllEquivalentFrameworks(framework *Framework) map[string]*Framework {
	// Loop through the frameworks, all frameworks that are not in results yet
	// will be added to toProcess to get the equivalent frameworks
	toProcess := []*Framework{framework}
	results := make(map[string]*Framework)
	results[framework.Framework] = framework

	for len(toProcess) > 0 {
		current := toProcess[len(toProcess)-1]
		toProcess = toProcess[:len(toProcess)-1]

		if equivalents, ok := f.equivalentFrameworkMap[current.Framework]; ok {
			for key, eq := range equivalents {
				if _, seen := results[key]; !seen {
					results[key] = eq
					toProcess = append(toProcess, eq)
				}
			}
		}
	}
	return results
}

func (f *FrameworkNameProvider) getOptionalFrameworks(profile int) map[string]*Framework {
	if frameworks, ok := f.portableOptionalFrameworkMap[profile]; ok {
		return frameworks
	}
	return make(map[string]*Framework)
}
