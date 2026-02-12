// Copyright (c) 2025 Kevin Berger <huhouhuam@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package creation

import (
	"fmt"
	"strings"

	"github.com/huhouhua/go-nuget/internal/framework"
	"github.com/huhouhua/go-nuget/internal/meta"
)

// FrameworkReferenceGroup A group of FrameworkReference with the same target framework.
type FrameworkReferenceGroup struct {
	TargetFramework     *framework.Framework
	FrameworkReferences []*FrameworkReference
}

type PackageDependencyGroup struct {
	TargetFramework *framework.Framework `json:"targetFramework"`
	Packages        []*meta.Dependency   `json:"dependencies"`
}

type FrameworkReference struct {
	Name string
}

type PackageReferenceSet struct {
	References      []string
	TargetFramework *framework.Framework
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
