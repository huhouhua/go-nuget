// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"archive/zip"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReaderNupkg(t *testing.T) {
	nupkgPath := "testdata/test.1.0.0.nupkg"
	r, err := zip.OpenReader(nupkgPath)

	require.NoError(t, err, "open file %s failed %v", nupkgPath, err)

	defer r.Close()

	var nuspecFile io.ReadCloser
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".nuspec") {
			nuspecFile, err = f.Open()
			require.NoError(t, err, "open %s file failed: %v", f.Name, err)
			defer nuspecFile.Close()
			break
		}
	}
	require.NotNil(t, nuspecFile, "No .nuspec file found in .nupkg")

	nuspecBytes, err := io.ReadAll(nuspecFile)
	require.NoError(t, err, "read .nuspec file failed")

	wantNuspecPath := "testdata/myTestLibrary.nuspec"
	wantData, err := os.ReadFile(wantNuspecPath)

	require.NoError(t, err, "read %v .nuspec file failed", wantNuspecPath)
	require.Equal(t, wantData, nuspecBytes)
}
