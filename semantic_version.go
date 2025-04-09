// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import "strings"

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

// ParseSections Parse the version string into version/ release/ build The goal of
// this code is to take the most direct and optimized path to parsing and validating a semver.
// Regex would be much cleaner, but due to the number of versions created in NuGet Regex is too slow.
func (semantic *SemanticVersion) ParseSections(value string) (versionString string,
	releaseLabels []string, buildMetadata string) {
	num1, num2 := -1, -1
	for index := 0; index < len(value); index++ {
		flag := index == len(value)-1
		if num1 < 0 {
			if flag || value[index] == '-' || value[index] == '+' {
				length := index
				if flag {
					length += 1
				}
				versionString = value[0:length]
				num1 = index
				if value[index] == '+' {
					num2 = index
				}
			}
		} else if num2 < 0 {
			if flag || value[index] == '+' {
				startIndex := num1 + 1
				num3 := index
				if flag {
					num3 += 1
				}
				str := value[startIndex : num3-startIndex]
				releaseLabels = strings.Split(str, ".")
				num2 = index
			}
		} else if flag {
			startIndex := num2 + 1
			num4 := index + 1
			buildMetadata = value[startIndex : num4-startIndex]
		}
	}
	return "", nil, ""
}
