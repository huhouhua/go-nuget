// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"strings"
)

// PackageFile is used in the NuSpec struct
type PackageFile struct {
	Source string `xml:"src,attr"`
	Target string `xml:"target,attr"`
}

type PackageFiles struct {
	File []*PackageFile `xml:"file"`
}

// Nuspec Represents a .nuspec XML file found in the root of the .nupck files
type Nuspec struct {
	XMLName  xml.Name     `xml:"package"`
	Xmlns    string       `xml:"xmlns,attr,omitempty"`
	Metadata *Metadata    `xml:"metadata"`
	Files    PackageFiles `xml:"files,omitempty"`
}

// ToBytes exports the nuspec to bytes in XML format
func (nsf *Nuspec) ToBytes() ([]byte, error) {
	var b bytes.Buffer
	// Unmarshal into XML
	output, err := xml.MarshalIndent(nsf, "", "  ")
	if err != nil {
		return nil, err
	}
	// Self-Close any empty XML elements (to match original Nuget output)
	// This assumes Indented Marshalling above, non Indented will break XML
	for bytes.Contains(output, []byte(`></`)) {
		i := bytes.Index(output, []byte(`></`))
		j := bytes.Index(output[i+1:], []byte(`>`))
		output = append(output[:i], append([]byte(` /`), output[i+j+1:]...)...)
	}
	// Write the XML Header
	b.WriteString(xml.Header)
	b.Write(output)
	return b.Bytes(), nil
}

// FromBytes parses a Nuspec file from a byte slice and returns a Nuspec struct.
func FromBytes(b []byte) (*Nuspec, error) {
	nsf := Nuspec{}
	err := xml.Unmarshal(b, &nsf)
	if err != nil {
		return nil, err
	}
	return &nsf, nil
}

// FromReader reads a Nuspec file from an io.ReadCloser, parses it, and returns a Nuspec struct.
// The reader will be fully read into memory.
func FromReader(r io.ReadCloser) (*Nuspec, error) {
	// Read contents of reader
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return FromBytes(b)
}

// FromFile reads a nuspec file from the file system
func FromFile(fn string) (*Nuspec, error) {
	// Open File
	xmlFile, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	return FromReader(xmlFile)
}

type PackageInfo struct {
	ID                       string              `xml:"id,omitempty"                       json:"ID,omitempty"`
	Version                  string              `xml:"version,omitempty"                  json:"version,omitempty"`
	Title                    string              `xml:"title,omitempty"                    json:"title,omitempty"`
	Authors                  string              `xml:"authors,omitempty"                  json:"authors,omitempty"`
	Owners                   string              `xml:"owners,omitempty"                   json:"owners,omitempty"`
	RequireLicenseAcceptance bool                `xml:"requireLicenseAcceptance,omitempty" json:"requireLicenseAcceptance,omitempty"`
	License                  string              `xml:"license,omitempty"                  json:"license,omitempty"`
	LicenseURL               string              `xml:"licenseUrl,omitempty"               json:"licenseURL,omitempty"`
	ProjectURL               string              `xml:"projectUrl,omitempty"               json:"projectURL,omitempty"`
	Readme                   string              `xml:"readme,omitempty"                   json:"readme,omitempty"`
	DevelopmentDependency    bool                `xml:"developmentDependency,omitempty"    json:"developmentDependency,omitempty"`
	Icon                     string              `xml:"icon,omitempty"                  json:"icon,omitempty"`
	IconURL                  string              `xml:"iconUrl,omitempty"                  json:"iconUrl,omitempty"`
	Description              string              `xml:"description,omitempty"              json:"description,omitempty"`
	Summary                  string              `xml:"summary,omitempty"                  json:"summary,omitempty"`
	ReleaseNotes             string              `xml:"releaseNotes,omitempty"             json:"releaseNotes,omitempty"`
	Copyright                string              `xml:"copyright,omitempty"                json:"copyright,omitempty"`
	Tags                     string              `xml:"tags,omitempty"                     json:"tags,omitempty"`
	Language                 string              `xml:"language,omitempty"                 json:"language,omitempty"`
	Serviceable              bool                `xml:"serviceable,omitempty"              json:"serviceable,omitempty"`
	PackageTypes             []*PackageType      `xml:"packageTypes,omitempty"              json:"packageTypes,omitempty"`
	Repository               *RepositoryMetadata `xml:"repository,omitempty"               json:"repository,omitempty"`
}

type Metadata struct {
	PackageInfo
	Dependencies        *Dependencies        `xml:"dependencies,omitempty"`
	FrameworkAssemblies *FrameworkAssemblies `xml:"frameworkAssemblies,omitempty"`
	References          *References          `xml:"references,omitempty"`
	FrameworkReferences *FrameworkReferences `xml:"frameworkReferences,omitempty"`
	ContentFile         *ContentFile         `xml:"contentFiles,omitempty"`
	MinClientVersion    string               `xml:"minClientVersion,attr"`
}

type PackageType struct {
	Name    string `xml:"name,attr"`
	Version string `xml:"version,attr"`
}

type RepositoryMetadata struct {
	Type   string `xml:"type,attr"`
	URL    string `xml:"url,attr"`
	Branch string `xml:"branch,attr"`
	Commit string `xml:"commit,attr"`
}

type Dependencies struct {
	Groups     []*DependenciesGroup `xml:"group"`
	Dependency []*Dependency        `xml:"dependency"`
}

type DependenciesGroup struct {
	TargetFramework string        `xml:"targetFramework,attr"`
	Dependencies    []*Dependency `xml:"dependency"`
}

// Dependency Represents a package dependency Id and allowed version range.
type Dependency struct {
	Id              string        `xml:"id,attr"      json:"id"`
	VersionRaw      string        `xml:"version,attr" json:"version"`
	ExcludeRaw      string        `xml:"exclude,attr" json:"exclude"`
	IncludeRaw      string        `xml:"include,attr" json:"include"`
	VersionRangeRaw string        `                   json:"range"`
	VersionRange    *VersionRange `xml:"-"`
	Include         []string      `xml:"-"`
	Exclude         []string      `xml:"-"`
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
	versionRanger, err := ParseVersionRange(rangeVersion)
	if err != nil {
		return err
	}
	d.VersionRange = versionRanger
	return nil
}

type References struct {
	Groups     []*ReferenceGroup `xml:"group"`
	References []*Reference      `xml:"reference"`
}

type ReferenceGroup struct {
	TargetFramework string       `xml:"targetFramework,attr"`
	References      []*Reference `xml:"reference"`
}

type FrameworkReferences struct {
	Groups []*FrameworkReferenceGroup `xml:"group"`
}

type FrameworkReferenceGroup struct {
	TargetFramework     string                `xml:"targetFramework,attr"`
	FrameworkReferences []*FrameworkReference `xml:"frameworkReference"`
}

type FrameworkAssemblies struct {
	FrameworkAssembly []*FrameworkAssembly `xml:"frameworkAssembly"`
}

type FrameworkAssembly struct {
	AssemblyName    []string `xml:"assemblyName,attr"`
	TargetFramework string   `xml:"targetFramework,attr"`
}

type Reference struct {
	File string `xml:"file,attr"`
}

type FrameworkReference struct {
	Name string `xml:"name,attr"`
}

type ContentFile struct {
	Files []*ContentFileItem `xml:"files"`
}

type ContentFileItem struct {
	Include      string `xml:"include,attr"`
	BuildAction  string `xml:"buildAction,attr"`
	CopyToOutput string `xml:"copyToOutput,attr"`
	Flatten      string `xml:"flatten,attr"`
}
