// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"fmt"
	"strings"

	"github.com/huhouhua/go-nuget"
)

// FrameworkReferenceGroup A group of FrameworkReference with the same target framework.
type FrameworkReferenceGroup struct {
	TargetFramework     *nuget.NuGetVersion
	FrameworkReferences []*FrameworkReference
}

type FrameworkReference struct {
	Name string
}

type PackageReferenceSet struct {
	References      []string
	TargetFramework *Framework
}

func (p *PackageReferenceSet) Validate() []string {
	var errs []string
	for _, reference := range p.References {
		if strings.TrimSpace(reference) == "" {
			errs = append(errs, "The required element File is missing from the manifest.")
		} else if strings.ContainsAny(reference, string(referenceFileInvalidCharacters)) {
			errs = append(errs, fmt.Sprintf("Assembly reference '%s' contains invalid characters.", reference))
		}
	}
	return errs
}
