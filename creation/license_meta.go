// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"fmt"
	"net/url"

	"github.com/huhouhua/go-nuget/version"
)

type LicenseType string

func (s LicenseType) String() string {
	return string(s)
}

const (
	File LicenseType = "file"

	Expression LicenseType = "expression"
)

// LicenseExpressionType Represents the expression type of a LicenseExpression.
// License type means that it's a License. Operator means that it's a LicenseOperator
type LicenseExpressionType int

const (
	License LicenseExpressionType = iota

	Operator LicenseExpressionType = iota
)

var (
	LicenseFileDeprecationURL  = url.URL{Scheme: "https", Host: "aka.ms", Path: "/deprecateLicenseUrl"}
	LicenseServiceLinkTemplate = "https://licenses.nuget.org/%s"
	LicenseEmptyVersion        = version.NewVersionFrom(1, 0, 0, "", "")
)

// LicenseExpression Represents a parsed LicenseExpression.
// This is an abstract class so based on the Type, it can be either a License or a LicenseOperator.
type LicenseExpression interface {
	// GetLicenseExpressionType The type of the LicenseExpression.
	// License type means that it's a License. Operator means that it's a LicenseOperator
	GetLicenseExpressionType()
}

type LicenseMetadata struct {

	// The LicenseType, never null
	licenseType LicenseType

	// license The license, never null, could be empty.
	license string

	// version LicenseMetadata (expression) version. Never null.
	version *version.Version
}

func NewLicense(licenseType LicenseType, license string, version *version.Version) *LicenseMetadata {
	return &LicenseMetadata{
		licenseType: licenseType,
		license:     license,
		version:     version,
	}
}

func (l *LicenseMetadata) GetLicenseType() LicenseType {
	return l.licenseType
}

func (l *LicenseMetadata) GetLicense() string {
	return l.license
}
func (l *LicenseMetadata) GetVersion() *version.Version {
	return l.version
}

func (l *LicenseMetadata) GetLicenseURL() (*url.URL, error) {
	switch l.licenseType {
	case File:
		return &LicenseFileDeprecationURL, nil
	case Expression:
		if u, err := url.Parse(fmt.Sprintf(LicenseServiceLinkTemplate, l.license)); err != nil {
			return nil, err
		} else {
			return u, nil
		}
	default:
		return nil, fmt.Errorf("unsupported license type: %v", l.licenseType)
	}
}
