// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"github.com/huhouhua/go-nuget"
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
	//licenseType nuget.LicenseType

	// license The license, never null, could be empty.
	license string

	// version LicenseMetadata (expression) version. Never null.
	version *nuget.NuGetVersion
}

func (l *LicenseMetadata) GetLicense() string {
	return l.license
}
func (l *LicenseMetadata) GetVersion() *nuget.NuGetVersion {
	return l.version
}
