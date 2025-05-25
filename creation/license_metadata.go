// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"fmt"
	"github.com/huhouhua/go-nuget"
	"net/url"
)

var (
	LicenseFileDeprecationURL  = url.URL{Scheme: "https", Host: "aka.ms", Path: "/deprecateLicenseUrl"}
	LicenseServiceLinkTemplate = "https://licenses.nuget.org/%s"
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
	version *nuget.NuGetVersion
}

func (l *LicenseMetadata) GetLicenseType() nuget.LicenseType {
	return l.licenseType
}

func (l *LicenseMetadata) GetLicense() string {
	return l.license
}
func (l *LicenseMetadata) GetVersion() *nuget.NuGetVersion {
	return l.version
}

func (l *LicenseMetadata) GetLicenseURL() (*url.URL, error) {
	switch l.licenseType {
	case nuget.File:
		return &LicenseFileDeprecationURL, nil
	case nuget.Expression:
		u, err := url.Parse(fmt.Sprintf(LicenseServiceLinkTemplate, l.license))
		if err != nil {
			return nil, err
		}
		return u, nil
	default:
		return nil, fmt.Errorf("unsupported license type: %v", l.licenseType)
	}
}
