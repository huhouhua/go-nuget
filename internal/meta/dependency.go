// Copyright (c) 2025 Kevin Berger <huhouhuam@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package meta

import (
	"strings"

	"github.com/huhouhua/go-nuget/version"
)

// Dependency Represents a package dependency Id and allowed version range.
type Dependency struct {
	Id              string                `xml:"id,attr"      json:"id"`
	VersionRaw      string                `xml:"version,attr" json:"version"`
	ExcludeRaw      string                `xml:"exclude,attr" json:"exclude"`
	IncludeRaw      string                `xml:"include,attr" json:"include"`
	VersionRangeRaw string                `                   json:"range"`
	VersionRange    *version.VersionRange `xml:"-"`
	Include         []string              `xml:"-"`
	Exclude         []string              `xml:"-"`
}

// Parse parses the dependency version and splits the include/exclude strings into slices.
func (d *Dependency) Parse() error {
	if d.ExcludeRaw != "" {
		d.Exclude = strings.Split(d.ExcludeRaw, ",")
	}
	if d.IncludeRaw != "" {
		d.Exclude = strings.Split(d.IncludeRaw, ",")
	}
	if d.VersionRaw != "" {
		return d.parseRange(d.VersionRaw)
	}
	if d.VersionRangeRaw != "" {
		return d.parseRange(d.VersionRangeRaw)
	}
	return nil
}

func (d *Dependency) parseRange(rangeVersion string) error {
	if versionRanger, err := version.ParseRange(rangeVersion); err != nil {
		return err
	} else {
		d.VersionRange = versionRanger
		return nil
	}
}
