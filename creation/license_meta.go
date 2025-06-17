// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"fmt"
	"net/url"

	"github.com/Masterminds/semver/v3"

	"github.com/huhouhua/go-nuget"
)

var (
	LicenseFileDeprecationURL  = url.URL{Scheme: "https", Host: "aka.ms", Path: "/deprecateLicenseUrl"}
	LicenseServiceLinkTemplate = "https://licenses.nuget.org/%s"
	LicenseEmptyVersion        = semver.New(1, 0, 0, "", "")
)

// LicenseExpression Represents a parsed NuGetLicenseExpression.
// This is an abstract class so based on the Type, it can be either a NuGetLicense or a LicenseOperator.
type LicenseExpression interface {
	// GetLicenseExpressionType The type of the NuGetLicenseExpression.
	// License type means that it's a NuGetLicense. Operator means that it's a LicenseOperator
	GetLicenseExpressionType()
}

type LicenseMetadata struct {

	// The LicenseType, never null
	licenseType nuget.LicenseType

	// license The license, never null, could be empty.
	license string

	// version LicenseMetadata (expression) version. Never null.
	version *semver.Version
}

func NewLicense(licenseType nuget.LicenseType, license string, version *semver.Version) *LicenseMetadata {
	return &LicenseMetadata{
		licenseType: licenseType,
		license:     license,
		version:     version,
	}
}

func (l *LicenseMetadata) GetLicenseType() nuget.LicenseType {
	return l.licenseType
}

func (l *LicenseMetadata) GetLicense() string {
	return l.license
}
func (l *LicenseMetadata) GetVersion() *semver.Version {
	return l.version
}

func (l *LicenseMetadata) GetLicenseURL() (*url.URL, error) {
	switch l.licenseType {
	case nuget.File:
		return &LicenseFileDeprecationURL, nil
	case nuget.Expression:
		if u, err := url.Parse(fmt.Sprintf(LicenseServiceLinkTemplate, l.license)); err != nil {
			return nil, err
		} else {
			return u, nil
		}
	default:
		return nil, fmt.Errorf("unsupported license type: %v", l.licenseType)
	}
}
