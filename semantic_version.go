// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"fmt"
	"strings"
)

type SemanticVersion struct {
	releaseLabels []string `json:"releaseLabels"`

	metadata string `json:"metadata"`

	// Major version X (X. y. z)
	Major int `json:"major"`

	// Minor version Y (x. Y. z)
	Minor int `json:"minor"`

	// Patch version Z (x. y. Z)
	Patch int `json:"patch"`
}

func NewSemanticVersion(version *Version, releaseLabels []string, metadata string) (*SemanticVersion, error) {
	if version == nil {
		return nil, fmt.Errorf("version is nil")
	}
	normalizedVersion, err := normalizeVersionValue(version)
	if err != nil {
		return nil, err
	}
	semantic := &SemanticVersion{
		Major:         normalizedVersion.Major,
		Minor:         normalizedVersion.Minor,
		Patch:         normalizedVersion.Build,
		metadata:      metadata,
		releaseLabels: nil,
	}
	if releaseLabels != nil && len(releaseLabels) > 0 {
		semantic.releaseLabels = releaseLabels
	}
	return semantic, nil
}

// ReleaseLabels A collection of pre-release labels attached to the version.
func (semantic *SemanticVersion) ReleaseLabels() []string {
	if semantic.releaseLabels == nil {
		semantic.releaseLabels = make([]string, 0)
	}
	return semantic.releaseLabels
}

// Release The full pre-release label for the version.
func (semantic *SemanticVersion) Release() string {
	if semantic.releaseLabels == nil {
		return ""
	}
	if len(semantic.releaseLabels) == 1 {
		return semantic.releaseLabels[0]
	}
	return strings.Join(semantic.releaseLabels, ".")
}

// IsPrerelease True if pre-release labels exist for the version.
func (semantic *SemanticVersion) IsPrerelease() bool {
	if semantic.releaseLabels == nil {
		return false
	}
	for index := 0; index < len(semantic.releaseLabels); index++ {
		if strings.TrimSpace(semantic.releaseLabels[index]) != "" {
			return true
		}
	}
	return false
}

// HasMetadata True if metadata exists for the version.
func (semantic *SemanticVersion) HasMetadata() bool {
	return strings.TrimSpace(semantic.Metadata()) != ""
}

// Metadata Build metadata attached to the version.
func (semantic *SemanticVersion) Metadata() string {
	return semantic.metadata
}

// ParseSections Parse the version string into version/release/build The goal of
// this code is to take the most direct and optimized path to parsing and validating a semver.
// Regex would be much cleaner, but due to the number of versions created in NuGet Regex is too slow.
func parseSections(value string) (versionString string,
	releaseLabels []string, buildMetadata string) {
	dashPos, plusPos := -1, -1
	var end = false
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

func normalizeVersionValue(version *Version) (*Version, error) {
	var normalized = version
	var err error
	if version.Build < 0 || version.Revision < 0 {
		normalized, err = NewVersion(
			version.Major,
			version.Minor,
			mathMax(version.Build, 0),
			mathMax(version.Revision, 0))
		if err != nil {
			return nil, err
		}
	}
	return normalized, nil
}

func mathMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}
