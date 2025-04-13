// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

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
	version, ok := parsedNuGetVersionsMapping[value]
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
		originalVersion := value
		var err error
		if strings.IndexAny(originalVersion, " ") > -1 {
			originalVersion = strings.ReplaceAll(value, " ", "")
		}
		version, err = NewNuGetVersion(version1, releaseLabels, buildMetadata, originalVersion)
		if err != nil {
			return false, nil, err
		}
		if len(parsedNuGetVersionsMapping) >= 500 {
			parsedNuGetVersionsMapping = make(map[string]*NuGetVersion)
		}
		parsedNuGetVersionsMapping[value] = version
		return true, version, nil
	}
	return false, nil, nil
}

func tryGetNormalizedVersion(str string) (*Version, bool) {
	str = strings.TrimSpace(str)
	if str == "" {
		return nil, false
	}

	lastParsedPosition, major, majorOk := parseSection(str, 0)
	lastParsedPosition, minor, minorOk := parseSection(str, lastParsedPosition)
	lastParsedPosition, build, buildOk := parseSection(str, lastParsedPosition)
	lastParsedPosition, revision, revisionOk := parseSection(str, lastParsedPosition)

	if majorOk && minorOk && buildOk && revisionOk && lastParsedPosition == len(str) {
		v, err := NewVersion(major, minor, build, revision)
		if err != nil {
			_ = fmt.Errorf(err.Error())
			return nil, false
		}
		return v, true
	}

	return nil, false
}

func parseSection(str string, start int) (end, versionNumber int, ok bool) {
	if start == len(str) {
		return start, 0, true
	}
	for end = start; end < len(str); end++ {
		ch := str[end]
		if ch != ' ' {
			if !isDigit(ch) {
				return end, 0, false
			}
			break
		}
	}
	var done, digitFound = false, false
	intermediateVersionNumber := int64(0)
	for ; end < len(str); end++ {
		ch := str[end]
		if isDigit(ch) {
			digitFound = true
			intermediateVersionNumber = intermediateVersionNumber*10 + int64(ch-'0')
			if intermediateVersionNumber > math.MaxInt32 {
				return end, 0, false
			}
		} else if ch == '.' {
			end++
			if end == len(str) {
				return end, 0, false
			}
			done = true
			break
		} else if ch != ' ' {
			break
		} else {
			return end, 0, false
		}
	}
	if !digitFound {
		return end, 0, false
	}
	if end == len(str) {
		done = true
	}
	if !done {
		for ; end < len(str); end++ {
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
		}

	}
	return end, int(intermediateVersionNumber), true
}

func isDigit(c uint8) bool {
	return c >= '0' && c <= '9'
}
