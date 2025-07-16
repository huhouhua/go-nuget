// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package version

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/Masterminds/semver/v3"
)

var (
	parsedVersionsMapping versionsMapping
)

func init() {
	parsedVersionsMapping = versionsMapping{
		versionMap: make(map[string]Version),
	}
}

type versionsMapping struct {
	mu         sync.RWMutex
	versionMap map[string]Version
}

func (c *versionsMapping) setVersion(key string, version Version) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.versionMap) >= 500 {
		c.versionMap = make(map[string]Version)
	}
	c.versionMap[key] = version
}

func (c *versionsMapping) getVersion(key string) (Version, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	version, ok := c.versionMap[key]
	return version, ok
}

type Version struct {
	Semver *semver.Version `json:"semver"`
	// Revision version R (x. y. z. R)
	Revision        int    `json:"revision"`
	OriginalVersion string `json:"originalVersion"`
}

// IsLegacyVersion returns True if the Version is using legacy behavior.
func (v *Version) IsLegacyVersion() bool {
	return v.Revision > 0
}

// IsSemVer2 returns true if version is a SemVer 2.0.0 version
func (v *Version) IsSemVer2() bool {
	return strings.TrimSpace(v.Semver.Prerelease()) != "" || strings.TrimSpace(v.Semver.Metadata()) != ""
}

func NewVersion(semver *semver.Version, revision int,
	originalVersion string) *Version {
	v := &Version{
		Semver:          semver,
		Revision:        revision,
		OriginalVersion: originalVersion,
	}
	return v
}

func NewVersionFrom(major, minor, patch uint64, pre, metadata string) *Version {
	v := semver.New(major, minor, patch, pre, metadata)
	return &Version{
		Semver:          v,
		OriginalVersion: v.Original(),
	}
}

// Parse a Version from a string representing the semantic version.
func Parse(value string) (*Version, error) {
	if strings.TrimSpace(value) == "" {
		return nil, errors.New("argument cannot be null or empty")
	}

	ok, version, err := TryParse(value)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("invalid semantic version")
	}
	return version, nil
}

func TryParse(value string) (bool, *Version, error) {
	if strings.TrimSpace(value) == "" {
		return false, nil, fmt.Errorf("argument cannot be null or empty")
	}
	if v, ok := parsedVersionsMapping.getVersion(value); ok {
		return true, &v, nil
	}
	if semVersion, err := semver.NewVersion(value); err == nil {
		v := NewVersion(semVersion, 0, value)
		parsedVersionsMapping.setVersion(value, *v)
		return true, v, nil
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
		if strings.ContainsAny(originalVersion, " ") {
			originalVersion = strings.ReplaceAll(value, " ", "")
		}
		version, err := ConvertVersion(version1, originalVersion, buildMetadata, releaseLabels)
		if err != nil {
			return false, nil, err
		}
		parsedVersionsMapping.setVersion(originalVersion, *version)
		return true, version, nil
	}
	return false, nil, nil
}

func tryGetNormalizedVersion(str string) (*SemanticVersion, bool) {
	str = strings.TrimSpace(str)
	if str == "" {
		return nil, false
	}

	lastParsedPosition, major, majorOk := parseSection(str, 0)
	lastParsedPosition, minor, minorOk := parseSection(str, lastParsedPosition)
	lastParsedPosition, build, buildOk := parseSection(str, lastParsedPosition)
	lastParsedPosition, revision, revisionOk := parseSection(str, lastParsedPosition)
	if majorOk && minorOk && buildOk && revisionOk && lastParsedPosition == len(str) {
		return &SemanticVersion{
			Major:    major,
			Minor:    minor,
			Build:    build,
			Revision: revision,
		}, true
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

type SemanticVersion struct {
	Major    int
	Minor    int
	Build    int
	Revision int
}

func ConvertVersion(
	version *SemanticVersion,
	originalVersion, metadata string,
	releaseLabels []string,
) (*Version, error) {
	if version == nil {
		return nil, fmt.Errorf("version is nil")
	}
	normalizedVersion := normalizeVersionValue(version)
	release := ""
	if len(releaseLabels) > 0 {
		release = strings.Join(releaseLabels, ".")
	}
	v := semver.New(uint64(normalizedVersion.Major), uint64(normalizedVersion.Minor),
		uint64(normalizedVersion.Build), release, metadata)
	return NewVersion(v, normalizedVersion.Revision, originalVersion), nil
}

// parseSections Parse the version string into version/release/build The goal of
// this code is to take the most direct and optimized path to parsing and validating a semver.
// Regex would be much cleaner, but due to the number of versions created in NuGet Regex is too slow.
func parseSections(value string) (versionString string,
	releaseLabels []string, buildMetadata string) {
	dashPos, plusPos := -1, -1
	var end bool
	for index := 0; index < len(value); index++ {
		end = index == len(value)-1
		if dashPos < 0 {
			if end || value[index] == '-' || value[index] == '+' {
				endPos := index
				if end {
					endPos += 1
				}
				versionString = value[0:endPos]
				dashPos = index
				if value[index] == '+' {
					plusPos = index
				}
			}
		} else if plusPos < 0 {
			if end || value[index] == '+' {
				start := dashPos + 1
				endPos := index
				if end {
					endPos += 1
				}
				str := value[start:endPos]
				releaseLabels = strings.Split(str, ".")
				plusPos = index
			}
		} else if end {
			start := plusPos + 1
			endPos := index + 1
			buildMetadata = value[start:endPos]
		}
	}
	return versionString, releaseLabels, buildMetadata
}

func isValidPart(s string, allowLeadingZeros bool) bool {
	if len(s) == 0 {
		// empty labels are not allowed
		return false
	}
	// 0 is fine, but 00 is not.
	// 0A counts as an alpha numeric string where zeros are not counted
	if !allowLeadingZeros && len(s) > 1 && s[0] == '0' {
		var allDigits = true

		// Check if all characters are digits.
		// The first is already checked above
		for i := 1; i < len(s); i++ {
			if !isDigit(s[i]) {
				allDigits = false
				break
			}
		}

		if allDigits {
			// leading zeros are not allowed in numeric labels
			return false
		}
	}
	for i := 0; i < len(s); i++ {
		if !((s[i] >= 48 && s[i] <= 57) || (s[i] >= 65 && s[i] <= 90) || (s[i] >= 97 && s[i] <= 122) || s[i] == 45) {
			return false
		}
	}
	return true
}

func isValid(s string, allowLeadingZeros bool) bool {
	parts := strings.Split(s, ".")

	// Check each part individually
	for i := 0; i < len(parts); i++ {
		if !isValidPart(parts[i], allowLeadingZeros) {
			return false
		}
	}
	return true
}

func normalizeVersionValue(version *SemanticVersion) *SemanticVersion {
	if version.Build < 0 || version.Revision < 0 {
		return &SemanticVersion{
			Major:    version.Major,
			Minor:    version.Minor,
			Build:    mathMax(version.Build, 0),
			Revision: mathMax(version.Revision, 0)}
	}
	return version
}

func mathMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}
