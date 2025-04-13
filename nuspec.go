// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import "encoding/xml"

type Nuspec struct {
	XMLName  xml.Name `xml:"package"`
	Metadata Metadata `xml:"metadata"`
}

type PackageInfo struct {
	ID                       string     `xml:"id"`
	Version                  string     `xml:"version"`
	Authors                  string     `xml:"authors"`
	Owners                   string     `xml:"owners"`
	RequireLicenseAcceptance bool       `xml:"requireLicenseAcceptance"`
	License                  string     `xml:"license"`
	LicenseURL               string     `xml:"licenseUrl"`
	ProjectURL               string     `xml:"projectUrl"`
	IconUrl                  string     `xml:"iconUrl"`
	Description              string     `xml:"description"`
	Summary                  string     `xml:"summary"`
	ReleaseNotes             string     `xml:"releaseNotes"`
	Copyright                string     `xml:"copyright"`
	Tags                     string     `xml:"tags"`
	Language                 string     `xml:"language"`
	Repository               Repository `xml:"repository"`
}

type Metadata struct {
	PackageInfo
	Dependencies        Dependencies        `xml:"dependencies"`
	FrameworkAssemblies FrameworkAssemblies `xml:"frameworkAssemblies"`
	References          References          `xml:"references"`
}

type Repository struct {
	Type   string `xml:"type,attr"`
	URL    string `xml:"url,attr"`
	Branch string `xml:"branch,attr"`
	Commit string `xml:"commit,attr"`
}

type Dependencies struct {
	Groups     []DependenciesGroup `xml:"group"`
	Dependency []Dependency        `xml:"dependency"`
}

type DependenciesGroup struct {
	TargetFramework string       `xml:"targetFramework,attr"`
	Dependencies    []Dependency `xml:"dependency"`
}

type Dependency struct {
	Id      string   `xml:"id,attr"`
	Version string   `xml:"version,attr"`
	Exclude []string `xml:"exclude,attr"`
	Include string   `xml:"include,attr"`
}

type References struct {
	Groups     []ReferenceGroup `xml:"group"`
	References []Reference      `xml:"reference"`
}

type ReferenceGroup struct {
	TargetFramework string      `xml:"targetFramework,attr"`
	References      []Reference `xml:"reference"`
}

type FrameworkAssemblies struct {
	FrameworkAssembly []FrameworkAssembly `xml:"frameworkAssembly"`
}

type FrameworkAssembly struct {
	AssemblyName    []string `xml:"assemblyName,attr"`
	TargetFramework string   `xml:"targetFramework,attr"`
}

type Reference struct {
	File string `xml:"file,attr"`
}
