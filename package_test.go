// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"github.com/stretchr/testify/require"
	"net/http"
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
