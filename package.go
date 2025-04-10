// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"unicode"
)

var (
	rwMu                       sync.RWMutex
	ParsedNuGetVersionsMapping map[string]*NuGetVersion
)

type PackageResource struct {
}

type NuGetVersion struct {
	*SemanticVersion

	Version *Version `json:"version"`
	// Revision version R (x. y. z. R)
	Revision        int    `json:"revision"`
	OriginalVersion string `json:"originalVersion"`
}

func NewNuGetVersion(version *Version, releaseLabels []string, metadata string, originalVersion string) (*NuGetVersion, error) {
	semantic, err := NewSemanticVersion(version, releaseLabels, metadata)
	if err != nil {
		return nil, err
	}
	normalizedVersion, err := normalizeVersionValue(version)
	if err != nil {
		return nil, err
	}
	return &NuGetVersion{
		SemanticVersion: semantic,
		Version:         normalizedVersion,
		Revision:        normalizedVersion.Revision,
		OriginalVersion: originalVersion,
	}, nil
}

// IsLegacyVersion returns True if the NuGetVersion is using legacy behavior.
func (resource *NuGetVersion) IsLegacyVersion() bool {
	return resource.Revision > 0
}

// IsSemVer2 returns true if version is a SemVer 2.0.0 version
func (resource *NuGetVersion) IsSemVer2() bool {
	return resource.releaseLabels != nil && len(resource.releaseLabels) > 1 || resource.HasMetadata()
}

// Parse Creates a NuGetVersion from a string representing the semantic version.
func Parse(value string) (*NuGetVersion, error) {
	if strings.TrimSpace(value) == "" {
		return nil, errors.New("argument cannot be null or empty")
	}

	ok, version, err := TryParse(value)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("value invalid")
	}
	return version, nil
}

func TryParse(value string) (bool, *NuGetVersion, error) {
	if strings.TrimSpace(value) == "" {
		return false, nil, fmt.Errorf("value is null or empty")
	}
	rwMu.RLock()
	defer rwMu.RUnlock()
	version, ok := ParsedNuGetVersionsMapping[value]
	if ok {
		return ok, version, nil
	}
	versionString, releaseLabels, buildMetadata := parseSections(strings.TrimSpace(value))
	if strings.TrimSpace(versionString) == "" {
		return false, nil, fmt.Errorf("versionString is null or empty")
	}

	if version1, vok := tryGetNormalizedVersion(versionString); vok {
		if releaseLabels != nil {
			for index := 0; index < len(releaseLabels); index++ {
				if !isValidPart(releaseLabels[index], false) {
					return false, nil, nil
				}
			}
		}
		if strings.TrimSpace(buildMetadata) != "" && !isValid(buildMetadata, true) {
			return false, nil, nil
		}
		str := value
		var err error
		if strings.IndexAny(str, " ") > -1 {
			str = strings.ReplaceAll(value, " ", "")
			version, err = NewNuGetVersion(version1, releaseLabels, buildMetadata, str)
			if err != nil {
				return false, nil, err
			}
			if len(ParsedNuGetVersionsMapping) >= 500 {
				ParsedNuGetVersionsMapping = make(map[string]*NuGetVersion)
				ParsedNuGetVersionsMapping[value] = version
			}
			return true, version, nil
		}
	}
	return false, nil, nil
}

func tryGetNormalizedVersion(str string) (*Version, bool) {
	str = strings.TrimSpace(str)
	if str == "" {
		return nil, false
	}

	var versions [4]int
	end := 0
	var ok bool

	for i := 0; i < 4; i++ {
		versions[i], end, ok = parseSection(str, end)
		if !ok {
			return nil, false
		}
	}
	if strings.TrimSpace(str[end:]) != "" {
		return nil, false
	}
	v, err := NewVersion(versions[0], versions[1], versions[2], versions[3])
	if err != nil {
		_ = fmt.Errorf(err.Error())
		return nil, false
	}
	return v, true
}

func parseSection(str string, start int) (end, versionNumber int, ok bool) {
	if start == len(str) {
		return start, 0, false
	}
	end = start
	for end < len(str) {
		ch := str[end]
		if ch != ' ' {
			if !unicode.IsDigit(rune(ch)) {
				return end, 0, false
			}
			break
		}
		end++
	}
	var flag1, flag2 = false, false
	num := int64(0)
	for end < len(str) {
		ch := str[end]
		if unicode.IsDigit(rune(ch)) {
			flag2 = true
			num = num*10 + int64(ch-'0')
			if num > math.MaxInt32 {
				return end, 0, false
			}
			end++
		} else {
			if ch == '.' {
				end++
				if end == len(str) {
					return end, 0, false
				}
				flag1 = true
				break
			}
			if ch != ' ' {
				return end, 0, false
			}
			break
		}
	}
	if !flag2 {
		return end, 0, false
	}
	if end == len(str) {
		flag1 = true
	}
	if !flag1 {
		for end < len(str) {
			ch := str[end]
			if ch != ' ' {
				if ch == '.' {
					end++
					if end == len(str) {
						return end, 0, false
					}
					break
				}
				return end, 0, false
			}
			end++
		}
	}
	return end, int(num), true
}
