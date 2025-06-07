// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/huhouhua/go-nuget"

	"github.com/Masterminds/semver/v3"
)

var (
	singleDigitVersionFrameworkMap = map[string]bool{
		"Windows":      true,
		"WindowsPhone": true,
		"Silverlight":  true,
	}
)

type FrameworkNameProvider struct {

	// identifierSynonyms Contains identifier -> identifier
	// Ex: .NET Framework -> .NET Framework
	// Ex: NET Framework -> .NET Framework
	// This includes self mappings.
	identifierSynonymsMap    map[string]string
	identifierToShortNameMap map[string]string
	identifierShortToLongMap map[string]string

	// profile -> supported frameworks, optional frameworks
	portableFrameworkMap         map[int][]*Framework
	portableOptionalFrameworkMap map[int][]*Framework

	profileShortToLongMap  map[string]string
	profilesToShortNameMap map[string]string

	// equivalent frameworks
	equivalentFrameworkMap []*KeyValuePair[*Framework, []*Framework]

	// Rewrite mappings
	shortNameRewrites []*KeyValuePair[*Framework, *Framework]
	fullNameRewrites  []*KeyValuePair[*Framework, *Framework]
}

func NewFrameworkNameProvider(
	mappings []FrameworkMappings,
	portableMappings []PortableFrameworkMappings,
) *FrameworkNameProvider {
	provider := &FrameworkNameProvider{
		identifierSynonymsMap:        make(map[string]string),
		identifierToShortNameMap:     make(map[string]string),
		identifierShortToLongMap:     make(map[string]string),
		portableFrameworkMap:         make(map[int][]*Framework),
		portableOptionalFrameworkMap: make(map[int][]*Framework),
		profileShortToLongMap:        make(map[string]string),
		profilesToShortNameMap:       make(map[string]string),
		equivalentFrameworkMap:       make([]*KeyValuePair[*Framework, []*Framework], 0),
		shortNameRewrites:            make([]*KeyValuePair[*Framework, *Framework], 0),
		fullNameRewrites:             make([]*KeyValuePair[*Framework, *Framework], 0),
	}

	provider.initMappings(mappings)
	provider.initPortableMappings(portableMappings)
	return provider
}
func (f *FrameworkNameProvider) GetIdentifier(framework string) string {
	return f.convertOrNormalize(framework, f.identifierSynonymsMap, f.identifierToShortNameMap)
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

func (f *FrameworkNameProvider) GetVersionString(framework string, version *semver.Version) string {
	if version.Major() == 0 && version.Minor() == 0 && version.Patch() == 0 {
		return ""
	}
	stack := []int{int(version.Major()), int(version.Minor()), int(version.Patch())}
	minCount := 2
	if singleDigitVersionFrameworkMap[framework] {
		minCount = 1
	}
	for len(stack) > minCount && stack[len(stack)-1] == 0 {
		stack = stack[:len(stack)-1]
	}
	hasDoubleDigit := strings.EqualFold(framework, ".NETCoreApp") ||
		strings.EqualFold(framework, ".NETStandard") ||
		anyGreaterThanNine(stack)
	if hasDoubleDigit {
		if len(stack) < 2 {
			stack = append(stack, 0)
		}
		return joinInts(stack, ".")
	}
	return joinInts(stack, "")
}
func anyGreaterThanNine(stack []int) bool {
	for _, v := range stack {
		if v > 9 {
			return true
		}
	}
	return false
}
func joinInts(nums []int, sep string) string {
	strs := make([]string, len(nums))
	for i, n := range nums {
		strs[i] = strconv.Itoa(n)
	}
	return strings.Join(strs, sep)
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
func (f *FrameworkNameProvider) GetShortIdentifier(identifier string) string {
	return f.convertOrNormalize(identifier, f.identifierToShortNameMap, f.identifierShortToLongMap)
}
func (f *FrameworkNameProvider) GetShortProfile(profile string) string {
	return f.convertOrNormalize(profile, f.profilesToShortNameMap, f.profileShortToLongMap)
}
func TryGetPortableProfileNumber(profile string) (int, bool) {
	if strings.HasPrefix(strings.ToLower(profile), "profile") {
		numStr := profile[7:] // Skip "Profile"
		if n, err := strconv.Atoi(numStr); err == nil {
			return n, true
		}
	}
	return -1, false
}

func (f *FrameworkNameProvider) GetPortableFrameworksWithInclude(
	profile string,
	includeOptional bool,
) ([]*Framework, error) {
	if profileNumber, ok := TryGetPortableProfileNumber(profile); !ok {
		return f.GetPortableFrameworks(profile)
	} else {
		if frameworks := f.getPortableFrameworksWithInclude(profileNumber, includeOptional); frameworks != nil && len(frameworks) > 0 {
			return frameworks, nil
		}
	}
	return make([]*Framework, 0), nil
}
func (f *FrameworkNameProvider) getPortableFrameworksWithInclude(profile int, includeOptional bool) []*Framework {
	var nuGetFrameworkSet1 []*Framework
	if nuGetFrameworkSet2, ok := f.portableFrameworkMap[profile]; ok {
		for _, nuGetFramework := range nuGetFrameworkSet2 {
			nuGetFrameworkSet1 = append(nuGetFrameworkSet1, nuGetFramework)
		}
	}
	if includeOptional {
		if nuGetFrameworkSet3, ok := f.portableOptionalFrameworkMap[profile]; ok {
			for _, nuGetFramework := range nuGetFrameworkSet3 {
				nuGetFrameworkSet1 = append(nuGetFrameworkSet1, nuGetFramework)
			}
		}
	}
	return nuGetFrameworkSet1
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
	reduced := make([]*Framework, 0)
	for k, vMap := range f.portableFrameworkMap {
		// to match the required set must be less than or the same count as the input
		// if we knew which frameworks were optional in the input we could rule out the lesser ones also
		if len(vMap) <= len(profileFrameworkMap) {
			for _, curFw := range profileFrameworkMap {
				var isOptional bool
				for _, optionalFramework := range f.getOptionalFrameworks(k) {
					if reflect.DeepEqual(optionalFramework, curFw) &&
						strings.EqualFold(optionalFramework.Profile, curFw.Profile) &&
						curFw.Version.Compare(optionalFramework.Version) >= 0 {
						isOptional = true
					}
				}
				if !isOptional {
					reduced = append(reduced, curFw)
				}
			}
			// check all frameworks while taking into account equivalent variations
			for _, permutationMap := range f.getEquivalentPermutations(vMap) {
				if len(reduced) != len(permutationMap) {
					continue
				}
				var isHasNoFound bool
				for _, fw := range reduced {
					ok := nuget.Some(permutationMap, func(framework *Framework) bool {
						return framework.Equals(fw)
					})
					if !ok {
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
		reduced = []*Framework{}
	}
	return -1
}
func (f *FrameworkNameProvider) GetShortNameReplacement(framework *Framework) *Framework {
	for _, rewrite := range f.shortNameRewrites {
		if rewrite.Key.Equals(framework) {
			return rewrite.Value
		}
	}
	return framework
}

func (f *FrameworkNameProvider) GetFullNameReplacement(framework *Framework) *Framework {
	for _, rewrite := range f.fullNameRewrites {
		if rewrite.Key.Equals(framework) {
			return rewrite.Value
		}
	}
	return framework
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
func (f *FrameworkNameProvider) getEquivalentPermutations(frameworks []*Framework) [][]*Framework {
	var results [][]*Framework

	if len(frameworks) == 0 {
		return results
	}

	// Select first framework from the set
	var current *Framework
	remaining := make([]*Framework, 0)
	first := true
	for _, fw := range frameworks {
		if first {
			current = fw
			first = false
			continue
		}
		remaining = append(remaining, fw)
	}

	// Get equivalent frameworks of current
	equalFrameworks := make([]*Framework, 0)
	if current != nil {
		for _, keyValue := range f.equivalentFrameworkMap {
			if keyValue.Key.Equals(current) {
				equalFrameworks = keyValue.Value
				break
			}
		}
		// include ourselves
		equalFrameworks = append(equalFrameworks, current)
	}

	for _, fw := range equalFrameworks {
		if len(remaining) > 0 {
			subPermutations := f.getEquivalentPermutations(remaining)
			for _, perm := range subPermutations {
				perm = append(perm, fw)
				// work backwards adding the frameworks into the sets
				results = append(results, perm)
			}
		} else {
			results = append(results, []*Framework{fw})
		}
	}

	return results
}

func (f *FrameworkNameProvider) removeDuplicateFramework(supportedFrameworks []*Framework) []*Framework {
	result := make([]*Framework, 0)
	existingFrameworks := make([]*Framework, 0)

	for _, framework := range supportedFrameworks {
		ok := nuget.Some(existingFrameworks, func(fw *Framework) bool {
			return fw.Equals(framework)
		})
		if !ok {
			result = append(result, framework)
			// Add in the existing framework (included here) and all equivalent frameworks
			for _, eq := range f.getAllEquivalentFrameworks(framework) {
				existingFrameworks = append(existingFrameworks, eq)
			}
		}
	}
	return result
}

// getAllEquivalentFrameworks Get all equivalent frameworks including the given framework
func (f *FrameworkNameProvider) getAllEquivalentFrameworks(framework *Framework) []*Framework {
	// Loop through the frameworks, all frameworks that are not in results yet
	// will be added to toProcess to get the equivalent frameworks
	toProcess := []*Framework{framework}
	results := make([]*Framework, 0)
	results = append(results, framework)
	for len(toProcess) > 0 {
		current := toProcess[len(toProcess)-1]
		toProcess = toProcess[:len(toProcess)-1]

		for _, keyValue := range f.equivalentFrameworkMap {
			if keyValue.Key.Equals(current) {
				for _, eq := range keyValue.Value {
					seen := nuget.Some(results, func(fw *Framework) bool {
						return fw.Equals(eq)
					})
					if !seen {
						results = append(results, eq)
						toProcess = append(toProcess, eq)
					}
				}
				break
			}
		}

	}
	return results
}

func (f *FrameworkNameProvider) getOptionalFrameworks(profile int) []*Framework {
	if frameworks, ok := f.portableOptionalFrameworkMap[profile]; ok {
		return frameworks
	}
	return make([]*Framework, 0)
}

func (p *FrameworkNameProvider) initMappings(mappings []FrameworkMappings) {
	if mappings == nil {
		return
	}
	for _, mapping := range mappings {
		// equivalent frameworks
		p.addEquivalentFrameworks(mapping.GetEquivalentFrameworkMap())

		// add synonyms
		p.addFrameworkSynonyms(mapping.GetIdentifierSynonymsMap())

		// populate short <-> long
		p.addIdentifierShortNames(mapping.GetIdentifierShortNameMap())

		// add rewrite rules
		p.addShortNameRewriteMappings(mapping.GetShortNameReplacementMap())
		p.addFullNameRewriteMappings(mapping.GetFullNameReplacementMap())
	}
}

func (p *FrameworkNameProvider) initPortableMappings(portableMappings []PortableFrameworkMappings) {
	if portableMappings == nil {
		return
	}
	for _, portableMapping := range portableMappings {
		// populate portable framework names
		p.addPortableProfileMappings(portableMapping.GetProfileFrameworkMap())

		// populate portable optional frameworks
		p.addPortableOptionalFrameworks(portableMapping.GetProfileOptionalFrameworkMap())
	}
}

// addEquivalentFrameworks  2 way framework equivalence
func (p *FrameworkNameProvider) addEquivalentFrameworks(mappings []*KeyValuePair[*Framework, *Framework]) {
	if mappings == nil {
		return
	}
	for _, pair := range mappings {
		remaining := []*Framework{pair.Key, pair.Value}
		seen := make([]*Framework, 0)

		for len(remaining) > 0 {
			n := len(remaining) - 1
			next := remaining[n]
			remaining = remaining[:n]

			ok := nuget.Some(seen, func(framework *Framework) bool {
				return framework.Equals(next)
			})
			if ok {
				continue
			}
			seen = append(seen, next)
			found := false
			eqSet := make([]*Framework, 0)
			for _, keyValue := range p.equivalentFrameworkMap {
				if keyValue.Key.Equals(next) {
					found = true
					eqSet = keyValue.Value
				}
			}
			if !found {
				// initialize set
				eqSet = make([]*Framework, 0)
				p.equivalentFrameworkMap = append(p.equivalentFrameworkMap, &KeyValuePair[*Framework, []*Framework]{
					Key:   next,
					Value: eqSet,
				})
			} else {
				// explore all equivalent
				for _, value := range eqSet {
					remaining = append(remaining, value)
				}
			}
		}

		// add this equivalency rule, enforcing transitivity
		for _, framework := range seen {
			for _, other := range seen {
				if !framework.Equals(other) {
					for _, keyValue := range p.equivalentFrameworkMap {
						if keyValue.Key.Equals(framework) {
							keyValue.Value = append(keyValue.Value, other)
							break
						}
					}
				}
			}
		}
	}
}

func (p *FrameworkNameProvider) addFrameworkSynonyms(mappings []*KeyValuePair[string, string]) {
	if mappings == nil {
		return
	}
	for _, pair := range mappings {
		if _, ok := p.identifierSynonymsMap[pair.Key]; !ok {
			p.identifierSynonymsMap[pair.Key] = pair.Value
		}
	}
}

func (p *FrameworkNameProvider) addIdentifierShortNames(mappings []*KeyValuePair[string, string]) {
	if mappings == nil {
		return
	}
	for _, pair := range mappings {
		shortName := pair.Value
		longName := pair.Key
		if _, ok := p.identifierSynonymsMap[pair.Value]; !ok {
			p.identifierSynonymsMap[pair.Value] = pair.Key
		}
		p.identifierShortToLongMap[shortName] = longName
		p.identifierToShortNameMap[longName] = shortName
	}
}

// addPortableProfileMappings Add supported frameworks for each portable profile number
func (p *FrameworkNameProvider) addPortableProfileMappings(mappings []*KeyValuePair[int, []*Framework]) {
	if mappings == nil {
		return
	}
	for _, pair := range mappings {

		if _, ok := p.portableFrameworkMap[pair.Key]; !ok {
			p.portableFrameworkMap[pair.Key] = []*Framework{}
		}
		frameworks, _ := p.portableFrameworkMap[pair.Key]
		for _, fw := range pair.Value {
			frameworks = append(frameworks, fw)
		}
	}
}

// addPortableOptionalFrameworks Add optional frameworks for each portable profile number
func (p *FrameworkNameProvider) addPortableOptionalFrameworks(mappings []*KeyValuePair[int, []*Framework]) {
	if mappings == nil {
		return
	}
	for _, pair := range mappings {
		if _, ok := p.portableOptionalFrameworkMap[pair.Key]; !ok {
			p.portableOptionalFrameworkMap[pair.Key] = []*Framework{}
		}
		frameworks, _ := p.portableOptionalFrameworkMap[pair.Key]
		for _, fw := range pair.Value {
			frameworks = append(frameworks, fw)
		}
	}
}

func (p *FrameworkNameProvider) addShortNameRewriteMappings(mappings []*KeyValuePair[*Framework, *Framework]) {
	if mappings == nil {
		return
	}
	for _, mapping := range mappings {
		hasContains := nuget.Some(p.shortNameRewrites, func(k *KeyValuePair[*Framework, *Framework]) bool {
			return k.Key.Equals(mapping.Key)
		})
		if !hasContains {
			p.shortNameRewrites = append(p.shortNameRewrites, mapping)
		}
	}
}

func (p *FrameworkNameProvider) addFullNameRewriteMappings(mappings []*KeyValuePair[*Framework, *Framework]) {
	if mappings == nil {
		return
	}
	for _, mapping := range mappings {
		hasContains := nuget.Some(p.fullNameRewrites, func(k *KeyValuePair[*Framework, *Framework]) bool {
			return k.Key.Equals(mapping.Key)
		})
		if !hasContains {
			p.fullNameRewrites = append(p.fullNameRewrites, mapping)
		}
	}
}
