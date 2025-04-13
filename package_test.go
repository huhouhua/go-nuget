// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"strings"
	"testing"
)

// https://api.nuget.org/v3-flatcontainer/newtonsoft.json/index.json
func TestPackageResource_ListAllVersions(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/v3-flatcontainer/newtonsoft.json/index.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		mustWriteHTTPResponse(t, w, "testdata/list_all_versions.json")
	})

	want := []*NuGetVersion{{
		Version: &Version{
			Major:    6,
			Minor:    0,
			Build:    1,
			Revision: 0,
		},
		SemanticVersion: &SemanticVersion{
			releaseLabels: []string{"beta1"},
			metadata:      "",
			Major:         6,
			Minor:         0,
			Patch:         1,
		},
		Revision:        0,
		OriginalVersion: "6.0.1-beta1",
	}, {
		Version: &Version{
			Major:    6,
			Minor:    0,
			Build:    1,
			Revision: 0,
		},
		SemanticVersion: &SemanticVersion{
			releaseLabels: nil,
			metadata:      "",
			Major:         6,
			Minor:         0,
			Patch:         1,
		},
		Revision:        0,
		OriginalVersion: "6.0.1",
	}}

	b, resp, err := client.FindPackage.ListAllVersions("newtonsoft.json", nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, want, b)
}

type Nuspec struct {
	XMLName  xml.Name `xml:"package"`
	Metadata Metadata `xml:"metadata"`
}

type Metadata struct {
	ID          string `xml:"id"`
	Version     string `xml:"version"`
	Authors     string `xml:"authors"`
	Description string `xml:"description"`
	LicenseURL  string `xml:"licenseUrl"`
	ProjectURL  string `xml:"projectUrl"`
}

func TestReaderNupkg(t *testing.T) {
	nupkgPath := "testdata/newtonsoft.json.6.0.1-beta1.nupkg" // 替换成你的 nupkg 路径
	r, err := zip.OpenReader(nupkgPath)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	var nuspecFile io.ReadCloser
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".nuspec") {
			nuspecFile, err = f.Open()
			if err != nil {
				panic(err)
			}
			defer nuspecFile.Close()
			break
		}
	}

	if nuspecFile == nil {
		panic("No .nuspec file found in .nupkg")
	}

	var nuspec Nuspec
	decoder := xml.NewDecoder(nuspecFile)
	err = decoder.Decode(&nuspec)
	if err != nil {
		panic(err)
	}

	fmt.Println("Package ID:", nuspec.Metadata.ID)
	fmt.Println("Version:", nuspec.Metadata.Version)
	fmt.Println("Authors:", nuspec.Metadata.Authors)
	fmt.Println("Description:", nuspec.Metadata.Description)
}
