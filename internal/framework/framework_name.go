// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package framework

import (
	"fmt"
	"strings"

	nugetVersion "github.com/huhouhua/go-nuget/version"
)

type FrameworkName struct {
	identifier string
	profile    string
	version    nugetVersion.Version
}

// NewFrameworkName Parses strings in the following format: "<identifier>, Version=[v|V]<version>, Profile=<profile>"
//   - The identifier and version is required, profile is optional
//   - Only three components are allowed.
//   - The version string must be in the System.Version format; an optional "v" or "V" prefix is allowed
func NewFrameworkName(frameworkName string) (*FrameworkName, error) {
	if strings.TrimSpace(frameworkName) == "" {
		return nil, fmt.Errorf("frameworkName cannot be empty")
	}
	parts := strings.SplitN(frameworkName, ",", 4)
	if len(parts) != 2 && len(parts) != 3 {
		return nil, fmt.Errorf("frameworkName must have 2 or 3 components")
	}

	identifier := strings.TrimSpace(parts[0])
	if identifier == "" {
		return nil, fmt.Errorf("frameworkName identifier cannot be empty")
	}

	var versionStr string
	profile := ""
	versionFound := false

	for _, part := range parts[1:] {
		// Get the key/value pair separated by '='
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid component: %q", part)
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch strings.ToLower(key) {
		case "version":
			versionFound = true
			// Allow the version to include a 'v' or 'V' prefix...
			if len(value) > 0 && (value[0] == 'v' || value[0] == 'V') {
				value = value[1:]
			}
			versionStr = value
		case "profile":
			profile = value
		default:
			return nil, fmt.Errorf("invalid key: %q", key)
		}
	}
	if !versionFound {
		return nil, fmt.Errorf("frameworkName must contain a version")
	}
	version, err := nugetVersion.Parse(versionStr)
	if err != nil {
		return nil, fmt.Errorf("invalid version: %w", err)
	}
	return &FrameworkName{
		identifier: identifier,
		version:    *version,
		profile:    profile,
	}, nil
}

func (f *FrameworkName) GetVersion() nugetVersion.Version {
	return f.version
}
func (f *FrameworkName) GetIdentifier() string {
	return f.identifier
}
func (f *FrameworkName) GetProfile() string {
	return f.profile
}
