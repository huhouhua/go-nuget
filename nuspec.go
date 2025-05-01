// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"sync"
)

type Nuspec struct {
	XMLName  xml.Name  `xml:"package"`
	Metadata *Metadata `xml:"metadata"`
}

type PackageInfo struct {
	ID                       string      `xml:"id"`
	Version                  string      `xml:"version"`
	Authors                  string      `xml:"authors"`
	Owners                   string      `xml:"owners"`
	RequireLicenseAcceptance bool        `xml:"requireLicenseAcceptance"`
	License                  string      `xml:"license"`
	LicenseURL               string      `xml:"licenseUrl"`
	ProjectURL               string      `xml:"projectUrl"`
	IconUrl                  string      `xml:"iconUrl"`
	Description              string      `xml:"description"`
	Summary                  string      `xml:"summary"`
	ReleaseNotes             string      `xml:"releaseNotes"`
	Copyright                string      `xml:"copyright"`
	Tags                     string      `xml:"tags"`
	Language                 string      `xml:"language"`
	Repository               *Repository `xml:"repository"`
}

type Metadata struct {
	PackageInfo
	Dependencies        *Dependencies        `xml:"dependencies"`
	FrameworkAssemblies *FrameworkAssemblies `xml:"frameworkAssemblies"`
	References          *References          `xml:"references"`
}

type Repository struct {
	Type   string `xml:"type,attr"`
	URL    string `xml:"url,attr"`
	Branch string `xml:"branch,attr"`
	Commit string `xml:"commit,attr"`
}

type Dependencies struct {
	Groups     []*DependenciesGroup `xml:"group"`
	Dependency []*Dependency        `xml:"dependency" `
}

type DependenciesGroup struct {
	TargetFramework string        `xml:"targetFramework,attr"`
	Dependencies    []*Dependency `xml:"dependency" `
}

// Dependency Represents a package dependency Id and allowed version range.
type Dependency struct {
	Id              string        `xml:"id,attr" json:"id"`
	VersionRaw      string        `xml:"version,attr" json:"version"`
	ExcludeRaw      string        `xml:"exclude,attr" json:"exclude"`
	IncludeRaw      string        `xml:"include,attr" json:"include"`
	VersionRangeRaw string        `json:"range"`
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

type PackageArchiveReader struct {
	nuspec     *Nuspec
	writer     io.Writer
	archive    *zip.Reader
	nuspecFile io.ReadCloser
	once       sync.Once
}

func NewPackageArchiveReader(r io.Writer) (*PackageArchiveReader, error) {
	p := &PackageArchiveReader{
		writer: r,
	}
	if err := p.parse(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *PackageArchiveReader) parse() error {
	// Ensure p.writer is a *bytes.Buffer
	buf, ok := p.writer.(*bytes.Buffer)
	if !ok {
		return fmt.Errorf("expected *bytes.Buffer, got %T", p.writer)
	}
	// Create a zip reader from the buffer
	r := buf.Bytes()
	archive := bytes.NewReader(r)
	var err error
	if p.archive, err = zip.NewReader(archive, int64(len(r))); err != nil {
		return err
	}
	// Extract the nuspec file
	if p.nuspecFile, err = p.extractNuspecFile(); err != nil {
		return err
	}
	return nil
}

func (p *PackageArchiveReader) Nuspec() (*Nuspec, error) {
	if p.nuspec != nil {
		return p.nuspec, nil
	}
	var err error
	p.once.Do(func() {
		defer p.nuspecFile.Close()
		// Decode the XML content into the Nuspec struct
		decoder := xml.NewDecoder(p.nuspecFile)
		err = decoder.Decode(&p.nuspec)
	})

	return p.nuspec, err
}

func (p *PackageArchiveReader) extractNuspecFile() (io.ReadCloser, error) {
	for _, file := range p.archive.File {
		if strings.HasSuffix(file.Name, ".nuspec") {
			if nuspecFile, err := file.Open(); err != nil {
				return nil, err
			} else {
				return nuspecFile, nil
			}
		}
	}
	return nil, fmt.Errorf("no .nuspec file found in the .nupkg archive")
}
